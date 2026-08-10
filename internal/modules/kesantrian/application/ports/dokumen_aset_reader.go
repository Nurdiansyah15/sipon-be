package ports

import "context"

type DokumenAsetDownloadResult struct {
	AccessURL string
	ExpiresIn int
}

type DokumenAsetReader interface {
	GetDownloadURL(ctx context.Context, id string, isAuthenticated bool) (*DokumenAsetDownloadResult, error)
}
