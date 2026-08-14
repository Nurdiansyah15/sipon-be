package persistence

import (
	"context"
	"database/sql"
	"time"

	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userentity "sipon-be/internal/modules/identity/domain/user/entity"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *userentity.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal memulai transaksi database", err)
	}
	defer tx.Rollback()

	if err := r.saveUser(ctx, tx, user); err != nil {
		return err
	}

	for _, credential := range user.Credentials {
		if err := r.saveCredential(ctx, tx, credential); err != nil {
			return err
		}
		for _, identity := range credential.LoginIdentities {
			if err := r.saveLoginIdentity(ctx, tx, identity); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal menyimpan transaksi", err)
	}
	return nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *userentity.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal memulai transaksi database", err)
	}
	defer tx.Rollback()

	query := `UPDATE users SET
		username = $1, fullname = $2, email = $3, phone = $4,
		avatar_key = $5, status = $6, updated_at = $7, last_login_at = $8,
		failed_login_attempts = $9, locked_until = $10
		WHERE id = $11 AND deleted_at IS NULL`

	_, err = tx.ExecContext(ctx, query,
		user.Username.String(),
		nullString(user.Fullname),
		user.Email.String(),
		nullPhone(user.PhoneNumber),
		nullString(user.AvatarKey),
		string(user.Status),
		user.UpdatedAt,
		nullTime(user.LastLoginAt),
		user.FailedLoginAttempts,
		nullTime(user.LockedUntil),
		user.ID,
	)
	if err != nil {
		return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal memperbarui pengguna", err)
	}

	for _, credential := range user.Credentials {
		credQuery := `INSERT INTO credentials (id, user_id, type, secret_hash, last_changed_at, is_primary, created_at, updated_at, last_login_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, NOW()), $8, $9, $10)
			ON CONFLICT (id) DO UPDATE SET
				secret_hash = EXCLUDED.secret_hash,
				last_changed_at = EXCLUDED.last_changed_at,
				is_primary = EXCLUDED.is_primary,
				updated_at = EXCLUDED.updated_at,
				last_login_at = EXCLUDED.last_login_at,
				deleted_at = EXCLUDED.deleted_at`

		var secretHash interface{}
		if credential.SecretHash != nil {
			h := credential.SecretHash.String()
			secretHash = h
		}

		var createdAt interface{}
		if !credential.UpdatedAt.IsZero() {
			createdAt = credential.UpdatedAt
		}

		_, err = tx.ExecContext(ctx, credQuery,
			credential.ID,
			credential.UserID,
			string(credential.Type),
			secretHash,
			nullTime(credential.LastChangedAt),
			credential.IsPrimary,
			createdAt,
			credential.UpdatedAt,
			nullTime(credential.LastLoginAt),
			nullTime(credential.DeletedAt),
		)
		if err != nil {
			return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal menyimpan kredensial", err)
		}

		for _, identity := range credential.LoginIdentities {
			liQuery := `INSERT INTO user_identities (id, user_id, credential_id, kind, value, status, is_primary, verified_at, created_at, updated_at, deleted_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, NOW()), $10, $11)
				ON CONFLICT (id) DO UPDATE SET
					value = EXCLUDED.value,
					status = EXCLUDED.status,
					is_primary = EXCLUDED.is_primary,
					verified_at = EXCLUDED.verified_at,
					updated_at = EXCLUDED.updated_at,
					deleted_at = EXCLUDED.deleted_at`

			var liCreatedAt interface{}
			if !identity.CreatedAt.IsZero() {
				liCreatedAt = identity.CreatedAt
			}

			_, err = tx.ExecContext(ctx, liQuery,
				identity.ID,
				identity.UserID,
				identity.CredentialID,
				string(identity.Kind),
				identity.Value,
				string(identity.Status),
				identity.IsPrimary,
				nullTime(identity.VerifiedAt),
				liCreatedAt,
				identity.UpdatedAt,
				nullTime(identity.DeletedAt),
			)
			if err != nil {
				return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal menyimpan identitas login", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal menyimpan transaksi", err)
	}
	return nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*userentity.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, username, fullname, email, phone, avatar_key, status, created_at, updated_at, last_login_at, deleted_at, failed_login_attempts, locked_until FROM users WHERE id = $1 AND deleted_at IS NULL`, id)

	var m UserModel
	if err := scanUser(row, &m); err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.WrapMsg(userconstant.ErrCodeUserNotFound, "Pengguna tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal mencari pengguna berdasarkan ID", err)
	}

	user, err := userFromModel(m)
	if err != nil {
		return nil, err
	}

	credentials, err := r.loadCredentials(ctx, id)
	if err != nil {
		return nil, err
	}

	identities, err := r.loadLoginIdentities(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, cred := range credentials {
		for _, li := range identities {
			if li.CredentialID == cred.ID {
				cred.LoginIdentities = append(cred.LoginIdentities, li)
			}
		}
		user.Credentials = append(user.Credentials, cred)
	}

	return user, nil
}

func (r *PostgresUserRepository) FindByLoginIdentifier(ctx context.Context, identifier uservo.LoginIdentifier) (*userentity.User, error) {
	query := `SELECT u.id, u.username, u.fullname, u.email, u.phone, u.avatar_key, u.status, u.created_at, u.updated_at, u.last_login_at, u.deleted_at, u.failed_login_attempts, u.locked_until
		FROM users u
		INNER JOIN user_identities li ON li.user_id = u.id
		WHERE li.kind = $1 AND li.value = $2 AND li.deleted_at IS NULL AND u.deleted_at IS NULL
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, string(identifier.Kind), identifier.Value)

	var m UserModel
	if err := scanUser(row, &m); err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.WrapMsg(userconstant.ErrCodeUserNotFound, "Pengguna tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal mencari pengguna berdasarkan identitas login", err)
	}

	user, err := userFromModel(m)
	if err != nil {
		return nil, err
	}

	credentials, err := r.loadCredentials(ctx, m.ID)
	if err != nil {
		return nil, err
	}

	identities, err := r.loadLoginIdentities(ctx, m.ID)
	if err != nil {
		return nil, err
	}

	for _, cred := range credentials {
		for _, li := range identities {
			if li.CredentialID == cred.ID {
				cred.LoginIdentities = append(cred.LoginIdentities, li)
			}
		}
		user.Credentials = append(user.Credentials, cred)
	}

	return user, nil
}

