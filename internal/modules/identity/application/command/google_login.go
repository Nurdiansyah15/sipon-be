package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userentity "sipon-be/internal/modules/identity/domain/user/entity"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type GoogleLoginUseCase struct {
	userRepo              userrepo.UserRepository
	tokenGen              ports.TokenGenerator
	googleVerifier        ports.GoogleOAuthVerifier
	allowedGoogleClientID []string
	transactor            ports.Transactor
}

func NewGoogleLoginUseCase(
	userRepo userrepo.UserRepository,
	tokenGen ports.TokenGenerator,
	googleVerifier ports.GoogleOAuthVerifier,
	allowedGoogleClientIDs []string,
	transactor ports.Transactor,
) *GoogleLoginUseCase {
	ids := make([]string, 0, len(allowedGoogleClientIDs))
	for _, id := range allowedGoogleClientIDs {
		v := strings.TrimSpace(id)
		if v == "" {
			continue
		}
		ids = append(ids, v)
	}
	return &GoogleLoginUseCase{
		userRepo:              userRepo,
		tokenGen:              tokenGen,
		googleVerifier:        googleVerifier,
		allowedGoogleClientID: ids,
		transactor:            transactor,
	}
}

func (uc *GoogleLoginUseCase) Execute(ctx context.Context, req dto.GoogleLoginRequest, ipAddress string) (*dto.LoginResponse, error) {
	idToken := req.ResolveIDToken()
	if idToken == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID token wajib diisi", nil)
	}
	if len(uc.allowedGoogleClientID) == 0 {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "allowed google client ids kosong", errors.New("google client ids belum dikonfigurasi"))
	}

	identityInfo, err := uc.googleVerifier.VerifyIDToken(ctx, idToken, uc.allowedGoogleClientID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "Token Google tidak valid", err)
	}

	googleSub := strings.TrimSpace(identityInfo.Subject)
	if googleSub == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "Token Google tidak valid", nil)
	}

	emailVO, err := uservo.NewEmail(identityInfo.Email)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "Email Google tidak valid", err)
	}

	user, err := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierKindGoogle, googleSub)
	if err != nil {
		if !isUserNotFound(err) {
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		user = nil
	}

	if user == nil {
		user, err = uc.resolveOrCreateUserByEmail(ctx, emailVO, googleSub, identityInfo.Name, identityInfo.Picture)
		if err != nil {
			return nil, err
		}
	}

	return issueGoogleTokenPair(ctx, uc, user, req.DeviceID, ipAddress)
}

func (uc *GoogleLoginUseCase) resolveOrCreateUserByEmail(ctx context.Context, email uservo.Email, googleSub, googleName, pictureURL string) (*userentity.User, error) {
	userByEmail, err := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierKindEmail, email.String())
	if err != nil {
		if !isUserNotFound(err) {
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		return uc.createNewGoogleUser(ctx, email, googleSub, googleName, pictureURL)
	}

	emailIdentity := userByEmail.FindLoginIdentity(userconstant.LoginIdentifierKindEmail, email.String())
	if emailIdentity == nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "identitas email kosong pada user", errors.New("email identity not found"))
	}
	if !emailIdentity.IsVerified() {
		return nil, kernel.WrapMsg(application.ErrCodeForbidden, "Email belum terverifikasi", nil)
	}

	if err := uc.linkGoogleIdentity(ctx, userByEmail, googleSub); err != nil {
		return nil, err
	}
	return userByEmail, nil
}

func (uc *GoogleLoginUseCase) linkGoogleIdentity(ctx context.Context, user *userentity.User, googleSub string) error {
	existing := user.FindLoginIdentity(userconstant.LoginIdentifierKindGoogle, googleSub)
	if existing != nil {
		if !existing.IsVerified() {
			existing.MarkVerified()
		}
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui data pengguna", err)
		}
		return nil
	}

	googleCredential := user.FindCredential(userconstant.CredentialTypeGoogle)
	if googleCredential == nil {
		googleCredential = userentity.NewGoogleCredential(uuid.NewString(), user.ID, true)
		user.AddCredential(googleCredential)
	}

	now := time.Now()
	googleIdentity, err := userentity.NewLoginIdentity(
		uuid.NewString(),
		user.ID,
		googleCredential.ID,
		userconstant.LoginIdentifierKindGoogle,
		googleSub,
		true,
		&now,
	)
	if err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat identitas login Google", err)
	}
	googleCredential.AddLoginIdentity(googleIdentity)

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui data pengguna", err)
	}
	return nil
}

