package ports

import "context"

type SantriBasicInfo struct {
	SantriID string
	UserID   string
	NIS      *string
	Status   string
}

type KesantrianReader interface {
	ListActiveSantriIDs(ctx context.Context) ([]string, error)
	GetSantriByUserID(ctx context.Context, userID string) (*SantriBasicInfo, error)
	GetSantriByID(ctx context.Context, santriID string) (*SantriBasicInfo, error)
	ListActiveSantriWithUserID(ctx context.Context) ([]SantriBasicInfo, error)
}
