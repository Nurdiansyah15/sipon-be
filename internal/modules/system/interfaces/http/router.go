package http

import (
	"github.com/gin-gonic/gin"
)

const (
	permissionManageSystemSettings = "manage_system_settings"
)

// RegisterRoutes memasang endpoint admin untuk master scope. Semua route
// dibatasi JWT + principal + permission manage_system_settings.
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *SystemHandler,
	authMiddleware gin.HandlerFunc,
	principalMiddleware gin.HandlerFunc,
) {
	api := router.Group("/api/v1/web/system/scopes")
	api.Use(authMiddleware)
	api.Use(principalMiddleware)
	api.Use(RequirePermission(permissionManageSystemSettings))
	{
		api.GET("", handler.ListScopes)
		api.GET("/:id", handler.GetScope)
		api.POST("", handler.CreateScope)
		api.PUT("/:id", handler.UpdateScope)
		api.DELETE("/:id", handler.DeleteScope)
	}
}
