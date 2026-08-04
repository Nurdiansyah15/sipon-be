package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, h *ArticleHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	articles := router.Group("/api/v1/web/articles")
	articles.Use(jwtAuth, principalLoad)
	{
		articles.GET("", h.ListArticles)
		articles.GET("/:id", h.GetArticle)

		articles.POST("/media/presign", h.PresignThumbnail)
		articles.POST("/media/confirm", h.ConfirmThumbnail)

		write := articles.Group("")
		write.Use(middleware.RequirePermission("create_article"))
		{
			write.POST("", h.CreateArticle)
		}

		edit := articles.Group("")
		edit.Use(middleware.RequirePermission("edit_article"))
		{
			edit.PUT("/:id", h.UpdateArticle)
			edit.DELETE("/:id", h.DeleteArticle)
		}

		publish := articles.Group("")
		publish.Use(middleware.RequirePermission("publish_article"))
		{
			publish.POST("/:id/publish", h.PublishArticle)
			publish.POST("/:id/archive", h.ArchiveArticle)
		}
	}

	categories := router.Group("/api/v1/web/article-categories")
	categories.Use(jwtAuth, principalLoad)
	{
		categories.GET("", h.ListCategories)

		catWrite := categories.Group("")
		catWrite.Use(middleware.RequirePermission("manage_article_category"))
		{
			catWrite.POST("", h.CreateCategory)
			catWrite.PUT("/:id", h.UpdateCategory)
			catWrite.DELETE("/:id", h.DeleteCategory)
		}
	}
}

func RegisterSourceRoutes(router *gin.RouterGroup, h *SourceHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	sources := router.Group("/api/v1/web/article-sources")
	sources.Use(jwtAuth, principalLoad)
	{
		sources.GET("", h.ListSources)

		write := sources.Group("")
		write.Use(middleware.RequirePermission("manage_article_sources"))
		{
			write.POST("", h.CreateSource)
			write.PUT("/:source_id", h.UpdateSource)
			write.DELETE("/:source_id", h.DeleteSource)
			write.POST("/:source_id/scrape", h.ScrapeAllNow)
			write.POST("/:source_id/categories", h.CreateSourceCategory)
			write.PUT("/categories/:category_id", h.UpdateSourceCategory)
			write.DELETE("/categories/:category_id", h.DeleteSourceCategory)
		}
	}
}
