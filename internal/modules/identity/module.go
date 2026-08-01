package identity

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/command"
	"sipon-be/internal/modules/identity/application/query"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/modules/identity/infrastructure/cache"
	"sipon-be/internal/modules/identity/infrastructure/external"
	"sipon-be/internal/modules/identity/infrastructure/persistence"
	"sipon-be/internal/modules/identity/infrastructure/principal"
	identityHTTP "sipon-be/internal/modules/identity/interfaces/http"
	"sipon-be/internal/shared/config"
)

type Module struct {
	Handler           *identityHTTP.IdentityHandler
	MiddlewareBuilder *identityHTTP.MiddlewareBuilder
	Repositories      Repositories
	Services          Services
}

type Repositories struct {
	User           domain.UserRepository
	Verification   domain.VerificationRepository
	Role           domain.RoleRepository
	UserRole       domain.UserRoleRepository
	RolePermission domain.RolePermissionRepository
	RoleScope      domain.RoleScopeRepository
}

type Services struct {
	Hasher                 *external.BcryptPasswordHasher
	TokenGen               *external.JWTTokenGenerator
	EmailSender            application.EmailSender
	SMSSender              application.SMSSender
	OTPGen                 *external.CryptoOTPGenerator
	FileUploader           application.FileUploader
	SessionRevocationStore *cache.RedisSessionRevocationStore
	PrincipalCache         *cache.RedisPrincipalCache
	PrincipalBuilder       *principal.Builder
	RateLimiter            *cache.RedisRateLimiter
}