func (r *PostgresUserRepository) FindByIdentity(ctx context.Context, kind userconstant.LoginIdentifierKind, value string) (*userentity.User, error) {
	return r.FindByLoginIdentifier(ctx, uservo.LoginIdentifier{Kind: kind, Value: value})
}

func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*userentity.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, username, fullname, email, phone, avatar_key, status, created_at, updated_at, last_login_at, deleted_at, failed_login_attempts, locked_until FROM users WHERE username = $1 AND deleted_at IS NULL`, username)

	var m UserModel
	if err := scanUser(row, &m); err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.WrapMsg(userconstant.ErrCodeUserNotFound, "Pengguna tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal mencari pengguna berdasarkan username", err)
	}

	user, err := userFromModel(m)
	if err != nil {
		return nil, err
	}

	credentials, err := r.loadCredentials(ctx, m.ID)
	if err != nil {
		return nil, err
	}

	identities, err := r.loadLoginIdentities(ctx, m.ID)
	if err != nil {
		return nil, err
	}

	for _, cred := range credentials {
		for _, li := range identities {
			if li.CredentialID == cred.ID {
				cred.LoginIdentities = append(cred.LoginIdentities, li)
			}
		}
		user.Credentials = append(user.Credentials, cred)
	}

	return user, nil
}

func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`, username).Scan(&exists)
	if err != nil {
		return false, kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal memeriksa ketersediaan username", err)
	}
	return exists, nil
}

