package entity

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

func buildTestUser(t *testing.T, withLocalPassword bool) *User {
	t.Helper()

	username, err := uservo.NewUsername("testuser")
	if err != nil {
		t.Fatalf("buat username: %v", err)
	}
	email, err := uservo.NewEmail("test@example.com")
	if err != nil {
		t.Fatalf("buat email: %v", err)
	}

	user, err := NewUser(uuid.NewString(), username, nil, email, nil)
	if err != nil {
		t.Fatalf("buat user: %v", err)
	}

	var local *Credential
	if withLocalPassword {
		hashed, err := uservo.NewHashedPassword("$2a$10$abcdefghijklmnopqrstuvwxyz0123456789")
		if err != nil {
			t.Fatalf("buat hashed password: %v", err)
		}
		local = NewLocalCredential(uuid.NewString(), user.ID, hashed, true)
	} else {
		local = NewLocalCredentialWithoutPassword(uuid.NewString(), user.ID, true)
	}
	user.AddCredential(local)
	return user
}

func assertKernelCode(t *testing.T, err error, want kernel.Code) {
	t.Helper()
	var ke *kernel.AppError
	if !errors.As(err, &ke) {
		t.Fatalf("harus berupa *kernel.AppError, dapat: %v", err)
	}
	if ke.Code != want {
		t.Fatalf("code harus %s, dapat: %s", want, ke.Code)
	}
}

func TestUser_LinkGoogleCredential(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		user := buildTestUser(t, true)
		if err := user.LinkGoogleCredential(uuid.NewString(), "google-sub-123"); err != nil {
			t.Fatalf("link google: %v", err)
		}

		cred := user.FindCredential(userconstant.CredentialTypeGoogle)
		if cred == nil {
			t.Fatal("credential GOOGLE tidak ditemukan")
		}
		if cred.DeletedAt != nil {
			t.Fatal("credential GOOGLE tidak boleh ter-soft-delete")
		}

		li := user.FindLoginIdentity(userconstant.LoginIdentifierKindGoogle, "google-sub-123")
		if li == nil {
			t.Fatal("login identity GOOGLE tidak ditemukan")
		}
		if !li.IsVerified() {
			t.Fatal("login identity GOOGLE harus verified")
		}
	})

	t.Run("rejected when already linked", func(t *testing.T) {
		user := buildTestUser(t, true)
		if err := user.LinkGoogleCredential(uuid.NewString(), "google-sub-123"); err != nil {
			t.Fatalf("link google pertama: %v", err)
		}

		err := user.LinkGoogleCredential(uuid.NewString(), "google-sub-456")
		assertKernelCode(t, err, userconstant.ErrCodeGoogleAlreadyLinked)
	})
}

func TestUser_UnlinkGoogleCredential(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		user := buildTestUser(t, true)
		if err := user.LinkGoogleCredential(uuid.NewString(), "google-sub-123"); err != nil {
			t.Fatalf("link google: %v", err)
		}

		if err := user.UnlinkGoogleCredential(); err != nil {
			t.Fatalf("unlink google: %v", err)
		}

		cred := user.FindCredential(userconstant.CredentialTypeGoogle)
		if cred == nil {
			t.Fatal("credential GOOGLE tidak ditemukan")
		}
		if cred.DeletedAt == nil {
			t.Fatal("credential GOOGLE harus ter-soft-delete")
		}
	})

	t.Run("rejected when no local password", func(t *testing.T) {
		user := buildTestUser(t, false)
		if err := user.LinkGoogleCredential(uuid.NewString(), "google-sub-123"); err != nil {
			t.Fatalf("link google: %v", err)
		}

		err := user.UnlinkGoogleCredential()
		assertKernelCode(t, err, userconstant.ErrCodeGoogleUnlinkRequiresPassword)
	})

	t.Run("rejected when not linked", func(t *testing.T) {
		user := buildTestUser(t, true)

		err := user.UnlinkGoogleCredential()
		assertKernelCode(t, err, userconstant.ErrCodeGoogleNotLinked)
	})

	t.Run("rejected after already unlinked", func(t *testing.T) {
		user := buildTestUser(t, true)
		if err := user.LinkGoogleCredential(uuid.NewString(), "google-sub-123"); err != nil {
			t.Fatalf("link google: %v", err)
		}
		if err := user.UnlinkGoogleCredential(); err != nil {
			t.Fatalf("unlink google: %v", err)
		}

		err := user.UnlinkGoogleCredential()
		assertKernelCode(t, err, userconstant.ErrCodeGoogleNotLinked)
	})
}
