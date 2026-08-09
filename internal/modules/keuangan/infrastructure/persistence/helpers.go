package persistence

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"sipon-be/internal/modules/keuangan/domain/account/constant"
)

type scanner interface {
	Scan(dest ...interface{}) error
}

func nullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func nullTimeVal(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func nullFloat64(f *float64) interface{} {
	if f == nil {
		return nil
	}
	return *f
}

func nullBool(b *bool) interface{} {
	if b == nil {
		return nil
	}
	return *b
}

func nullInt32(i *int32) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

func nullSubType(s *constant.AccountSubType) interface{} {
	if s == nil {
		return nil
	}
	return string(*s)
}

func subTypeFromNull(ns sql.NullString) *constant.AccountSubType {
	if !ns.Valid {
		return nil
	}
	v := constant.AccountSubType(ns.String)
	return &v
}

func strFromNull(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func timeFromNull(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	v := nt.Time
	return &v
}

func float64FromNull(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	v := nf.Float64
	return &v
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