func (r *PostgresUserRepository) ExistsByLoginIdentity(ctx context.Context, kind userconstant.LoginIdentifierKind, value string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_identities WHERE kind = $1 AND value = $2 AND deleted_at IS NULL)`, string(kind), value).Scan(&exists)
	if err != nil {
		return false, kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal memeriksa ketersediaan identitas login", err)
	}
	return exists, nil
}

func (r *PostgresUserRepository) UpdateUsername(ctx context.Context, userID, newUsername string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET username = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`, newUsername, userID)
	if err != nil {
		return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal memperbarui username", err)
	}
	return nil
}

func (r *PostgresUserRepository) saveUser(ctx context.Context, tx *sql.Tx, user *userentity.User) error {
	query := `INSERT INTO users (id, username, fullname, email, phone, avatar_key, status, created_at, updated_at, last_login_at, deleted_at, failed_login_attempts, locked_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := tx.ExecContext(ctx, query,
		user.ID,
		user.Username.String(),
		nullString(user.Fullname),
		user.Email.String(),
		nullPhone(user.PhoneNumber),
		nullString(user.AvatarKey),
		string(user.Status),
		user.CreatedAt,
		user.UpdatedAt,
		nullTime(user.LastLoginAt),
		nullTime(user.DeletedAt),
		user.FailedLoginAttempts,
		nullTime(user.LockedUntil),
	)
	if err != nil {
		return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal menyimpan pengguna", err)
	}
	return nil
}

func (r *PostgresUserRepository) saveCredential(ctx context.Context, tx *sql.Tx, credential *userentity.Credential) error {
	query := `INSERT INTO credentials (id, user_id, type, secret_hash, last_changed_at, is_primary, created_at, updated_at, last_login_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	var secretHash interface{}
	if credential.SecretHash != nil {
		h := credential.SecretHash.String()
		secretHash = h
	}

	_, err := tx.ExecContext(ctx, query,
		credential.ID,
		credential.UserID,
		string(credential.Type),
		secretHash,
		nullTime(credential.LastChangedAt),
		credential.IsPrimary,
		credential.UpdatedAt,
		credential.UpdatedAt,
		nullTime(credential.LastLoginAt),
		nullTime(credential.DeletedAt),
	)
	if err != nil {
		return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal menyimpan kredensial", err)
	}
	return nil
}

func (r *PostgresUserRepository) saveLoginIdentity(ctx context.Context, tx *sql.Tx, identity *userentity.LoginIdentity) error {
	query := `INSERT INTO user_identities (id, user_id, credential_id, kind, value, status, is_primary, verified_at, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := tx.ExecContext(ctx, query,
		identity.ID,
		identity.UserID,
		identity.CredentialID,
		string(identity.Kind),
		identity.Value,
		string(identity.Status),
		identity.IsPrimary,
		nullTime(identity.VerifiedAt),
		identity.CreatedAt,
		identity.UpdatedAt,
		nullTime(identity.DeletedAt),
	)
	if err != nil {
		return kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal menyimpan identitas login", err)
	}
	return nil
}

func (r *PostgresUserRepository) loadCredentials(ctx context.Context, userID string) ([]*userentity.Credential, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, type, secret_hash, last_changed_at, is_primary, created_at, updated_at, last_login_at, deleted_at FROM credentials WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	if err != nil {
		return nil, kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal memuat kredensial", err)
	}
	defer rows.Close()

	var credentials []*userentity.Credential
	for rows.Next() {
		var m CredentialModel
		if err := rows.Scan(&m.ID, &m.UserID, &m.Type, &m.SecretHash, &m.LastChangedAt, &m.IsPrimary, &m.CreatedAt, &m.UpdatedAt, &m.LastLoginAt, &m.DeletedAt); err != nil {
			return nil, kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal membaca data kredensial", err)
		}

		var secretHash *uservo.HashedPassword
		if m.SecretHash.Valid {
			h, err := uservo.NewHashedPassword(m.SecretHash.String)
			if err != nil {
				return nil, err
			}
			secretHash = &h
		}

		c := &userentity.Credential{
			ID:            m.ID,
			UserID:        m.UserID,
			Type:          userconstant.CredentialType(m.Type),
			SecretHash:    secretHash,
			LastChangedAt: timePtr(m.LastChangedAt),
			IsPrimary:     m.IsPrimary,
			UpdatedAt:     m.UpdatedAt,
			LastLoginAt:   timePtr(m.LastLoginAt),
			DeletedAt:     timePtr(m.DeletedAt),
		}
		credentials = append(credentials, c)
	}
	return credentials, rows.Err()
}

func (r *PostgresUserRepository) loadLoginIdentities(ctx context.Context, userID string) ([]*userentity.LoginIdentity, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, credential_id, kind, value, status, is_primary, verified_at, created_at, updated_at, deleted_at FROM user_identities WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	if err != nil {
		return nil, kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal memuat identitas login", err)
	}
	defer rows.Close()

	var identities []*userentity.LoginIdentity
	for rows.Next() {
		var m LoginIdentityModel
		if err := rows.Scan(&m.ID, &m.UserID, &m.CredentialID, &m.Kind, &m.Value, &m.Status, &m.IsPrimary, &m.VerifiedAt, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, kernel.WrapMsg(userconstant.ErrCodeInternal, "gagal membaca data identitas login", err)
		}

		li := &userentity.LoginIdentity{
			ID:           m.ID,
			UserID:       m.UserID,
			CredentialID: m.CredentialID,
			Kind:         userconstant.LoginIdentifierKind(m.Kind),
			Value:        m.Value,
			Status:       userconstant.LoginIdentityStatus(m.Status),
			IsPrimary:    m.IsPrimary,
			VerifiedAt:   timePtr(m.VerifiedAt),
			CreatedAt:    m.CreatedAt,
			UpdatedAt:    m.UpdatedAt,
			DeletedAt:    timePtr(m.DeletedAt),
		}
		identities = append(identities, li)
	}
	return identities, rows.Err()
}

func scanUser(row interface {
	Scan(dest ...interface{}) error
}, m *UserModel) error {
	return row.Scan(
		&m.ID, &m.Username, &m.Fullname, &m.Email, &m.Phone,
		&m.AvatarKey, &m.Status, &m.CreatedAt, &m.UpdatedAt,
		&m.LastLoginAt, &m.DeletedAt, &m.FailedLoginAttempts, &m.LockedUntil,
	)
}

func userFromModel(m UserModel) (*userentity.User, error) {
	username, err := uservo.NewUsername(m.Username)
	if err != nil {
		return nil, err
	}

	email, err := uservo.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}

	var phone *uservo.PhoneNumber
	if m.Phone.Valid {
		pn, err := uservo.NewPhoneNumber(m.Phone.String)
		if err != nil {
			return nil, err
		}
		phone = &pn
	}

	var avatarKey *string
	if m.AvatarKey.Valid {
		avatarKey = &m.AvatarKey.String
	}

	var fullname *string
	if m.Fullname.Valid {
		fullname = &m.Fullname.String
	}

	return &userentity.User{
		ID:                  m.ID,
		Username:            username,
		Fullname:            fullname,
		Email:               email,
		PhoneNumber:         phone,
		AvatarKey:           avatarKey,
		Status:              userconstant.UserStatus(m.Status),
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
		LastLoginAt:         timePtr(m.LastLoginAt),
		DeletedAt:           timePtr(m.DeletedAt),
		FailedLoginAttempts: m.FailedLoginAttempts,
		LockedUntil:         timePtr(m.LockedUntil),
	}, nil
}

func nullString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func nullPhone(p *uservo.PhoneNumber) interface{} {
	if p == nil {
		return nil
	}
	return p.String()
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func timePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
