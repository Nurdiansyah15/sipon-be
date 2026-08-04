package http

import (
	"time"

	"github.com/google/uuid"
)

func generateThumbnailObjectName() string {
	now := time.Now()
	return "articles/thumbnails/" + now.Format("2006/01/02") + "/" + uuid.NewString()
}
