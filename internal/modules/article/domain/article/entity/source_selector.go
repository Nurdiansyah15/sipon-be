package entity

import "time"

type SourceSelector struct {
	ID              string
	SourceID        string
	ContentSelector *string
	AuthorSelector  *string
	TagsSelector    *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type SourceSelectorParams struct {
	ID              string
	SourceID        string
	ContentSelector *string
	AuthorSelector  *string
	TagsSelector    *string
}

func NewSourceSelector(params SourceSelectorParams) *SourceSelector {
	now := time.Now()
	return &SourceSelector{
		ID:              params.ID,
		SourceID:        params.SourceID,
		ContentSelector: normalizeSourceStr(params.ContentSelector),
		AuthorSelector:  normalizeSourceStr(params.AuthorSelector),
		TagsSelector:    normalizeSourceStr(params.TagsSelector),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
