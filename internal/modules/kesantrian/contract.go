package kesantrian

import (
	"context"
	"time"
)

type SantriDokumenInput struct {
	Kind             string
	Key              string
	OriginalFilename *string
	MimeType         *string
	Size             *int64
	VerifiedBy       *string
	VerifiedAt       *time.Time
}

type CreateSantriFromPendaftaranInput struct {
	UserID    string
	Gender    string
	EntryYear string

	// ProgramID adalah referensi ke programs.id di module akademik. Ketika
	// diisi, kesantrian akan memanggil kontrak akademik untuk mencatat
	// pemetaan santri→program. Program (string) tetap sebagai cache display.
	ProgramID *string

	Nickname        *string
	Program         *string
	Hobby           *string
	Purpose         *string
	MotivationEntry *string
	POB             *string
	DOB             *time.Time
	Blood           *string

	Address     *string
	SubDistrict *string
	District    *string
	Province    *string
	PostalCode  *string

	PreviousPondokName    *string
	PreviousPondokAddress *string
	PreviousPondokDiv     *string
	PreviousPondokTime    *string

	NIK   *string
	NoKK  *string
	NISN  *string
	NoKIP *string
	NoKKS *string
	NoPKH *string

	Workplace  *string
	Department *string

	HomeStatus *string

	Father         *string
	FatherPN       *string
	FatherNIK      *string
	FatherJob      *string
	FatherGraduate *string
	FatherIncome   *string

	Mother         *string
	MotherPN       *string
	MotherNIK      *string
	MotherJob      *string
	MotherGraduate *string
	MotherIncome   *string

	GuardianRelationship *string
	Guardian             *string
	GuardianPN           *string
	GuardianNIK          *string
	GuardianJob          *string
	GuardianGraduate     *string
	GuardianIncome       *string

	Dokumen []SantriDokumenInput
}

type CreateSantriFromPendaftaranResult struct {
	SantriID string
	NIS      string
}

type SantriBasicInfo struct {
	SantriID string
	UserID   string
	NIS      *string
	Status   string
}

type Contract interface {
	CreateSantriFromPendaftaran(ctx context.Context, in CreateSantriFromPendaftaranInput) (*CreateSantriFromPendaftaranResult, error)
	ListActiveSantriIDs(ctx context.Context) ([]string, error)
	GetSantriByUserID(ctx context.Context, userID string) (*SantriBasicInfo, error)
	GetSantriByID(ctx context.Context, santriID string) (*SantriBasicInfo, error)
	ListActiveSantriWithUserID(ctx context.Context) ([]SantriBasicInfo, error)
}

var _ Contract = (*Module)(nil)
