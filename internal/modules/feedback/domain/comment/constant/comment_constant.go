package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeCommentNotFound          kernel.Code = "COMMENT_NOT_FOUND"
	CodeCommentPersistenceFailed kernel.Code = "COMMENT_PERSISTENCE_FAILED"
	CodeCommentQueryFailed       kernel.Code = "COMMENT_QUERY_FAILED"
	CodeCommentEmptyBody         kernel.Code = "COMMENT_EMPTY_BODY"
	CodeCommentInvalidReply      kernel.Code = "COMMENT_INVALID_REPLY"
	CodeCommentAlreadyTakedown   kernel.Code = "COMMENT_ALREADY_TAKEDOWN"
	CodeCommentNotTakedown       kernel.Code = "COMMENT_NOT_TAKEDOWN"
)
