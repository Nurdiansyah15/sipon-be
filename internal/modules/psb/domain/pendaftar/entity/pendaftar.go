package entity

import (
	"time"

	"sipon-be/internal/modules/psb/domain/pendaftar/constant"
	"sipon-be/internal/shared/kernel"
)

type Pendaftar struct {
	ID            string
	UserID        string
	PsbSettingID  string
	Gender        string
	Program       *string

	Nickname        *string
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

	Status     constant.PendaftarStatus
	AcceptedBy *string
	AcceptedAt *time.Time
	SantriID   *string
	NIS        *string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewPendaftar(id, userID, psbSettingID, gender string, program *string) (*Pendaftar, error) {
	if id == "" || userID == "" || psbSettingID == "" || gender == "" {
		return nil, kernel.New(constant.CodePendaftarNotFound)
	}
	now := time.Now()
	return &Pendaftar{
		ID:           id,
		UserID:       userID,
		PsbSettingID: psbSettingID,
		Gender:       gender,
		Program:      program,
		Status:       constant.StatusDraft,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (p *Pendaftar) UpsertFormulir(fn func(p *Pendaftar)) error {
	if p.Status != constant.StatusDraft && p.Status != constant.StatusPerluRevisi {
		return kernel.New(constant.CodePendaftarInvalidStatus)
	}
	fn(p)
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Pendaftar) Submit() error {
	if p.Status != constant.StatusDraft && p.Status != constant.StatusPerluRevisi {
		return kernel.New(constant.CodePendaftarInvalidStatus)
	}
	p.Status = constant.StatusDiajukan
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Pendaftar) RequestRevision() error {
	if p.Status != constant.StatusDiajukan {
		return kernel.New(constant.CodePendaftarInvalidStatus)
	}
	p.Status = constant.StatusPerluRevisi
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Pendaftar) Reject() error {
	if p.Status != constant.StatusDiajukan {
		return kernel.New(constant.CodePendaftarInvalidStatus)
	}
	p.Status = constant.StatusDitolak
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Pendaftar) Accept(acceptedBy string) error {
	if p.Status != constant.StatusDiajukan {
		return kernel.New(constant.CodePendaftarInvalidStatus)
	}
	now := time.Now()
	p.Status = constant.StatusDiterima
	p.AcceptedBy = &acceptedBy
	p.AcceptedAt = &now
	p.UpdatedAt = now
	return nil
}

func (p *Pendaftar) MarkNotReregistered() error {
	if p.Status != constant.StatusDiterima {
		return kernel.New(constant.CodePendaftarInvalidStatus)
	}
	p.Status = constant.StatusMengundurkanDiri
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Pendaftar) SubmitDaftarUlang() error {
	if p.Status != constant.StatusDiterima && p.Status != constant.StatusPerluRevisiDaftarUlang {
		return kernel.New(constant.CodePendaftarInvalidStatus)
	}
	p.Status = constant.StatusDaftarUlang
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Pendaftar) RequestRevisionDaftarUlang() error {
	if p.Status != constant.StatusDaftarUlang {
		return kernel.New(constant.CodePendaftarInvalidStatus)
	}
	p.Status = constant.StatusPerluRevisiDaftarUlang
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Pendaftar) GenerateNIS(santriID, nis string) error {
	if p.Status != constant.StatusDaftarUlang {
		return kernel.New(constant.CodePendaftarInvalidStatus)
	}
	p.Status = constant.StatusSelesai
	p.SantriID = &santriID
	p.NIS = &nis
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Pendaftar) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}