func (uc *GoogleLoginUseCase) createNewGoogleUser(ctx context.Context, email uservo.Email, googleSub, googleName, pictureURL string) (*userentity.User, error) {
	username, err := generateGoogleUsername(ctx, uc.userRepo, googleName, email.String())
	if err != nil {
		return nil, err
	}

	fullName := googleName
	if fullName == "" {
		fullName = email.String()
	}
	var fullnamePtr *string
	if strings.TrimSpace(fullName) != "" {
		fullnamePtr = &fullName
	}

	userID := uuid.NewString()
	user, err := userentity.NewUser(userID, username, fullnamePtr, email, nil)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserIDRequired, userconstant.ErrCodeUserEmailRequired:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if p := strings.TrimSpace(pictureURL); p != "" {
		avatarKey := "ext:" + p
		user.AvatarKey = &avatarKey
	}

	now := time.Now()

	localCredential := userentity.NewLocalCredentialWithoutPassword(uuid.NewString(), user.ID, true)
	emailIdentity, err := userentity.NewLoginIdentity(
		uuid.NewString(),
		user.ID,
		localCredential.ID,
		userconstant.LoginIdentifierKindEmail,
		email.String(),
		true,
		&now,
	)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	localCredential.AddLoginIdentity(emailIdentity)

	usernameIdentity, err := userentity.NewLoginIdentity(
		uuid.NewString(),
		user.ID,
		localCredential.ID,
		userconstant.LoginIdentifierKindUsername,
		username.String(),
		true,
		&now,
	)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	localCredential.AddLoginIdentity(usernameIdentity)
	user.AddCredential(localCredential)

	googleCredential := userentity.NewGoogleCredential(uuid.NewString(), user.ID, false)
	googleIdentity, err := userentity.NewLoginIdentity(
		uuid.NewString(),
		user.ID,
		googleCredential.ID,
		userconstant.LoginIdentifierKindGoogle,
		googleSub,
		true,
		&now,
	)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	googleCredential.AddLoginIdentity(googleIdentity)
	user.AddCredential(googleCredential)

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.userRepo.Save(txCtx, user)
	}); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal menyimpan data pengguna", err)
	}

	return user, nil
}

func issueGoogleTokenPair(ctx context.Context, uc *GoogleLoginUseCase, user *userentity.User, deviceID, _ string) (*dto.LoginResponse, error) {
	if err := user.EnsureCanLogin(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserBanned, userconstant.ErrCodeUserNotActive:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, ke.Message, ke)
			case userconstant.ErrCodeUserLockedOut:
				return nil, kernel.WrapMsg(application.ErrCodeTooManyRequests, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	user.MarkLogin()
	user.ResetFailedAttempts()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui data pengguna", err)
	}

	if deviceID == "" {
		deviceID = uuid.NewString()
	}

	sessionID := uuid.NewString()
	accessToken, err := uc.tokenGen.GenerateAccessToken(user.ID, sessionID, deviceID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat access token", err)
	}
	refreshToken, err := uc.tokenGen.GenerateRefreshToken(user.ID, deviceID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat refresh token", err)
	}

	emailLI := user.FindLoginIdentity(userconstant.LoginIdentifierKindEmail, user.Email.String())
	isEmailVerified := emailLI != nil && emailLI.IsVerified()

	var phoneStr *string
	var isPhoneVerified bool
	if user.PhoneNumber != nil {
		s := user.PhoneNumber.String()
		phoneStr = &s
		if li := user.FindLoginIdentity(userconstant.LoginIdentifierKindPhone, s); li != nil {
			isPhoneVerified = li.IsVerified()
		}
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: dto.UserMe{
			ID:              user.ID,
			Username:        user.Username.String(),
			Email:           user.Email.String(),
			IsEmailVerified: isEmailVerified,
			Fullname:        user.Fullname,
			Phone:           phoneStr,
			IsPhoneVerified: isPhoneVerified,
			Status:          string(user.Status),
			CreatedAt:       user.CreatedAt,
			HasPassword:     user.HasLocalPassword(),
			AvatarURL:       nil,
		},
	}, nil
}

func isUserNotFound(err error) bool {
	var ke *kernel.AppError
	if errors.As(err, &ke) {
		return ke.Code == userconstant.ErrCodeUserNotFound
	}
	return false
}

func generateGoogleUsername(ctx context.Context, userRepo userrepo.UserRepository, googleName, email string) (uservo.Username, error) {
	base := strings.ToLower(strings.TrimSpace(googleName))
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.ReplaceAll(base, ".", "_")
	base = strings.ReplaceAll(base, "-", "_")

	if base == "" {
		base = strings.Split(strings.ToLower(email), "@")[0]
	}
	base = strings.ReplaceAll(base, "+", "_")

	// sanitize ke karakter yang diizinkan username
	var sb strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			sb.WriteRune(r)
		}
	}
	base = sb.String()
	if len(base) > 25 {
		base = base[:25]
	}
	if len(base) < 3 {
		base = "user" + base
	}

	candidate := base
	for i := 1; ; i++ {
		exists, err := userRepo.ExistsByUsername(ctx, candidate)
		if err != nil {
			return uservo.Username{}, kernel.WrapMsg(application.ErrCodeInternal, "gagal memeriksa ketersediaan username", err)
		}
		if !exists {
			break
		}
		candidate = base + "_" + fmt.Sprintf("%04d", i)
	}

	username, err := uservo.NewUsername(candidate)
	if err != nil {
		return uservo.Username{}, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat username", err)
	}
	return username, nil
}
