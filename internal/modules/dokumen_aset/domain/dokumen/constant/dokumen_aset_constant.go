package constant

import "sipon-be/internal/shared/kernel"

type Kategori string

const (
	KategoriFormulir Kategori = "formulir"
	KategoriSurat    Kategori = "surat"
	KategoriPanduan  Kategori = "panduan"
	KategoriBrosur   Kategori = "brosur"
	KategoriLainnya  Kategori = "lainnya"
)

var ValidKategoris = map[Kategori]bool{
	KategoriFormulir: true,
	KategoriSurat:    true,
	KategoriPanduan:  true,
	KategoriBrosur:   true,
	KategoriLainnya:  true,
}

var AllowedContentTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"application/pdf": true,
	"application/msword":                                                       true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.ms-excel":                                                  true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/zip":                                                           true,
}

const (
	CodeDokumenNotFound          kernel.Code = "DOKUMEN_ASET_NOT_FOUND"
	CodeDokumenPersistenceFailed kernel.Code = "DOKUMEN_ASET_PERSISTENCE_FAILED"
	CodeDokumenQueryFailed       kernel.Code = "DOKUMEN_ASET_QUERY_FAILED"
	CodeDokumenInvalidKategori   kernel.Code = "DOKUMEN_ASET_INVALID_KATEGORI"
)
