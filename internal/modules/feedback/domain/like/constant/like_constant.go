package constant

import "sipon-be/internal/shared/kernel"

type LikeTargetType string

const (
	TargetFeedback LikeTargetType = "feedback"
	TargetComment  LikeTargetType = "comment"
)

const (
	CodeLikeNotFound          kernel.Code = "LIKE_NOT_FOUND"
	CodeLikePersistenceFailed kernel.Code = "LIKE_PERSISTENCE_FAILED"
	CodeLikeQueryFailed       kernel.Code = "LIKE_QUERY_FAILED"
	CodeLikeInvalidTarget     kernel.Code = "LIKE_INVALID_TARGET"
)
