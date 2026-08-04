package article

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/article/application/command"
	"sipon-be/internal/modules/article/application/query"
	"sipon-be/internal/modules/article/infrastructure/external"
	"sipon-be/internal/modules/article/infrastructure/persistence"
	articleHTTP "sipon-be/internal/modules/article/interfaces/http"
	"sipon-be/internal/shared/config"
)

type Module struct {
	handler       *articleHTTP.ArticleHandler
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

	return &Module{
		handler:       handler,
		jwtAuth:       jwtAuth,
		principalLoad: principalLoad,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	articleHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}
