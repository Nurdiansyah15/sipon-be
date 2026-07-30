package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type PostgresVerificationRepository struct {
	db *sql.DB
}

func NewPostgresVerificationRepository(db *sql.DB) *PostgresVerificationRepository {
	return &PostgresVerificationRepository{db: db}
}

func (r *PostgresVerificationRepository) Save(ctx context.Context, code *domain.VerificationCode) error {
	query := `INSERT INTO verification_codes (id, user_id, code, purpose, expires_at, used_at, created_at, new_identity_value)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		code.ID, code.UserID, code.Code.String(), string(code.Purpose),
		code.ExpiresAt, nullTime(code.UsedAt), code.CreatedAt,
		strPtr(code.NewIdentityValue),
	)
	if err != nil {
		return fmt.Errorf("save verification code: %w", err)
	}
	return nil
}

func (r *PostgresVerificationRepository) FindLatestByUserAndPurpose(ctx context.Context, userID string, purpose domain.CodePurpose) (*domain.VerificationCode, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, code, purpose, expires_at, used_at, created_at, new_identity_value FROM verification_codes WHERE user_id = $1 AND purpose = $2 ORDER BY created_at DESC LIMIT 1`, userID, string(purpose))

	var m VerificationCodeModel
	if err := row.Scan(&m.ID, &m.UserID, &m.Code, &m.Purpose, &m.ExpiresAt, &m.UsedAt, &m.CreatedAt, &m.NewIdentityValue); err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.New(domain.ErrCodeVerificationCodeNotFound)
		}
		return nil, fmt.Errorf("find latest verification code: %w", err)
	}

	otp, err := domain.NewOTPCode(m.Code)
	if err != nil {
		return nil, err
	}

	var newIdentityValue *string
	if m.NewIdentityValue.Valid {
		newIdentityValue = &m.NewIdentityValue.String
	}

	var usedAt *time.Time
	if m.UsedAt.Valid {
		usedAt = &m.UsedAt.Time
	}

	return &domain.VerificationCode{
		ID:               m.ID,
		UserID:           m.UserID,
		Code:             otp,
		Purpose:          domain.CodePurpose(m.Purpose),
		ExpiresAt:        m.ExpiresAt,
		UsedAt:           usedAt,
		CreatedAt:        m.CreatedAt,
		NewIdentityValue: newIdentityValue,
	}, nil
}

func (r *PostgresVerificationRepository) Update(ctx context.Context, code *domain.VerificationCode) error {
	query := `UPDATE verification_codes SET code = $1, purpose = $2, expires_at = $3, used_at = $4, new_identity_value = $5 WHERE id = $6`

	_, err := r.db.ExecContext(ctx, query,
		code.Code.String(), string(code.Purpose),
		code.ExpiresAt, nullTime(code.UsedAt),
		strPtr(code.NewIdentityValue), code.ID,
	)
	if err != nil {
		return fmt.Errorf("update verification code: %w", err)
	}
	return nil
}
