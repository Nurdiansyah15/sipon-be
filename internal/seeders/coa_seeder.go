package seeders

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
)

type COASeeder struct{}

func (s *COASeeder) Name() string { return "coa" }

func (s *COASeeder) Run(ctx context.Context, db *sql.DB) error {
	type acct struct {
		Code, Name, Type, NormalBalance string
		Level, IsPostable, IsSystem     int
	}
	roots := []acct{
		{Code: "1000", Name: "ASET", Type: "asset", NormalBalance: "debit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "1100", Name: "Aset Lancar", Type: "asset", NormalBalance: "debit", Level: 1, IsPostable: 0, IsSystem: 0},
		{Code: "1101", Name: "Kas", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0},
		{Code: "1102", Name: "Bank", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0},
		{Code: "1103", Name: "Piutang Santri", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0},
		{Code: "1200", Name: "Aset Tetap", Type: "asset", NormalBalance: "debit", Level: 1, IsPostable: 0, IsSystem: 0},
		{Code: "1201", Name: "Tanah", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0},
		{Code: "1202", Name: "Bangunan", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0},
		{Code: "1203", Name: "Peralatan", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0},
		{Code: "1204", Name: "Kendaraan", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0},

		{Code: "2000", Name: "KEWAJIBAN", Type: "liability", NormalBalance: "credit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "2100", Name: "Kewajiban Lancar", Type: "liability", NormalBalance: "credit", Level: 1, IsPostable: 0, IsSystem: 0},
		{Code: "2101", Name: "Utang Usaha", Type: "liability", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0},
		{Code: "2102", Name: "Uang Muka Santri", Type: "liability", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0},
		{Code: "2103", Name: "Biaya Diterima Dimuka", Type: "liability", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0},

		{Code: "3000", Name: "EKUITAS", Type: "equity", NormalBalance: "credit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "3100", Name: "Modal", Type: "equity", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "3200", Name: "Saldo Laba", Type: "equity", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "3201", Name: "Laba Tahun Berjalan", Type: "equity", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0},

		{Code: "4000", Name: "PENDAPATAN", Type: "revenue", NormalBalance: "credit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "4100", Name: "Pendapatan SPP", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "4200", Name: "Pendapatan UKT", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "4300", Name: "Pendapatan Daftar Ulang", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "4400", Name: "Pendapatan Insidental", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "4500", Name: "Pendapatan Donasi", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "4600", Name: "Pendapatan Lainnya", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0},

		{Code: "5000", Name: "BEBAN", Type: "expense", NormalBalance: "debit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "5100", Name: "Beban Gaji", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "5200", Name: "Beban Operasional", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "5300", Name: "Beban Pemeliharaan", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "5400", Name: "Beban Penyusutan", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0},
		{Code: "5500", Name: "Beban Lainnya", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0},
	}

	rootIDs := make(map[string]string)
	usergodID := "00000000-0000-0000-0000-000000000000"

	for _, a := range roots {
		id := uuid.New().String()

		var desc sql.NullString
		if a.Level == 0 {
			desc = sql.NullString{String: "Root - " + a.Name, Valid: true}
		} else if a.IsPostable == 0 {
			desc = sql.NullString{String: "Group - " + a.Name, Valid: true}
		}

		_, err := db.ExecContext(ctx, `
			INSERT INTO accounts (id, code, name, type, parent_id, level, is_postable, normal_balance, description, is_active, is_system, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name,
				type = EXCLUDED.type,
				parent_id = EXCLUDED.parent_id,
				level = EXCLUDED.level,
				is_postable = EXCLUDED.is_postable,
				normal_balance = EXCLUDED.normal_balance,
				description = EXCLUDED.description,
				is_active = EXCLUDED.is_active,
				is_system = EXCLUDED.is_system,
				updated_at = NOW()
		`, id, a.Code, a.Name, a.Type, nil, a.Level, a.IsPostable == 1, a.NormalBalance, desc, true, a.IsSystem == 1, usergodID)
		if err != nil {
			log.Printf("[coa_seeder] error upserting account %s: %v\n", a.Code, err)
			return err
		}
		rootIDs[a.Code] = id
		log.Printf("[coa_seeder] upserted account: %s - %s\n", a.Code, a.Name)
	}

	for _, a := range roots {
		if a.Level == 0 {
			continue
		}
		parentCode := parentCodeFor(a.Code)
		if parentCode == "" {
			continue
		}
		parentID := rootIDs[parentCode]

		_, err := db.ExecContext(ctx, `
			UPDATE accounts SET parent_id = $1, updated_at = NOW()
			WHERE code = $2
		`, parentID, a.Code)
		if err != nil {
			log.Printf("[coa_seeder] error setting parent for %s: %v\n", a.Code, err)
			return err
		}
	}

	return nil
}

func parentCodeFor(code string) string {
	switch {
	case code == "1100" || code == "1200":
		return "1000"
	case code >= "1101" && code <= "1105":
		return "1100"
	case code >= "1201" && code <= "1204":
		return "1200"
	case code == "2100":
		return "2000"
	case code >= "2101" && code <= "2103":
		return "2100"
	case code == "3100" || code == "3200":
		return "3000"
	case code == "3201":
		return "3200"
	case code >= "4100" && code <= "4999":
		return "4000"
	case code >= "5100" && code <= "5999":
		return "5000"
	default:
		return ""
	}
}
