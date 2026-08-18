package persistence

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func jsonBytes(v json.RawMessage) []byte {
	if len(v) == 0 {
		return []byte("{}")
	}
	return []byte(v)
}

func nullUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

func timeFromPtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func strFromPtr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func timeFromNull(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	v := nt.Time
	return &v
}

func strFromNull(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func uuidFromNull(ns sql.NullString) (*uuid.UUID, error) {
	if !ns.Valid {
		return nil, nil
	}
	uid, err := uuid.Parse(ns.String)
	if err != nil {
		return nil, err
	}
	return &uid, nil
}
