package identity

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"sipon-be/internal/modules/identity/application/command"
	ports "sipon-be/internal/modules/identity/application/ports"
	"sipon-be/internal/modules/identity/application/query"
	roleservice "sipon-be/internal/modules/identity/domain/role/service"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	"sipon-be/internal/modules/identity/infrastructure/cache"
	"sipon-be/internal/modules/identity/infrastructure/external"
	"sipon-be/internal/modules/identity/infrastructure/external/googleoauth"
	"sipon-be/internal/modules/identity/infrastructure/persistence"
	"sipon-be/internal/modules/identity/infrastructure/principal"
	identityHTTP "sipon-be/internal/modules/identity/interfaces/http"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/scheduler/application"
)

// Module's exported surface is method-only, by design — zero exported
// fields. cmd/api/main.go gets RegisterRoutes + RateLimiter(); other modules
// get Contract (contract.go) and nothing else. A future module's own
// constructor MUST declare its dependency as `identity.Contract`, never
// `*identity.Module` — that's what actually keeps it from reaching
// RegisterRoutes/RateLimiter, which exist only for main.go. See
// docs/architecture/module-boundaries.md.
type Module struct {
	handler           *identityHTTP.IdentityHandler
	middlewareBuilder *identityHTTP.MiddlewareBuilder
	rateLimiter       ports.RateLimiter
	principalBuilder  *principal.Builder
	fileUploader      ports.FileUploader

	// getUserSummaryUC backs Contract.GetUserSummary — see contract.go.
	getUserSummaryUC *query.GetUserSummaryUseCase

	// The 3 fields below back Contract.CreateAccountWithNIS/
	// AddNISLoginIdentity/UpdateFullname — see contract.go.
	createAccountWithNISUC *command.CreateAccountWithNISUseCase
	addNISLoginIdentityUC  *command.AddNISLoginIdentityUseCase
	updateFullnameUC       *command.UpdateFullnameUseCase

	// getUserScopeSetUC backs Contract.GetUserScopeSet — see contract.go.
	getUserScopeSetUC *query.GetUserScopeSetUseCase
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

	roleAssignment := roleservice.NewUserRoleAssignmentService(roleRepo, userRoleRepo)

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
		cfg.Minio.PublicEndpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.Bucket,
		cfg.Minio.PrivateBucket,
		cfg.Minio.UseSSL,
		cfg.Minio.InternalUseSSL,
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

	googleVerifier := googleoauth.NewVerifier()

	googleLoginUC := command.NewGoogleLoginUseCase(
		userRepo, tokenGen, googleVerifier, cfg.Google.ClientIDs, transactor,
	)

	getLinkedAccountsUC := query.NewGetLinkedAccountsUseCase(userRepo)
	linkGoogleUC := command.NewLinkGoogleUseCase(userRepo, googleVerifier, cfg.Google.ClientIDs)
	unlinkGoogleUC := command.NewUnlinkGoogleUseCase(userRepo)

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
	getUserSummaryUC := query.NewGetUserSummaryUseCase(userRepo)

	createAccountWithNISUC := command.NewCreateAccountWithNISUseCase(userRepo, hasher)
	addNISLoginIdentityUC := command.NewAddNISLoginIdentityUseCase(userRepo)
	updateFullnameUC := command.NewUpdateFullnameUseCase(userRepo)

	getUserScopeSetUC := query.NewGetUserScopeSetUseCase(userRoleRepo, roleRepo, roleScopeRepo)

	listRolesUC := query.NewListRolesUseCase(roleListRepo)
	getRoleUC := query.NewGetRoleUseCase(roleRepo, rolePermRepo)
	listPermissionsUC := query.NewListPermissionsUseCase()

	createRoleUC := command.NewCreateRoleUseCase(roleRepo, rolePermRepo)
	updateRoleUC := command.NewUpdateRoleUseCase(roleRepo, rolePermRepo)

	assignUserRoleUC := command.NewAssignUserRoleUseCase(
		roleRepo, userRoleRepo, userRepo, rolePermRepo, principalCache,
	)

	updateUserRoleUC := command.NewUpdateUserRoleUseCase(userRoleRepo, userRepo, roleRepo, rolePermRepo, principalCache)
	deactivateUserRoleUC := command.NewDeactivateUserRoleUseCase(userRoleRepo, userRepo, roleRepo, rolePermRepo, principalCache)
	reactivateUserRoleUC := command.NewReactivateUserRoleUseCase(userRoleRepo, userRepo, roleRepo, rolePermRepo, principalCache)
	deleteUserRoleUC := command.NewDeleteUserRoleUseCase(userRoleRepo, principalCache)

	listUserRolesUC := query.NewListUserRolesUseCase(
		userRoleListRepo, userRepo, roleRepo, rolePermRepo,
	)
	getUserRoleUC := query.NewGetUserRoleUseCase(userRoleRepo, userRepo, roleRepo, rolePermRepo)

	assignRolePermissionUC := command.NewAssignRolePermissionUseCase(roleRepo, rolePermRepo, userRoleRepo, principalCache)
	deleteRolePermissionUC := command.NewDeleteRolePermissionUseCase(roleRepo, rolePermRepo, userRoleRepo, principalCache)

	listScopesUC := query.NewListScopesUseCase(roleScopeRepo)
	assignRoleScopeUC := command.NewAssignRoleScopeUseCase(roleRepo, roleScopeRepo, userRoleRepo, principalCache)
	deleteRoleScopeUC := command.NewDeleteRoleScopeUseCase(roleScopeRepo, roleRepo, userRoleRepo, principalCache)

	handler := identityHTTP.NewIdentityHandler(
		registerUC,
		loginUC,
		googleLoginUC,
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
		getLinkedAccountsUC,
		linkGoogleUC,
		unlinkGoogleUC,
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
		handler:                handler,
		middlewareBuilder:      middlewareBuilder,
		rateLimiter:            rateLimiter,
		principalBuilder:       principalBuilder,
		fileUploader:           fileUploader,
		getUserSummaryUC:       getUserSummaryUC,
		createAccountWithNISUC: createAccountWithNISUC,
		addNISLoginIdentityUC:  addNISLoginIdentityUC,
		updateFullnameUC:       updateFullnameUC,
		getUserScopeSetUC:      getUserScopeSetUC,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	identityHTTP.RegisterRoutes(grp, m.handler, m.middlewareBuilder)
}

// RateLimiter exposes the rate limiter for cmd/api/main.go's global
// middleware wiring — see docs/architecture/module-boundaries.md.
func (m *Module) RateLimiter() ports.RateLimiter {
	return m.rateLimiter
}

// AuthMiddleware and PrincipalMiddleware exist solely for cmd/api/main.go to
// hand ready-made gin.HandlerFunc values to OTHER modules (e.g. kesantrian)
// that need JWT auth / principal loading but must never gain access to
// identity's TokenGenerator/SessionStore/PrincipalBuilder/PrincipalCache
// themselves. Deliberately on *Module, NOT on Contract — exactly like
// RateLimiter() — so a module receiving identity.Contract cannot reach
// these. See docs/architecture/module-boundaries.md.
func (m *Module) AuthMiddleware() gin.HandlerFunc {
	return m.middlewareBuilder.JWTAuth()
}

func (m *Module) PrincipalMiddleware() gin.HandlerFunc {
	return m.middlewareBuilder.PrincipalLoad()
}

func (m *Module) GetUserSummary(ctx context.Context, userID string) (*UserSummary, error) {
	res, err := m.getUserSummaryUC.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &UserSummary{
		UserID:    res.ID,
		Username:  res.Username,
		Email:     res.Email,
		IsActive:  res.Status == string(userconstant.UserStatusActive),
		Fullname:  res.Fullname,
		Phone:     res.Phone,
		AvatarKey: res.AvatarKey,
	}, nil
}

func (m *Module) GetPrincipal(ctx context.Context, userID string) (*Principal, error) {
	p, err := m.principalBuilder.Build(ctx, userID)
	if err != nil {
		return nil, err
	}
	scopes := make([]ScopeInfo, len(p.Scopes))
	for i, s := range p.Scopes {
		scopes[i] = ScopeInfo{ScopeType: s.ScopeType, ScopeID: s.ScopeID}
	}
	return &Principal{UserID: p.UserID, Roles: p.Roles, Permissions: p.Permissions, Scopes: scopes}, nil
}

func (m *Module) CreateAccountWithNIS(ctx context.Context, in CreateAccountInput) (*CreateAccountResult, error) {
	res, err := m.createAccountWithNISUC.Execute(ctx, command.CreateAccountWithNISInput{
		Username: in.Username,
		Email:    in.Email,
		Fullname: in.Fullname,
		NISValue: in.NISValue,
	})
	if err != nil {
		return nil, err
	}
	return &CreateAccountResult{
		UserID:            res.UserID,
		GeneratedPassword: res.GeneratedPassword,
	}, nil
}

func (m *Module) AddNISLoginIdentity(ctx context.Context, userID, nisValue string) error {
	return m.addNISLoginIdentityUC.Execute(ctx, userID, nisValue)
}

func (m *Module) GetUserScopeSet(ctx context.Context, userID string) (*UserScopeSet, error) {
	res, err := m.getUserScopeSetUC.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	values := make([]UserScopeValue, len(res.Values))
	for i, v := range res.Values {
		values[i] = UserScopeValue{ScopeType: v.ScopeType, ScopeValue: v.ScopeValue}
	}
	return &UserScopeSet{
		HasSystemRole: res.HasSystemRole,
		Values:        values,
	}, nil
}

func (m *Module) UpdateFullname(ctx context.Context, userID string, fullname string) error {
	return m.updateFullnameUC.Execute(ctx, userID, fullname)
}

func (m *Module) EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error {
	return m.fileUploader.EnsurePendingUploadLifecycle(ctx, expireDays)
}

func (m *Module) RegisterSchedulerHandlers(_ *application.Registry) {
}
