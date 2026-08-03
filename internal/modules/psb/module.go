package psb

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/identity"
	"sipon-be/internal/modules/kesantrian"
	"sipon-be/internal/modules/psb/application/command"
	"sipon-be/internal/modules/psb/application/query"
	"sipon-be/internal/modules/psb/infrastructure/external"
	"sipon-be/internal/modules/psb/infrastructure/kesantriangateway"
	"sipon-be/internal/modules/psb/infrastructure/persistence"
	psbHTTP "sipon-be/internal/modules/psb/interfaces/http"
	"sipon-be/internal/shared/config"
)

type Module struct {
	handler       *psbHTTP.PsbHandler
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
	)

	settingQuery := query.NewSettingQueryUseCase(settingRepo)
	getPendaftaran := query.NewGetPendaftaranUseCase(pendaftarRepo, settingRepo)
	listPendaftaran := query.NewListPendaftaranUseCase(pendaftarRepo)
	listReviews := query.NewListReviewsUseCase(reviewRepo)
	dokumenList := query.NewDokumenListUseCase(pendaftarRepo, settingRepo, dokumenRepo)

	upsertFormulir := command.NewUpsertFormulirUseCase(settingRepo, pendaftarRepo)
	pendaftarAction := command.NewPendaftarActionUseCase(pendaftarRepo)
	adminReview := command.NewAdminReviewUseCase(pendaftarRepo, reviewRepo)
	generateNIS := command.NewGenerateNISUseCase(pendaftarRepo, settingRepo, dokumenRepo, kesantrianGW)

	dokumenPresign := command.NewDokumenPresignUseCase(fileUploader)
	dokumenConfirm := command.NewDokumenConfirmUseCase(pendaftarRepo, settingRepo, dokumenRepo, fileUploader, transactor)
	dokumenDelete := command.NewDokumenDeleteUseCase(pendaftarRepo, dokumenRepo, fileUploader, settingRepo, transactor)
	dokumenVerify := command.NewDokumenVerifyUseCase(dokumenRepo)
	dokumenReject := command.NewDokumenRejectUseCase(dokumenRepo)

	manageSetting := command.NewManageSettingUseCase(settingRepo)
	purgePeriod := command.NewPurgePeriodUseCase(settingRepo, pendaftarRepo, dokumenRepo, reviewRepo)

	handler := psbHTTP.NewPsbHandler(
		settingQuery, getPendaftaran, listPendaftaran, listReviews,
		upsertFormulir, pendaftarAction, adminReview, generateNIS,
		dokumenPresign, dokumenConfirm, dokumenDelete, dokumenVerify, dokumenReject, dokumenList,
		manageSetting, purgePeriod,
	)

	return &Module{handler: handler, jwtAuth: jwtAuth, principalLoad: principalLoad}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	psbHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}
