package http

import (
	"github.com/gin-gonic/gin"

	ports "sipon-be/internal/modules/identity/application/ports"
	"sipon-be/internal/modules/identity/infrastructure/cache"
	"sipon-be/internal/modules/identity/infrastructure/principal"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/middleware"
)

type MiddlewareBuilder struct {
	TokenGen         ports.TokenGenerator
	SessionStore     ports.SessionRevocationStore
	PrincipalBuilder *principal.Builder
	PrincipalCache   *cache.RedisPrincipalCache
	RateLimiter      ports.RateLimiter
	RateLimitConfig  config.RateLimitConfig
}

func (b *MiddlewareBuilder) JWTAuth() gin.HandlerFunc {
	return middleware.JWTAuth(b.TokenGen, b.SessionStore)
}

func (b *MiddlewareBuilder) PrincipalLoad() gin.HandlerFunc {
	return middleware.PrincipalLoader(b.PrincipalBuilder, b.PrincipalCache)
}

func (b *MiddlewareBuilder) RequirePermission(permissions ...string) gin.HandlerFunc {
	return middleware.RequirePermission(permissions...)
}

func (b *MiddlewareBuilder) RequireRole(roles ...string) gin.HandlerFunc {
	return middleware.RequireRole(roles...)
}

func (b *MiddlewareBuilder) RequireAdmin() gin.HandlerFunc {
	return middleware.RequireAdmin()
}

func (b *MiddlewareBuilder) RateLimitByIP() gin.HandlerFunc {
	return middleware.RateLimitByIP(b.RateLimiter, b.RateLimitConfig)
}

func (b *MiddlewareBuilder) RateLimitByUser() gin.HandlerFunc {
	return middleware.RateLimitByUser(b.RateLimiter, b.RateLimitConfig)
}

func (b *MiddlewareBuilder) RateLimitByAuth() gin.HandlerFunc {
	return middleware.RateLimitByAuth(b.RateLimiter, b.RateLimitConfig)
}

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *IdentityHandler,
	mb *MiddlewareBuilder,
) {
	auth := router.Group("/api/v1/web/auth")
	auth.Use(mb.RateLimitByAuth())
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/request-otp", handler.RequestIdentityOTP)
		auth.POST("/verify-otp", handler.VerifyIdentityOTP)
		auth.POST("/refresh-token", handler.RefreshToken)
		auth.POST("/password/forgot", handler.ForgotPassword)
		auth.POST("/password/reset", handler.ResetPassword)
	}

	api := router.Group("/api/v1")
	api.Use(mb.JWTAuth())
	{
		api.GET("/auth/session", mb.PrincipalLoad(), handler.GetSession)
		api.POST("/auth/logout", handler.Logout)
	}

	web := router.Group("/api/v1/web")
	web.Use(mb.JWTAuth())
	web.Use(mb.RateLimitByUser())
	{
		authGroup := web.Group("/auth")
		authGroup.Use(mb.PrincipalLoad())
		{
			authGroup.GET("/me", handler.Me)
			authGroup.GET("/profile", handler.Profile)
			authGroup.POST("/change-password", handler.ChangePassword)
			authGroup.POST("/set-password", handler.SetPassword)
			authGroup.POST("/change-email/request", handler.RequestChangeIdentityEmail)
			authGroup.POST("/change-email/confirm", handler.ConfirmChangeIdentityEmail)
			authGroup.POST("/change-phone/request", handler.RequestChangeIdentityPhone)
			authGroup.POST("/change-phone/confirm", handler.ConfirmChangeIdentityPhone)
			authGroup.PUT("/profile", handler.UpdateProfile)
			authGroup.GET("/check-username", handler.CheckUsername)
			authGroup.POST("/change-username", handler.ChangeUsername)
			authGroup.POST("/profile/avatar/presign", handler.AvatarPresign)
			authGroup.POST("/profile/avatar/confirm", handler.AvatarConfirm)
			authGroup.DELETE("/profile/avatar", handler.AvatarDelete)
		}

		users := web.Group("/users")
		users.Use(mb.PrincipalLoad())
		{
			users.GET("", mb.RequirePermission("manage_users"), handler.ListUsers)
			users.GET("/:user_id", mb.RequirePermission("manage_users"), handler.GetUser)
			users.POST("", mb.RequirePermission("manage_users"), handler.CreateUser)
			users.POST("/:user_id/reset-password", mb.RequirePermission("reset_user_password"), handler.ResetUserPassword)
			users.POST("/:user_id/deactivate", mb.RequirePermission("deactivate_user"), handler.DeactivateUser)
			users.POST("/:user_id/reactivate", mb.RequirePermission("deactivate_user"), handler.ReactivateUser)
		}

		rp := web.Group("/role-permission")
		rp.Use(mb.PrincipalLoad())
		{
			readRoleGuard := mb.RequirePermission("manage_roles", "manage_role_permissions", "assign_role")
			userRoleReadGuard := mb.RequirePermission("assign_role", "manage_users")

			rp.GET("/roles", readRoleGuard, handler.ListRoles)
			rp.GET("/roles/:role_id", readRoleGuard, handler.GetRole)
			rp.GET("/permission-keys", readRoleGuard, handler.ListPermissions)
			rp.POST("/roles", mb.RequirePermission("manage_roles"), handler.CreateRole)
			rp.PUT("/roles/:role_id", mb.RequirePermission("manage_roles"), handler.UpdateRole)
			rp.POST("/roles/:role_id/permissions", mb.RequirePermission("manage_role_permissions"), handler.AssignRolePermission)
			rp.DELETE("/roles/:role_id/permissions/:permission_key", mb.RequirePermission("manage_role_permissions"), handler.DeleteRolePermission)
			rp.GET("/user-roles", userRoleReadGuard, handler.ListUserRoles)
			rp.GET("/user-roles/:user_role_id", userRoleReadGuard, handler.GetUserRole)
			rp.POST("/user-roles", mb.RequirePermission("assign_role"), handler.AssignUserRole)
			rp.PUT("/user-roles/:user_role_id", mb.RequirePermission("assign_role"), handler.UpdateUserRole)
			rp.POST("/user-roles/:user_role_id/deactivate", mb.RequirePermission("assign_role"), handler.DeactivateUserRole)
			rp.POST("/user-roles/:user_role_id/reactivate", mb.RequirePermission("assign_role"), handler.ReactivateUserRole)
			rp.DELETE("/user-roles/:user_role_id", mb.RequirePermission("assign_role"), handler.DeleteUserRole)
			rp.GET("/roles/:role_id/scopes", readRoleGuard, handler.ListScopes)
			rp.POST("/roles/:role_id/scopes", mb.RequirePermission("manage_role_permissions"), handler.AssignRoleScope)
			rp.DELETE("/roles/:role_id/scopes/:scope_id", mb.RequirePermission("manage_role_permissions"), handler.DeleteRoleScope)
		}
	}
}

func NewMiddlewareBuilder(
	tokenGen ports.TokenGenerator,
	sessionStore ports.SessionRevocationStore,
	principalBuilder *principal.Builder,
	principalCache *cache.RedisPrincipalCache,
	rateLimiter ports.RateLimiter,
	rateLimitConfig config.RateLimitConfig,
) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		TokenGen:         tokenGen,
		SessionStore:     sessionStore,
		PrincipalBuilder: principalBuilder,
		PrincipalCache:   principalCache,
		RateLimiter:      rateLimiter,
		RateLimitConfig:  rateLimitConfig,
	}
}
