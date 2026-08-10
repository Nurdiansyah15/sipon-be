package dokumen_aset

import "context"

type DokumenAsetDownloadResult struct {
	AccessURL string
	ExpiresIn int
}

type Contract interface {
	GetDownloadURL(ctx context.Context, id string, isAuthenticated bool) (*DokumenAsetDownloadResult, error)
}

var _ Contract = (*Module)(nil)
