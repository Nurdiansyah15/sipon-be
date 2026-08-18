package psb

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/identity"
	"sipon-be/internal/modules/kesantrian"
	"sipon-be/internal/modules/psb/application/command"
	ports "sipon-be/internal/modules/psb/application/ports"
	"sipon-be/internal/modules/psb/application/query"
	"sipon-be/internal/modules/psb/infrastructure/external"
	"sipon-be/internal/modules/psb/infrastructure/kesantriangateway"
	"sipon-be/internal/modules/psb/infrastructure/persistence"
	psbHTTP "sipon-be/internal/modules/psb/interfaces/http"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/messaging"
)

type Module struct {
	handler       *psbHTTP.PsbHandler
	fileUploader  ports.FileUploader
	jwtAuth       gin.HandlerFunc
	principalLoad gin.HandlerFunc
}

func NewModule(
	db *sql.DB,
	cfg *config.Config,
	_ identity.Contract,
	kesantrianContract kesantrian.Contract,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	settingRepo := persistence.NewPostgresSettingRepository(db)
	pendaftarRepo := persistence.NewPostgresPendaftarRepository(db)
	dokumenRepo := persistence.NewPostgresDokumenRepository(db)
	reviewRepo := persistence.NewPostgresReviewRepository(db)
	transactor := persistence.NewPostgresTransactor(db)

	kesantrianGW := kesantriangateway.New(kesantrianContract)

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

	settingQuery := query.NewSettingQueryUseCase(settingRepo)
	getPendaftaran := query.NewGetPendaftaranUseCase(pendaftarRepo, settingRepo)
	listPendaftaran := query.NewListPendaftaranUseCase(pendaftarRepo)
	listReviews := query.NewListReviewsUseCase(reviewRepo)
	dokumenList := query.NewDokumenListUseCase(pendaftarRepo, settingRepo, dokumenRepo)
	dokumenAccess := query.NewDokumenAccessUseCase(pendaftarRepo, settingRepo, dokumenRepo, fileUploader)

	upsertFormulir := command.NewUpsertFormulirUseCase(settingRepo, pendaftarRepo, dokumenRepo, fileUploader, kesantrianGW)
	pendaftarAction := command.NewPendaftarActionUseCase(pendaftarRepo, dokumenRepo, fileUploader)
	adminReview := command.NewAdminReviewUseCase(pendaftarRepo, reviewRepo, dokumenRepo)
	generateNIS := command.NewGenerateNISUseCase(pendaftarRepo, settingRepo, dokumenRepo, kesantrianGW)

	dokumenPresign := command.NewDokumenPresignUseCase(fileUploader)
	dokumenDelete := command.NewDokumenDeleteUseCase(pendaftarRepo, dokumenRepo, fileUploader, settingRepo, transactor)
	dokumenVerify := command.NewDokumenVerifyUseCase(dokumenRepo)
	dokumenReject := command.NewDokumenRejectUseCase(dokumenRepo)

	manageSetting := command.NewManageSettingUseCase(settingRepo)
	purgePeriod := command.NewPurgePeriodUseCase(settingRepo, pendaftarRepo, dokumenRepo, reviewRepo)

	handler := psbHTTP.NewPsbHandler(
		settingQuery, getPendaftaran, listPendaftaran, listReviews,
		upsertFormulir, pendaftarAction, adminReview, generateNIS,
		dokumenPresign, dokumenDelete, dokumenVerify, dokumenReject, dokumenList, dokumenAccess,
		manageSetting, purgePeriod,
	)

	return &Module{handler: handler, fileUploader: fileUploader, jwtAuth: jwtAuth, principalLoad: principalLoad}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	psbHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}

func (m *Module) EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error {
	return m.fileUploader.EnsurePendingUploadLifecycle(ctx, expireDays)
}

func (m *Module) RegisterMessageHandlers(_ *messaging.Registry) ([]messaging.Binding, error) {
	return nil, nil
}
