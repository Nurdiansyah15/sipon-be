package seeders

import (
	"context"
	"database/sql"
	"log"

	scopeconstant "sipon-be/internal/modules/system/domain/scope/constant"
)

type ScopeSeeder struct{}

func (s *ScopeSeeder) Name() string { return "scope" }

func (s *ScopeSeeder) Run(ctx context.Context, db *sql.DB) error {
	defaults := []struct {
		scopeType   string
		code        string
		name        string
		description string
	}{
		{
			scopeType:   string(scopeconstant.ScopeTypeGender),
			code:        scopeconstant.ScopeCodeMale,
			name:        "Laki-laki",
			description: "Scope untuk data laki-laki",
		},
		{
			scopeType:   string(scopeconstant.ScopeTypeGender),
			code:        scopeconstant.ScopeCodeFemale,
			name:        "Perempuan",
			description: "Scope untuk data perempuan",
		},
	}

	for _, d := range defaults {
		_, err := db.ExecContext(ctx, `
			INSERT INTO scopes (scope_type, code, name, description, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, TRUE, NOW(), NOW())
			ON CONFLICT (scope_type, code) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				updated_at = NOW()
		`, d.scopeType, d.code, d.name, d.description)
		if err != nil {
			return err
		}
		log.Printf("[scope_seeder] upserted scope: %s:%s\n", d.scopeType, d.code)
	}
	return nil
}
