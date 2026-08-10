package constant

import "sipon-be/internal/shared/kernel"

type FeedbackCategory string

const (
	CategorySaran      FeedbackCategory = "saran"
	CategoryPengaduan  FeedbackCategory = "pengaduan"
	CategoryPertanyaan FeedbackCategory = "pertanyaan"
	CategoryApresiasi  FeedbackCategory = "apresiasi"
	CategoryLainnya    FeedbackCategory = "lainnya"
)

var ValidCategories = map[FeedbackCategory]bool{
	CategorySaran:      true,
	CategoryPengaduan:  true,
	CategoryPertanyaan: true,
	CategoryApresiasi:  true,
	CategoryLainnya:    true,
}

const (
	CodeFeedbackNotFound          kernel.Code = "FEEDBACK_NOT_FOUND"
	CodeFeedbackPersistenceFailed kernel.Code = "FEEDBACK_PERSISTENCE_FAILED"
	CodeFeedbackQueryFailed       kernel.Code = "FEEDBACK_QUERY_FAILED"
	CodeFeedbackInvalidCategory   kernel.Code = "FEEDBACK_INVALID_CATEGORY"
	CodeFeedbackEmptyTitle        kernel.Code = "FEEDBACK_EMPTY_TITLE"
	CodeFeedbackEmptyBody         kernel.Code = "FEEDBACK_EMPTY_BODY"
	CodeFeedbackAlreadyTakedown   kernel.Code = "FEEDBACK_ALREADY_TAKEDOWN"
	CodeFeedbackNotTakedown       kernel.Code = "FEEDBACK_NOT_TAKEDOWN"
	CodeFeedbackDeleted           kernel.Code = "FEEDBACK_DELETED"
)