func NewModule(db *sql.DB, redisClient *redis.Client, cfg *config.Config) *Module {
	userRepo := persistence.NewPostgresUserRepository(db)
	verifRepo := persistence.NewPostgresVerificationRepository(db)
	roleRepo := persistence.NewPostgresRoleRepository(db)
	userRoleRepo := persistence.NewPostgresUserRoleRepository(db)
	rolePermRepo := persistence.NewPostgresRolePermissionRepository(db)
	roleScopeRepo := persistence.NewPostgresRoleScopeRepository(db)
	transactor := persistence.NewPostgresTransactor(db)

	userListRepo := persistence.NewPostgresQueryRepo(db)
	roleListRepo := persistence.NewPostgresRoleListRepo(db)
	userRoleListRepo := persistence.NewPostgresUserRoleListRepo(db)

	roleAssignment := domain.NewUserRoleAssignmentService(roleRepo, userRoleRepo)

	hasher := external.NewBcryptPasswordHasher()
	otpGen := external.NewCryptoOTPGenerator()

	tokenGen := external.NewJWTTokenGenerator(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	emailSender := external.NewSMTPEmailSender(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
	)

	smsSender := external.NewFonnteSMSSender(cfg.Fonnte.Token, cfg.Fonnte.URL)

	fileUploader, _ := external.NewMinioFileUploader(
		cfg.Minio.Endpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.Bucket,
		cfg.Minio.PrivateBucket,
		cfg.Minio.UseSSL,
	)

	sessionRevocationStore := cache.NewRedisSessionRevocationStore(redisClient)
	principalCache := cache.NewRedisPrincipalCache(redisClient)
	rateLimiter := cache.NewRedisRateLimiter(redisClient)

	principalBuilder := principal.NewBuilder(userRoleRepo, roleRepo, rolePermRepo, roleScopeRepo)

	registerUC := command.NewRegisterUseCase(
		userRepo, verifRepo, roleRepo, userRoleRepo,
		hasher, otpGen, emailSender, smsSender,
		tokenGen, transactor, roleAssignment,
	)

	loginUC := command.NewLoginUseCase(userRepo, hasher, tokenGen)

	refreshTokenUC := command.NewRefreshTokenUseCase(tokenGen, sessionRevocationStore, userRepo)

	changePasswordUC := command.NewChangePasswordLocalUseCase(userRepo, hasher)
	setPasswordUC := command.NewSetPasswordLocalUseCase(userRepo, hasher)

	requestOTPUC := command.NewRequestIdentityOTPUseCase(
		userRepo, verifRepo, otpGen, emailSender, smsSender,
	)

	verifyOTPUC := command.NewVerifyIdentityOTPUseCase(userRepo, verifRepo)

	meUC := query.NewMeUseCase(userRepo, fileUploader)

	forgotPasswordUC := command.NewForgotPasswordUseCase(
		userRepo, verifRepo, otpGen, emailSender,
	)

	resetPasswordUC := command.NewResetPasswordUseCase(userRepo, verifRepo, hasher)

	requestChangeIdentityUC := command.NewRequestChangeIdentityUseCase(
		userRepo, verifRepo, otpGen, emailSender, smsSender,
	)

	confirmChangeIdentityUC := command.NewConfirmChangeIdentityUseCase(
		userRepo, verifRepo, transactor,
	)

	getSessionUC := query.NewGetSessionUseCase(
		userRepo, userRoleRepo, roleRepo, rolePermRepo, roleScopeRepo,
	)

	logoutUC := command.NewLogoutUseCase(sessionRevocationStore, cfg.JWT.AccessTokenTTL)

	getProfileUC := query.NewGetProfileUseCase(
		userRepo, userRoleRepo, roleRepo, rolePermRepo, roleScopeRepo, fileUploader,
	)

	updateProfileUC := command.NewUpdateProfileUseCase(userRepo)

	checkUsernameUC := query.NewCheckUsernameUseCase(userRepo)
	changeUsernameUC := command.NewChangeUsernameUseCase(userRepo)

	avatarPresignUC := command.NewAvatarPresignUseCase(fileUploader)
	avatarConfirmUC := command.NewAvatarConfirmUseCase(userRepo, transactor, fileUploader)
	avatarDeleteUC := command.NewAvatarDeleteUseCase(userRepo, fileUploader)

	createUserUC := command.NewCreateUserUseCase(userRepo, hasher)

	resetUserPasswordUC := command.NewResetUserPasswordUseCase(userRepo, hasher)
	deactivateUserUC := command.NewDeactivateUserUseCase(userRepo)
	reactivateUserUC := command.NewReactivateUserUseCase(userRepo)

	listUsersUC := query.NewListUsersUseCase(userListRepo, userRoleRepo, roleRepo)
	getUserUC := query.NewGetUserUseCase(userRepo, userRoleRepo, roleRepo)

	listRolesUC := query.NewListRolesUseCase(roleListRepo)
	getRoleUC := query.NewGetRoleUseCase(roleRepo, rolePermRepo)
	listPermissionsUC := query.NewListPermissionsUseCase()

	createRoleUC := command.NewCreateRoleUseCase(roleRepo, rolePermRepo)
	updateRoleUC := command.NewUpdateRoleUseCase(roleRepo, rolePermRepo)

	assignUserRoleUC := command.NewAssignUserRoleUseCase(
		roleRepo, userRoleRepo, userRepo, rolePermRepo,
	)

	updateUserRoleUC := command.NewUpdateUserRoleUseCase(userRoleRepo, userRepo, roleRepo, rolePermRepo)
	deactivateUserRoleUC := command.NewDeactivateUserRoleUseCase(userRoleRepo, userRepo, roleRepo, rolePermRepo)
	reactivateUserRoleUC := command.NewReactivateUserRoleUseCase(userRoleRepo, userRepo, roleRepo, rolePermRepo)
	deleteUserRoleUC := command.NewDeleteUserRoleUseCase(userRoleRepo)

	listUserRolesUC := query.NewListUserRolesUseCase(
		userRoleListRepo, userRepo, roleRepo, rolePermRepo,
	)
	getUserRoleUC := query.NewGetUserRoleUseCase(userRoleRepo, userRepo, roleRepo, rolePermRepo)

	assignRolePermissionUC := command.NewAssignRolePermissionUseCase(roleRepo, rolePermRepo)
	deleteRolePermissionUC := command.NewDeleteRolePermissionUseCase(roleRepo, rolePermRepo)

	listScopesUC := query.NewListScopesUseCase(roleScopeRepo)
	assignRoleScopeUC := command.NewAssignRoleScopeUseCase(roleRepo, roleScopeRepo)
	deleteRoleScopeUC := command.NewDeleteRoleScopeUseCase(roleScopeRepo)

	handler := identityHTTP.NewIdentityHandler(
		registerUC,
		loginUC,
		refreshTokenUC,
		changePasswordUC,
		setPasswordUC,
		requestOTPUC,
		verifyOTPUC,
		meUC,
		forgotPasswordUC,
		resetPasswordUC,
		requestChangeIdentityUC,
		confirmChangeIdentityUC,
		getSessionUC,
		logoutUC,
		getProfileUC,
		updateProfileUC,
		checkUsernameUC,
		changeUsernameUC,
		avatarPresignUC,
		avatarConfirmUC,
		avatarDeleteUC,
		createUserUC,
		resetUserPasswordUC,
		deactivateUserUC,
		reactivateUserUC,
		listUsersUC,
		getUserUC,
		listRolesUC,
		getRoleUC,
		listPermissionsUC,
		createRoleUC,
		updateRoleUC,
		assignUserRoleUC,
		updateUserRoleUC,
		deactivateUserRoleUC,
		reactivateUserRoleUC,
		deleteUserRoleUC,
		listUserRolesUC,
		getUserRoleUC,
		assignRolePermissionUC,
		deleteRolePermissionUC,
		listScopesUC,
		assignRoleScopeUC,
		deleteRoleScopeUC,
		tokenGen,
	)

	middlewareBuilder := identityHTTP.NewMiddlewareBuilder(
		tokenGen,
		sessionRevocationStore,
		principalBuilder,
		principalCache,
		rateLimiter,
		cfg.RateLimit,
	)

	return &Module{
		Handler:           handler,
		MiddlewareBuilder: middlewareBuilder,
		Repositories: Repositories{
			User:           userRepo,
			Verification:   verifRepo,
			Role:           roleRepo,
			UserRole:       userRoleRepo,
			RolePermission: rolePermRepo,
			RoleScope:      roleScopeRepo,
		},
		Services: Services{
			Hasher:                 hasher,
			TokenGen:               tokenGen,
			EmailSender:            emailSender,
			SMSSender:              smsSender,
			OTPGen:                 otpGen,
			FileUploader:           fileUploader,
			SessionRevocationStore: sessionRevocationStore,
			PrincipalCache:         principalCache,
			PrincipalBuilder:       principalBuilder,
			RateLimiter:            rateLimiter,
		},
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.(*gin.RouterGroup)
	identityHTTP.RegisterRoutes(grp, m.Handler, m.MiddlewareBuilder)
}
