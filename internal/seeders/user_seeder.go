package seeders

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
)

type UserSeeder struct{}

func (s *UserSeeder) Name() string { return "user" }

func (s *UserSeeder) Run(ctx context.Context, db *sql.DB) error {
	var count int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_identities WHERE kind = $1 AND value = $2 AND deleted_at IS NULL`,
		string(userconstant.LoginIdentifierKindUsername), "usergod").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("[user_seeder] usergod sudah ada, skip")
		return nil
	}

	userID := uuid.New().String()
	credID := uuid.New().String()
	identEmailID := uuid.New().String()
	identUsernameID := uuid.New().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Usergod1234"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (id, username, fullname, email, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`, userID, "usergod", "Developer / Vendor", "usergod@sipon.dev", string(userconstant.UserStatusActive))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO credentials (id, user_id, type, secret_hash, is_primary, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`, credID, userID, string(userconstant.CredentialTypeLocal), string(hashedPassword), true)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_identities (id, user_id, credential_id, kind, value, status, is_primary, verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), NOW())
	`, identEmailID, userID, credID, string(userconstant.LoginIdentifierKindEmail), "usergod@sipon.dev", string(userconstant.LoginIdentityStatusVerified), true)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_identities (id, user_id, credential_id, kind, value, status, is_primary, verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), NOW())
	`, identUsernameID, userID, credID, string(userconstant.LoginIdentifierKindUsername), "usergod", string(userconstant.LoginIdentityStatusVerified), false)
	if err != nil {
		return err
	}

	var roleID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = $1`, string(roleconstant.UserGodRoleName)).Scan(&roleID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_roles (id, user_id, role_id, scope_type, assigned_at, assigned_by, is_active)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), $4, true)
	`, userID, roleID, string(roleconstant.ScopeTypeGlobal), "system")
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("[user_seeder] created usergod user with usergod role")
	return nil
}
