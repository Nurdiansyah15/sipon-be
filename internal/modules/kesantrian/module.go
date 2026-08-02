package kesantrian

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"sipon-be/internal/modules/identity"
	"sipon-be/internal/modules/kesantrian/application/command"
	"sipon-be/internal/modules/kesantrian/application/query"
	"sipon-be/internal/modules/kesantrian/infrastructure/external"
	"sipon-be/internal/modules/kesantrian/infrastructure/identitygateway"
	"sipon-be/internal/modules/kesantrian/infrastructure/persistence"
	kesantrianHTTP "sipon-be/internal/modules/kesantrian/interfaces/http"
	"sipon-be/internal/shared/config"
)

// Module's exported surface is method-only, zero exported fields — mirrors
// identity.Module. Only RegisterRoutes exists here (no RateLimiter/Contract
// needed yet: kesantrian has no rate limiter of its own and no other module
// calls into it — YAGNI, see docs/architecture/module-boundaries.md).
type Module struct {
	handler       *kesantrianHTTP.SantriHandler
	jwtAuth       gin.HandlerFunc
	principalLoad gin.HandlerFunc
}

// NewModule takes identity as identity.Contract (not *identity.Module) —
// this is the enforcement point that keeps kesantrian from ever reaching
// identity's RegisterRoutes/RateLimiter/domain/application internals. jwtAuth
// and principalLoad are the two ready-made middleware funcs sourced from
// identity.Module.AuthMiddleware()/PrincipalMiddleware() in cmd/api/main.go
// — see docs/architecture/module-boundaries.md and the kesantrian port plan.
func NewModule(
	db *sql.DB,
	redisClient *redis.Client,
	cfg *config.Config,
	identityContract identity.Contract,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	santriRepo := persistence.NewPostgresSantriRepository(db)
	dokumenRepo := persistence.NewPostgresSantriDokumenRepository(db)
	requestRepo := persistence.NewPostgresSantriRequestRepository(db)
	transactor := persistence.NewPostgresTransactor(db)

	provisioner := identitygateway.New(identityContract)

	fileUploader, _ := external.NewMinioFileUploader(
		cfg.Minio.Endpoint,
		cfg.Minio.PublicEndpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.Bucket,
		cfg.Minio.PrivateBucket,
		cfg.Minio.UseSSL,
	)

	getSantriUC := query.NewGetSantriUseCase(santriRepo, provisioner, fileUploader)
	updateSantriUC := command.NewUpdateSantriUseCase(santriRepo, provisioner)
	requestSantriUC := command.NewRequestSantriUseCase(santriRepo, requestRepo, transactor)

	createSantriUC := command.NewCreateSantriUseCase(santriRepo, provisioner, transactor)
	importSantriUC := command.NewImportSantriUseCase(santriRepo, provisioner, transactor)
	listSantriUC := query.NewListSantriUseCase(santriRepo, provisioner)
	listSantriRequestsUC := query.NewListSantriRequestsUseCase(requestRepo, provisioner)
	approveSantriRequestUC := command.NewApproveSantriRequestUseCase(requestRepo, santriRepo, provisioner, transactor)
	rejectSantriRequestUC := command.NewRejectSantriRequestUseCase(requestRepo, transactor)

	dokumenPresignUC := command.NewDokumenPresignUseCase(fileUploader)
	dokumenConfirmUC := command.NewDokumenConfirmUseCase(santriRepo, dokumenRepo, fileUploader, transactor)
	dokumenListUC := query.NewDokumenListUseCase(santriRepo, dokumenRepo)
	adminDokumenListUC := query.NewAdminDokumenListUseCase(santriRepo, dokumenRepo)
	dokumenAccessUC := query.NewDokumenAccessUseCase(santriRepo, dokumenRepo, fileUploader)
	dokumenDeleteUC := command.NewDokumenDeleteUseCase(santriRepo, dokumenRepo, fileUploader, transactor)
	dokumenVerifyUC := command.NewDokumenVerifyUseCase(dokumenRepo, transactor)
	dokumenRejectUC := command.NewDokumenRejectUseCase(dokumenRepo, transactor)

	handler := kesantrianHTTP.NewSantriHandler(
		getSantriUC,
		updateSantriUC,
		requestSantriUC,
		createSantriUC,
		importSantriUC,
		listSantriUC,
		listSantriRequestsUC,
		approveSantriRequestUC,
		rejectSantriRequestUC,
		dokumenPresignUC,
		dokumenConfirmUC,
		dokumenListUC,
		adminDokumenListUC,
		dokumenAccessUC,
		dokumenDeleteUC,
		dokumenVerifyUC,
		dokumenRejectUC,
	)

	return &Module{handler: handler, jwtAuth: jwtAuth, principalLoad: principalLoad}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	kesantrianHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}
