package ports

import "context"

type GoogleIdentityInfo struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

// GoogleOAuthVerifier memvalidasi Google ID Token dan mengembalikan identitas user.
type GoogleOAuthVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string, allowedClientIDs []string) (*GoogleIdentityInfo, error)
}
