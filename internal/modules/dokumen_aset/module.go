package dokumen_aset

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/dokumen_aset/application/command"
	ports "sipon-be/internal/modules/dokumen_aset/application/ports"
	"sipon-be/internal/modules/dokumen_aset/application/query"
	"sipon-be/internal/modules/dokumen_aset/infrastructure/external"
	"sipon-be/internal/modules/dokumen_aset/infrastructure/persistence"
	dokumenAsetHTTP "sipon-be/internal/modules/dokumen_aset/interfaces/http"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/scheduler/application"
)

type Module struct {
	handler       *dokumenAsetHTTP.DokumenAsetHandler
	downloadUC    *query.DownloadDokumenAsetUseCase
	fileUploader  ports.FileUploader
	jwtAuth       gin.HandlerFunc
	principalLoad gin.HandlerFunc
}

func NewModule(
	db *sql.DB,
	cfg *config.Config,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	dokumenRepo := persistence.NewPostgresDokumenAsetRepository(db)
	transactor := persistence.NewPostgresTransactor(db)

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

	presignUC := command.NewCreateDokumenAsetPresignUseCase(fileUploader)
	confirmUC := command.NewCreateDokumenAsetConfirmUseCase(dokumenRepo, fileUploader, transactor)
	updateUC := command.NewUpdateDokumenAsetUseCase(dokumenRepo)
	deleteUC := command.NewDeleteDokumenAsetUseCase(dokumenRepo, fileUploader, transactor)
	listUC := query.NewListDokumenAsetUseCase(dokumenRepo)
	getUC := query.NewGetDokumenAsetUseCase(dokumenRepo)
	downloadUC := query.NewDownloadDokumenAsetUseCase(dokumenRepo, fileUploader)

	handler := dokumenAsetHTTP.NewDokumenAsetHandler(
		presignUC, confirmUC, updateUC, deleteUC,
		listUC, getUC, downloadUC,
	)

	return &Module{handler: handler, downloadUC: downloadUC, fileUploader: fileUploader, jwtAuth: jwtAuth, principalLoad: principalLoad}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	dokumenAsetHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}

func (m *Module) EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error {
	return m.fileUploader.EnsurePendingUploadLifecycle(ctx, expireDays)
}

func (m *Module) GetDownloadURL(ctx context.Context, id string, isAuthenticated bool) (*DokumenAsetDownloadResult, error) {
	result, err := m.downloadUC.Execute(ctx, id, isAuthenticated)
	if err != nil {
		return nil, err
	}
	return &DokumenAsetDownloadResult{
		AccessURL: result.AccessURL,
		ExpiresIn: result.ExpiresIn,
	}, nil
}

func (m *Module) RegisterSchedulerHandlers(_ *application.Registry) {
}
