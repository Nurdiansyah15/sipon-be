package persistence

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// scanner is satisfied by both *sql.Row and *sql.Rows.
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

func nullInt64(i *int64) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

func strFromNull(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func strPtrFromNull(ns sql.NullString) *string {
	return strFromNull(ns)
}

func timeFromNull(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	v := nt.Time
	return &v
}

func int64FromNull(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	v := ni.Int64
	return &v
}

// isUniqueViolation checks whether err is a Postgres unique_violation
// (SQLSTATE 23505) — used to translate a DB-level constraint failure
// (duplicate NIS, duplicate pending request, etc.) into a domain-level
// "already exists" error code instead of a generic internal error.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
