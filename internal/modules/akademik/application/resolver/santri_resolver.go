package resolver

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/shared/kernel"
)

// santriNotFoundCode is the kesantrian domain not-found code. It is matched
// against errors surfaced through the KesantrianReader port so the generic
// HTTP error can be produced without leaking cross-module constants.
const santriNotFoundCode kernel.Code = "SANTRI_NOT_FOUND"

// periodNotFoundCode is the academic period domain not-found code. It is
// matched against repository errors when locating the active (open) period.
const PeriodNotFoundCode kernel.Code = "ACADEMIC_PERIOD_NOT_FOUND"

// ResolveSantriByUserID resolves the authenticated user_id to a santri record
// via the kesantrian port. It maps a missing profile to a generic not-found
// error.
func ResolveSantriByUserID(ctx context.Context, reader ports.KesantrianReader, userID string) (*ports.SantriBasicInfo, error) {
	if userID == "" {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}
	info, err := reader.GetSantriByUserID(ctx, userID)
	if err != nil {
		if application.IsNotFoundErr(err, santriNotFoundCode) {
			return nil, kernel.New(application.ErrCodeNotFound)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if info == nil {
		return nil, kernel.New(application.ErrCodeNotFound)
	}
	return info, nil
}
