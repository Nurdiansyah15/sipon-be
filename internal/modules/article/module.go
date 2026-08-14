package article

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/article/application/command"
	"sipon-be/internal/modules/article/application/query"
	"sipon-be/internal/modules/article/infrastructure/external"
	"sipon-be/internal/modules/article/infrastructure/persistence"
	"sipon-be/internal/modules/article/infrastructure/scraper"
	articleHTTP "sipon-be/internal/modules/article/interfaces/http"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/scheduler/application"
)

type Module struct {
	handler       *articleHTTP.ArticleHandler
	sourceHandler *articleHTTP.SourceHandler
	jwtAuth       gin.HandlerFunc
	principalLoad gin.HandlerFunc
}

func NewModule(
	db *sql.DB,
	cfg *config.Config,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	articleRepo := persistence.NewPostgresArticleRepository(db)
	categoryRepo := persistence.NewPostgresCategoryRepository(db)
	transactor := persistence.NewPostgresTransactor(db)

	fileUploader, _ := external.NewMinioFileUploader(
		cfg.Minio.Endpoint,
		cfg.Minio.PublicEndpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.Bucket,
		cfg.Minio.UseSSL,
		cfg.Minio.InternalUseSSL,
	)

	createArticleUC := command.NewCreateArticleUseCase(articleRepo, fileUploader, transactor)
	updateArticleUC := command.NewUpdateArticleUseCase(articleRepo, fileUploader, transactor)
	deleteArticleUC := command.NewDeleteArticleUseCase(articleRepo)
	publishArticleUC := command.NewPublishArticleUseCase(articleRepo)
	archiveArticleUC := command.NewArchiveArticleUseCase(articleRepo)

	getArticleUC := query.NewGetArticleUseCase(articleRepo, fileUploader)
	listArticlesUC := query.NewListArticlesUseCase(articleRepo, fileUploader)

	createCategoryUC := command.NewCreateCategoryUseCase(categoryRepo)
	updateCategoryUC := command.NewUpdateCategoryUseCase(categoryRepo)
	deleteCategoryUC := command.NewDeleteCategoryUseCase(categoryRepo)
	listCategoriesUC := query.NewListCategoriesUseCase(categoryRepo)

	sourceRepo := persistence.NewPostgresSourceRepository(db)
	selectorRepo := persistence.NewPostgresSourceSelectorRepository(db)
	sourceCategoryRepo := persistence.NewPostgresSourceCategoryRepository(db)
	scraperSourceRepo := persistence.NewPostgresScraperSourceRepo(db)

	scrapePipeline := scraper.NewPipeline(articleRepo, 3)

	listSourcesUC := query.NewListSourcesUseCase(sourceRepo, selectorRepo, sourceCategoryRepo)
	createSourceUC := command.NewCreateSourceUseCase(sourceRepo, selectorRepo)
	updateSourceUC := command.NewUpdateSourceUseCase(sourceRepo, selectorRepo)
	deleteSourceUC := command.NewDeleteSourceUseCase(sourceRepo)
	createSourceCategoryUC := command.NewCreateSourceCategoryUseCase(sourceCategoryRepo)
	updateSourceCategoryUC := command.NewUpdateSourceCategoryUseCase(sourceCategoryRepo)
	deleteSourceCategoryUC := command.NewDeleteSourceCategoryUseCase(sourceCategoryRepo)
	triggerScrapeAllUC := command.NewTriggerScrapeAllUseCase(sourceRepo, sourceCategoryRepo, scraperSourceRepo, scrapePipeline)

	handler := articleHTTP.NewArticleHandler(
		createArticleUC,
		updateArticleUC,
		deleteArticleUC,
		publishArticleUC,
		archiveArticleUC,
		getArticleUC,
		listArticlesUC,
		createCategoryUC,
		updateCategoryUC,
		deleteCategoryUC,
		listCategoriesUC,
		fileUploader,
	)

	sourceHandler := articleHTTP.NewSourceHandler(
		listSourcesUC,
		createSourceUC,
		updateSourceUC,
		deleteSourceUC,
		createSourceCategoryUC,
		updateSourceCategoryUC,
		deleteSourceCategoryUC,
		triggerScrapeAllUC,
	)

	return &Module{
		handler:       handler,
		sourceHandler: sourceHandler,
		jwtAuth:       jwtAuth,
		principalLoad: principalLoad,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	articleHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
	articleHTTP.RegisterSourceRoutes(grp, m.sourceHandler, m.jwtAuth, m.principalLoad)
}

func (m *Module) RegisterSchedulerHandlers(_ *application.Registry) {
}
