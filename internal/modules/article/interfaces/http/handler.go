package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/command"
	"sipon-be/internal/modules/article/application/dto"
	ports "sipon-be/internal/modules/article/application/ports"
	"sipon-be/internal/modules/article/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
	"time"
)

type ArticleHandler struct {
	createArticleUC  *command.CreateArticleUseCase
	updateArticleUC  *command.UpdateArticleUseCase
	deleteArticleUC  *command.DeleteArticleUseCase
	publishArticleUC *command.PublishArticleUseCase
	archiveArticleUC *command.ArchiveArticleUseCase

	getArticleUC    *query.GetArticleUseCase
	listArticlesUC  *query.ListArticlesUseCase

	createCategoryUC *command.CreateCategoryUseCase
	updateCategoryUC *command.UpdateCategoryUseCase
	deleteCategoryUC *command.DeleteCategoryUseCase
	listCategoriesUC *query.ListCategoriesUseCase

	fileUploader ports.FileUploader
}

func NewArticleHandler(
	createArticleUC *command.CreateArticleUseCase,
	updateArticleUC *command.UpdateArticleUseCase,
	deleteArticleUC *command.DeleteArticleUseCase,
	publishArticleUC *command.PublishArticleUseCase,
	archiveArticleUC *command.ArchiveArticleUseCase,
	getArticleUC *query.GetArticleUseCase,
	listArticlesUC *query.ListArticlesUseCase,
	createCategoryUC *command.CreateCategoryUseCase,
	updateCategoryUC *command.UpdateCategoryUseCase,
	deleteCategoryUC *command.DeleteCategoryUseCase,
	listCategoriesUC *query.ListCategoriesUseCase,
	fileUploader ports.FileUploader,
) *ArticleHandler {
	return &ArticleHandler{
		createArticleUC:  createArticleUC,
		updateArticleUC:  updateArticleUC,
		deleteArticleUC:  deleteArticleUC,
		publishArticleUC: publishArticleUC,
		archiveArticleUC: archiveArticleUC,
		getArticleUC:     getArticleUC,
		listArticlesUC:   listArticlesUC,
		createCategoryUC: createCategoryUC,
		updateCategoryUC: updateCategoryUC,
		deleteCategoryUC: deleteCategoryUC,
		listCategoriesUC: listCategoriesUC,
		fileUploader:     fileUploader,
	}
}

func (h *ArticleHandler) ListArticles(c *gin.Context) {
	var req dto.ListArticlesQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listArticlesUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar artikel berhasil diambil", items, meta)
}

func (h *ArticleHandler) GetArticle(c *gin.Context) {
	articleID := c.Param("id")
	resp, err := h.getArticleUC.Execute(c.Request.Context(), articleID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "artikel berhasil diambil", resp)
}

func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createArticleUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "artikel berhasil dibuat", resp)
}

func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	userID := middleware.GetUserID(c)
	articleID := c.Param("id")
	var req dto.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateArticleUC.Execute(c.Request.Context(), articleID, userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "artikel berhasil diperbarui", resp)
}

func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	articleID := c.Param("id")
	if err := h.deleteArticleUC.Execute(c.Request.Context(), articleID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "artikel berhasil dihapus", nil)
}

func (h *ArticleHandler) PublishArticle(c *gin.Context) {
	articleID := c.Param("id")
	resp, err := h.publishArticleUC.Execute(c.Request.Context(), articleID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "artikel berhasil dipublikasikan", resp)
}

func (h *ArticleHandler) ArchiveArticle(c *gin.Context) {
	articleID := c.Param("id")
	resp, err := h.archiveArticleUC.Execute(c.Request.Context(), articleID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "artikel berhasil diarsipkan", resp)
}

func (h *ArticleHandler) ListCategories(c *gin.Context) {
	items, err := h.listCategoriesUC.Execute(c.Request.Context(), false)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar kategori berhasil diambil", items)
}

func (h *ArticleHandler) CreateCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createCategoryUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "kategori berhasil dibuat", resp)
}

func (h *ArticleHandler) UpdateCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	categoryID := c.Param("id")
	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateCategoryUC.Execute(c.Request.Context(), categoryID, userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "kategori berhasil diperbarui", resp)
}

func (h *ArticleHandler) DeleteCategory(c *gin.Context) {
	categoryID := c.Param("id")
	if err := h.deleteCategoryUC.Execute(c.Request.Context(), categoryID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "kategori berhasil dihapus", nil)
}

func (h *ArticleHandler) PresignThumbnail(c *gin.Context) {
	var req dto.PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	if h.fileUploader == nil {
		httperror.Handle(c, kernel.WrapMsg(application.ErrCodeInternal, "file uploader tidak tersedia", nil))
		return
	}
	presignURL, key, publicURL, err := h.fileUploader.RequestUpload(c.Request.Context(), generateThumbnailObjectName(), req.ContentType, 15*time.Minute)
	if err != nil {
		httperror.Handle(c, kernel.Wrap(application.ErrCodeInternal, err))
		return
	}
	respond.OK(c, "presign url berhasil dibuat", dto.PresignResponse{
		PresignURL: presignURL,
		Key:        key,
		PublicURL:  publicURL,
		ExpiresIn:  900,
	})
}

func (h *ArticleHandler) ConfirmThumbnail(c *gin.Context) {
	var req dto.ConfirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.fileUploader.ConfirmUpload(c.Request.Context(), req.Key); err != nil {
		httperror.Handle(c, kernel.Wrap(application.ErrCodeInternal, err))
		return
	}
	respond.OK(c, "thumbnail berhasil dikonfirmasi", nil)
}
