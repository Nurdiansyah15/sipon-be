package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userentity "sipon-be/internal/modules/identity/domain/user/entity"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

func buildLocalUser(t *testing.T, id, email string, withPassword bool) *fakeUserRepo {
	t.Helper()
	return buildLocalUserWithEmailVerified(t, id, email, withPassword, true)
}

// buildLocalUserWithEmailVerified sama seperti buildLocalUser, tapi
// verifiedEmail=false membuat email identity berstatus UNVERIFIED (tanpa
// verifiedAt). Dipakai untuk menguji guard "email harus verified".
func buildLocalUserWithEmailVerified(t *testing.T, id, email string, withPassword, verifiedEmail bool) *fakeUserRepo {
	t.Helper()
	repo := newFakeUserRepo()
	username, err := uservo.NewUsername("user_" + uuid.NewString()[:8])
	if err != nil {
		t.Fatalf("buat username: %v", err)
	}
	em, err := uservo.NewEmail(email)
	if err != nil {
		t.Fatalf("buat email: %v", err)
	}

	user, err := userentity.NewUser(id, username, nil, em, nil)
	if err != nil {
		t.Fatalf("buat user: %v", err)
	}

	var verifiedAt *time.Time
	if verifiedEmail {
		verifiedAt = nowPtr()
	}

	var cred *userentity.Credential
	if withPassword {
		hashed, err := uservo.NewHashedPassword("$2a$10$abcdefghijklmnopqrstuvwxyz0123456789")
		if err != nil {
			t.Fatalf("buat hashed: %v", err)
		}
		cred = userentity.NewLocalCredential(uuid.NewString(), id, hashed, true)
	} else {
		cred = userentity.NewLocalCredentialWithoutPassword(uuid.NewString(), id, true)
	}

	li, err := userentity.NewLoginIdentity(uuid.NewString(), id, cred.ID, userconstant.LoginIdentifierKindEmail, em.String(), true, verifiedAt)
	if err != nil {
		t.Fatalf("buat identity: %v", err)
	}
	if verifiedEmail {
		li.MarkVerified()
	}
	cred.AddLoginIdentity(li)
	user.AddCredential(cred)

	repo.index(user)
	return repo
}

func TestUnlinkGoogleUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := buildLocalUser(t, "u1", "test@example.com", true)
		user, _ := repo.FindByID(ctx, "u1")
		if err := user.LinkGoogleCredential(uuid.NewString(), "google-sub-1"); err != nil {
			t.Fatalf("link google: %v", err)
		}
		repo.index(user)

		uc := NewUnlinkGoogleUseCase(repo)
		if err := uc.Execute(ctx, "u1"); err != nil {
			t.Fatalf("unlink: %v", err)
		}

		user, err := repo.FindByID(ctx, "u1")
		if err != nil {
			t.Fatalf("find: %v", err)
		}
		cred := user.FindCredential(userconstant.CredentialTypeGoogle)
		if cred == nil || cred.DeletedAt == nil {
			t.Fatal("credential GOOGLE harus ter-soft-delete")
		}
	})

	t.Run("rejected when no local password", func(t *testing.T) {
		repo := buildLocalUser(t, "u1", "test@example.com", false)
		user, _ := repo.FindByID(ctx, "u1")
		if err := user.LinkGoogleCredential(uuid.NewString(), "google-sub-1"); err != nil {
			t.Fatalf("link google: %v", err)
		}
		repo.index(user)

		uc := NewUnlinkGoogleUseCase(repo)
		err := uc.Execute(ctx, "u1")
		var ke *kernel.AppError
		if !errors.As(err, &ke) {
			t.Fatalf("harus *kernel.AppError: %v", err)
		}
		if ke.Code != application.ErrCodeConflict {
			t.Fatalf("code harus %s, dapat %s", application.ErrCodeConflict, ke.Code)
		}
	})
}

func TestLinkGoogleUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := buildLocalUser(t, "u1", "test@example.com", true)
		verifier := &fakeGoogleVerifier{
			info: &ports.GoogleIdentityInfo{Subject: "google-sub-1", Email: "test@example.com"},
		}
		uc := NewLinkGoogleUseCase(repo, verifier, []string{"client-1"})

		if err := uc.Execute(ctx, "u1", dto.LinkGoogleRequest{IDToken: "tok"}); err != nil {
			t.Fatalf("link: %v", err)
		}

		user, err := repo.FindByID(ctx, "u1")
		if err != nil {
			t.Fatalf("find: %v", err)
		}
		cred := user.FindCredential(userconstant.CredentialTypeGoogle)
		if cred == nil {
			t.Fatal("credential GOOGLE tidak ada")
		}
		li := user.FindLoginIdentity(userconstant.LoginIdentifierKindGoogle, "google-sub-1")
		if li == nil {
			t.Fatal("identity GOOGLE tidak ada")
		}
	})

	t.Run("rejected when google sub used by other user", func(t *testing.T) {
		repo := buildLocalUser(t, "u1", "test@example.com", true)
		repo2 := buildLocalUser(t, "u2", "other@example.com", true)
		user2, _ := repo2.FindByID(ctx, "u2")
		if err := user2.LinkGoogleCredential(uuid.NewString(), "google-sub-1"); err != nil {
			t.Fatalf("link u2: %v", err)
		}
		repo2.index(user2)
		// merge kedua repo
		for uid, u := range repo2.users {
			repo.index(u)
			_ = uid
		}
		for k, v := range repo2.byIdentity {
			repo.byIdentity[k] = v
		}

		verifier := &fakeGoogleVerifier{
			info: &ports.GoogleIdentityInfo{Subject: "google-sub-1", Email: "test@example.com"},
		}
		uc := NewLinkGoogleUseCase(repo, verifier, []string{"client-1"})

		err := uc.Execute(ctx, "u1", dto.LinkGoogleRequest{IDToken: "tok"})
		var ke *kernel.AppError
		if !errors.As(err, &ke) {
			t.Fatalf("harus *kernel.AppError: %v", err)
		}
		if ke.Code != application.ErrCodeConflict {
			t.Fatalf("code harus conflict, dapat %s", ke.Code)
		}
	})

	t.Run("rejected when email unverified", func(t *testing.T) {
		repo := buildLocalUserWithEmailVerified(t, "u1", "test@example.com", true, false)
		verifier := &fakeGoogleVerifier{
			info: &ports.GoogleIdentityInfo{Subject: "google-sub-1", Email: "test@example.com"},
		}
		uc := NewLinkGoogleUseCase(repo, verifier, []string{"client-1"})

		err := uc.Execute(ctx, "u1", dto.LinkGoogleRequest{IDToken: "tok"})
		var ke *kernel.AppError
		if !errors.As(err, &ke) {
			t.Fatalf("harus *kernel.AppError: %v", err)
		}
		if ke.Code != application.ErrCodeForbidden {
			t.Fatalf("code harus %s, dapat %s", application.ErrCodeForbidden, ke.Code)
		}

		// Tidak boleh ada credential GOOGLE yang tertaut.
		user, _ := repo.FindByID(ctx, "u1")
		if cred := user.FindCredential(userconstant.CredentialTypeGoogle); cred != nil {
			t.Fatal("credential GOOGLE tidak boleh ada saat email unverified")
		}
	})

	t.Run("rejected when already linked", func(t *testing.T) {
		repo := buildLocalUser(t, "u1", "test@example.com", true)
		user, _ := repo.FindByID(ctx, "u1")
		if err := user.LinkGoogleCredential(uuid.NewString(), "google-sub-1"); err != nil {
			t.Fatalf("link awal: %v", err)
		}
		repo.index(user)

		verifier := &fakeGoogleVerifier{
			info: &ports.GoogleIdentityInfo{Subject: "google-sub-2", Email: "test@example.com"},
		}
		uc := NewLinkGoogleUseCase(repo, verifier, []string{"client-1"})

		err := uc.Execute(ctx, "u1", dto.LinkGoogleRequest{IDToken: "tok"})
		var ke *kernel.AppError
		if !errors.As(err, &ke) {
			t.Fatalf("harus *kernel.AppError: %v", err)
		}
		if ke.Code != application.ErrCodeConflict {
			t.Fatalf("code harus %s, dapat %s", application.ErrCodeConflict, ke.Code)
		}
	})
}
