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
		Code, Name, Type, NormalBalance, SubType string
		Level, IsPostable, IsSystem              int
	}
	roots := []acct{
		{Code: "1000", Name: "ASET", Type: "asset", NormalBalance: "debit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "1100", Name: "Aset Lancar", Type: "asset", NormalBalance: "debit", Level: 1, IsPostable: 0, IsSystem: 0},
		{Code: "1101", Name: "Kas", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "cash_bank"},
		{Code: "1102", Name: "Bank", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "cash_bank"},
		{Code: "1103", Name: "Piutang Santri", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "receivable"},
		{Code: "1104", Name: "Biaya Dibayar Dimuka", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "prepaid_expense"},
		{Code: "1105", Name: "Persediaan", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "inventory"},
		{Code: "1200", Name: "Aset Tetap", Type: "asset", NormalBalance: "debit", Level: 1, IsPostable: 0, IsSystem: 0},
		{Code: "1201", Name: "Tanah", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "fixed_asset"},
		{Code: "1202", Name: "Bangunan", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "fixed_asset"},
		{Code: "1203", Name: "Peralatan", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "fixed_asset"},
		{Code: "1204", Name: "Kendaraan", Type: "asset", NormalBalance: "debit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "fixed_asset"},
		{Code: "1205", Name: "Akumulasi Penyusutan Aset Tetap", Type: "asset", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "accumulated_depreciation"},
		{Code: "1300", Name: "Aset Tidak Berwujud", Type: "asset", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "intangible_asset"},
		{Code: "1400", Name: "Investasi Jangka Panjang", Type: "asset", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "investment"},
		{Code: "1500", Name: "Aset Lainnya", Type: "asset", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "other_asset"},

		{Code: "2000", Name: "KEWAJIBAN", Type: "liability", NormalBalance: "credit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "2100", Name: "Kewajiban Lancar", Type: "liability", NormalBalance: "credit", Level: 1, IsPostable: 0, IsSystem: 0},
		{Code: "2101", Name: "Utang Usaha", Type: "liability", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "payable"},
		{Code: "2102", Name: "Uang Muka Santri", Type: "liability", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "customer_advance"},
		{Code: "2103", Name: "Biaya Diterima Dimuka", Type: "liability", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "unearned_revenue"},
		{Code: "2104", Name: "Utang Pajak", Type: "liability", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "tax_payable"},
		{Code: "2105", Name: "Beban Masih Harus Dibayar", Type: "liability", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "accrued_liability"},
		{Code: "2200", Name: "Liabilitas Jangka Panjang", Type: "liability", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "long_term_liability"},
		{Code: "2300", Name: "Liabilitas Lainnya", Type: "liability", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "other_liability"},

		{Code: "3000", Name: "EKUITAS", Type: "equity", NormalBalance: "credit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "3100", Name: "Modal", Type: "equity", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "capital"},
		{Code: "3200", Name: "Saldo Laba", Type: "equity", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "retained_earnings"},
		{Code: "3201", Name: "Laba Tahun Berjalan", Type: "equity", NormalBalance: "credit", Level: 2, IsPostable: 1, IsSystem: 0, SubType: "current_year_earnings"},
		{Code: "3300", Name: "Prive/Distribusi", Type: "equity", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "withdrawal"},

		{Code: "4000", Name: "PENDAPATAN", Type: "revenue", NormalBalance: "credit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "4100", Name: "Pendapatan SPP", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "operating_revenue"},
		{Code: "4200", Name: "Pendapatan UKT", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "operating_revenue"},
		{Code: "4300", Name: "Pendapatan Daftar Ulang", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "operating_revenue"},
		{Code: "4400", Name: "Pendapatan Insidental", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "operating_revenue"},
		{Code: "4500", Name: "Pendapatan Donasi", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "non_operating_revenue"},
		{Code: "4600", Name: "Pendapatan Lainnya", Type: "revenue", NormalBalance: "credit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "non_operating_revenue"},

		{Code: "5000", Name: "BEBAN", Type: "expense", NormalBalance: "debit", Level: 0, IsPostable: 0, IsSystem: 1},
		{Code: "5100", Name: "Beban Gaji", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "operating_expense"},
		{Code: "5200", Name: "Beban Operasional", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "operating_expense"},
		{Code: "5300", Name: "Beban Pemeliharaan", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "operating_expense"},
		{Code: "5400", Name: "Beban Penyusutan", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "depreciation_expense"},
		{Code: "5500", Name: "Beban Lainnya", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "non_operating_expense"},
		{Code: "5600", Name: "Beban Pokok Penjualan (HPP)", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "cost_of_goods_sold"},
		{Code: "5700", Name: "Beban Pajak", Type: "expense", NormalBalance: "debit", Level: 1, IsPostable: 1, IsSystem: 0, SubType: "tax_expense"},
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

		var subType interface{}
		if a.SubType != "" {
			subType = a.SubType
		}

		err := db.QueryRowContext(ctx, `
			INSERT INTO accounts (id, code, name, type, sub_type, parent_id, level, is_postable, normal_balance, description, is_active, is_system, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name,
				type = EXCLUDED.type,
				sub_type = EXCLUDED.sub_type,
				level = EXCLUDED.level,
				is_postable = EXCLUDED.is_postable,
				normal_balance = EXCLUDED.normal_balance,
				description = EXCLUDED.description,
				is_active = EXCLUDED.is_active,
				is_system = EXCLUDED.is_system,
				updated_at = NOW()
			RETURNING id
		`, id, a.Code, a.Name, a.Type, subType, nil, a.Level, a.IsPostable == 1, a.NormalBalance, desc, true, a.IsSystem == 1, usergodID).Scan(&id)
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
	case code == "1100" || code == "1200" || code == "1300" || code == "1400" || code == "1500":
		return "1000"
	case code >= "1101" && code <= "1105":
		return "1100"
	case code >= "1201" && code <= "1205":
		return "1200"
	case code == "2100":
		return "2000"
	case code >= "2101" && code <= "2105":
		return "2100"
	case code == "2200" || code == "2300":
		return "2000"
	case code == "3100" || code == "3200" || code == "3300":
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
