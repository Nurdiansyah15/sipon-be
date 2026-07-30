package seeders

import (
	"context"
	"database/sql"
	"log"

	"sipon-be/internal/modules/identity/domain"
)

type RoleSeeder struct{}

func (s *RoleSeeder) Name() string { return "role" }

func (s *RoleSeeder) Run(ctx context.Context, db *sql.DB) error {
	for name, init := range domain.DefaultRolesInit {
		_, err := db.ExecContext(ctx, `
			INSERT INTO roles (id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), NOW())
			ON CONFLICT (name) DO UPDATE SET
				display_name = EXCLUDED.display_name,
				description = EXCLUDED.description,
				role_type = EXCLUDED.role_type,
				scope_type = EXCLUDED.scope_type,
				assignable = EXCLUDED.assignable,
				updated_at = NOW()
		`, string(name), init.DisplayName, init.Description, string(init.RoleType), string(init.ScopeType), init.Assignable)
		if err != nil {
			return err
		}
		log.Printf("[role_seeder] upserted role: %s\n", name)
	}
	return nil
}
