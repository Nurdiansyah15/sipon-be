package entity

import (
	"time"

	"sipon-be/internal/modules/kesantrian/domain/santri/constant"
	"sipon-be/internal/modules/kesantrian/domain/santri/valueobject"
	"sipon-be/internal/shared/kernel"
)

// Santri is 1:1 with identity's User (via UserID) — see
// docs/architecture/module-boundaries.md: kesantrian never stores a FK to
// identity's users table, only the plain UUID string.
type Santri struct {
	ID     string
	UserID string
	NIS    *valueobject.NIS

	Nickname        *string
	Program         *string
	Option          *string // '1' laki-laki / '2' perempuan, derived from NIS.Gender()
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

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Status           constant.SantriStatus
	StatusChangedBy  *string
	StatusChangedAt  *time.Time
	StatusNotes      *string
}

func NewSantri(id, userID string) (*Santri, error) {
	if id == "" || userID == "" {
		return nil, kernel.New(constant.CodeSantriNotFound)
	}
	now := time.Now()
	return &Santri{
		ID:        id,
		UserID:    userID,
		Status:    constant.SantriStatusSantri,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// SetNIS assigns the NIS value object and derives Option (gender) from it.
func (s *Santri) SetNIS(nis valueobject.NIS) {
	s.NIS = &nis
	gender := nis.Gender()
	s.Option = &gender
}

func (s *Santri) Update() {
	s.UpdatedAt = time.Now()
}

func (s *Santri) SoftDelete() {
	now := time.Now()
	s.DeletedAt = &now
	s.UpdatedAt = now
}

func (s *Santri) MarkAlumni(changedBy string) error {
	if s.Status != constant.SantriStatusSantri {
		return kernel.New(constant.CodeSantriInvalidStatus)
	}
	now := time.Now()
	s.Status = constant.SantriStatusAlumni
	s.StatusChangedBy = &changedBy
	s.StatusChangedAt = &now
	s.UpdatedAt = now
	return nil
}

func (s *Santri) MarkDropOut(changedBy string, notes *string) error {
	if s.Status != constant.SantriStatusSantri {
		return kernel.New(constant.CodeSantriInvalidStatus)
	}
	now := time.Now()
	s.Status = constant.SantriStatusDropOut
	s.StatusChangedBy = &changedBy
	s.StatusChangedAt = &now
	s.StatusNotes = notes
	s.UpdatedAt = now
	return nil
}
