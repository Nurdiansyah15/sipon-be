package repository

import (
	"context"

	"sipon-be/internal/modules/identity/domain/verification/constant"
	"sipon-be/internal/modules/identity/domain/verification/entity"
)

type VerificationRepository interface {
	Save(ctx context.Context, code *entity.VerificationCode) error
	FindLatestByUserAndPurpose(ctx context.Context, userID string, purpose constant.CodePurpose) (*entity.VerificationCode, error)
	Update(ctx context.Context, code *entity.VerificationCode) error
}
