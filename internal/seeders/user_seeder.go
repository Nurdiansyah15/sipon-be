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

type seedUserDef struct {
	Username string
	Password string
	Fullname string
	Email    string
	Role     roleconstant.RoleName
}

func (s *UserSeeder) Run(ctx context.Context, db *sql.DB) error {
	users := []seedUserDef{
		{
			Username: "usergod",
			Password: "Usergod1234",
			Fullname: "Developer / Vendor",
			Email:    "usergod@sipon.dev",
			Role:     roleconstant.UserGodRoleName,
		},
		{
			Username: "santri_a",
			Password: "Sipon123!",
			Fullname: "Santri a",
			Email:    "santri.a@sipon.dev",
			Role:     roleconstant.MemberRoleName,
		},
		{
			Username: "santri_b",
			Password: "Sipon123!",
			Fullname: "Santri b",
			Email:    "santri.b@yopmail.com",
			Role:     roleconstant.MemberRoleName,
		},
	}

	for _, u := range users {
		if err := seedUser(ctx, db, u); err != nil {
			return err
		}
	}
	return nil
}

func seedUser(ctx context.Context, db *sql.DB, u seedUserDef) error {
	var count int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_identities WHERE kind = $1 AND value = $2 AND deleted_at IS NULL`,
		string(userconstant.LoginIdentifierKindUsername), u.Username).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Printf("[user_seeder] %s sudah ada, skip", u.Username)
		return nil
	}

	userID := uuid.New().String()
	credID := uuid.New().String()
	identEmailID := uuid.New().String()
	identUsernameID := uuid.New().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
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
	`, userID, u.Username, u.Fullname, u.Email, string(userconstant.UserStatusActive))
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
	`, identEmailID, userID, credID, string(userconstant.LoginIdentifierKindEmail), u.Email, string(userconstant.LoginIdentityStatusVerified), true)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_identities (id, user_id, credential_id, kind, value, status, is_primary, verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), NOW())
	`, identUsernameID, userID, credID, string(userconstant.LoginIdentifierKindUsername), u.Username, string(userconstant.LoginIdentityStatusVerified), false)
	if err != nil {
		return err
	}

	var roleID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = $1`, string(u.Role)).Scan(&roleID)
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

	log.Printf("[user_seeder] created %s user with %s role", u.Username, u.Role)
	return nil
}
