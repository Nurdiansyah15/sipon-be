package ports

import (
	"context"

	"sipon-be/internal/modules/kesantrian"
)

type SantriBasicInfo struct {
	SantriID string
	UserID   string
	NIS      *string
	Status   string
	Fullname *string
}

type KesantrianReader interface {
	GetSantriByID(ctx context.Context, santriID string) (*SantriBasicInfo, error)
	GetSantriByUserID(ctx context.Context, userID string) (*SantriBasicInfo, error)
	ListActiveSantriWithUserID(ctx context.Context) ([]SantriBasicInfo, error)
}

func FromKesantrian(in *kesantrian.SantriBasicInfo) *SantriBasicInfo {
	if in == nil {
		return nil
	}
	return &SantriBasicInfo{
		SantriID: in.SantriID,
		UserID:   in.UserID,
		NIS:      in.NIS,
		Status:   in.Status,
		Fullname: in.Fullname,
	}
}
