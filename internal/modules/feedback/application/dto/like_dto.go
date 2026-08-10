package dto

type ToggleLikeRequest struct {
	TargetType string `json:"target_type" binding:"required"` // feedback | comment
	TargetID   string `json:"target_id" binding:"required"`
}

type ToggleLikeResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}
