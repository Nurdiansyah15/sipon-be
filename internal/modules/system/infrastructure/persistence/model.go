package persistence

import (
	"database/sql"
	"time"
)

type ScopeModel struct {
	ID          string
	ScopeType   string
	Code        string
	Name        string
	Description sql.NullString
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
