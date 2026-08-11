package seeders

import (
	"context"
	"database/sql"
	"log"

	programconstant "sipon-be/internal/modules/akademik/domain/program/constant"
)

type ProgramSeeder struct{}

func (s *ProgramSeeder) Name() string { return "program" }

func (s *ProgramSeeder) Run(ctx context.Context, db *sql.DB) error {
	type prog struct {
		Code string
		Name string
	}
	programs := []prog{
		{Code: programconstant.ProgramCodeTahfidz, Name: "Tahfidz"},
		{Code: programconstant.ProgramCodeKitab, Name: "Kitab"},
	}

	for _, p := range programs {
		_, err := db.ExecContext(ctx, `
			INSERT INTO programs (id, code, name, status, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, 'active', NOW(), NOW())
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name,
				status = EXCLUDED.status,
				updated_at = NOW()
		`, p.Code, p.Name)
		if err != nil {
			return err
		}
		log.Printf("[program_seeder] upserted program: %s - %s\n", p.Code, p.Name)
	}
	return nil
}
