package feedback

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/feedback/application/command"
	ports "sipon-be/internal/modules/feedback/application/ports"
	"sipon-be/internal/modules/feedback/application/query"
	"sipon-be/internal/modules/feedback/infrastructure/external"
	"sipon-be/internal/modules/feedback/infrastructure/identitygateway"
	"sipon-be/internal/modules/feedback/infrastructure/persistence"
	feedbackHTTP "sipon-be/internal/modules/feedback/interfaces/http"
	"sipon-be/internal/modules/identity"
	"sipon-be/internal/shared/config"
)

type Module struct {
	handler       *feedbackHTTP.FeedbackHandler
	fileUploader  ports.FileUploader
	jwtAuth       gin.HandlerFunc
	principalLoad gin.HandlerFunc
}

func NewModule(
	db *sql.DB,
	cfg *config.Config,
	identityContract identity.Contract,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	feedbackRepo := persistence.NewPostgresFeedbackRepository(db)
	commentRepo := persistence.NewPostgresCommentRepository(db)
	likeRepo := persistence.NewPostgresLikeRepository(db)
	attachmentRepo := persistence.NewPostgresAttachmentRepository(db)
	transactor := persistence.NewPostgresTransactor(db)

	identityGW := identitygateway.New(identityContract)

	fileUploader, _ := external.NewMinioFileUploader(
		cfg.Minio.Endpoint,
		cfg.Minio.PublicEndpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.Bucket,
		cfg.Minio.PrivateBucket,
		cfg.Minio.UseSSL,
	)

	listFeedbacks := query.NewListFeedbacksUseCase(feedbackRepo, attachmentRepo, likeRepo, identityGW)
	getFeedback := query.NewGetFeedbackUseCase(feedbackRepo, attachmentRepo, likeRepo, identityGW, fileUploader)
	listComments := query.NewListCommentsUseCase(commentRepo, feedbackRepo, likeRepo, identityGW)
	listAttachments := query.NewListAttachmentsUseCase(attachmentRepo, feedbackRepo, fileUploader)

	createFeedback := command.NewCreateFeedbackUseCase(feedbackRepo, transactor)
	updateFeedback := command.NewUpdateFeedbackUseCase(feedbackRepo)
	deleteFeedback := command.NewDeleteFeedbackUseCase(feedbackRepo, attachmentRepo, fileUploader, transactor)
	moderateFeedback := command.NewModerateFeedbackUseCase(feedbackRepo)
	createComment := command.NewCreateCommentUseCase(commentRepo, feedbackRepo, transactor)
	updateComment := command.NewUpdateCommentUseCase(commentRepo)
	deleteComment := command.NewDeleteCommentUseCase(commentRepo, feedbackRepo)
	moderateComment := command.NewModerateCommentUseCase(commentRepo)
	toggleLike := command.NewToggleLikeUseCase(likeRepo, feedbackRepo, commentRepo, transactor)
	attachment := command.NewAttachmentUseCase(attachmentRepo, feedbackRepo, fileUploader, transactor)

	handler := feedbackHTTP.NewFeedbackHandler(
		listFeedbacks, getFeedback, listComments, listAttachments,
		createFeedback, updateFeedback, deleteFeedback, moderateFeedback,
		createComment, updateComment, deleteComment, moderateComment,
		toggleLike, attachment,
	)

	return &Module{
		handler:       handler,
		fileUploader:  fileUploader,
		jwtAuth:       jwtAuth,
		principalLoad: principalLoad,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	feedbackHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}

func (m *Module) EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error {
	return m.fileUploader.EnsurePendingUploadLifecycle(ctx, expireDays)
}
