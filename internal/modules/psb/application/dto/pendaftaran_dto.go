package dto

import "time"

type UpsertFormulirRequest struct {
	Nickname        *string    `json:"nickname,omitempty"`
	Program         *string    `json:"program,omitempty"`
	Hobby           *string    `json:"hobby,omitempty"`
	Purpose         *string    `json:"purpose,omitempty"`
	MotivationEntry *string    `json:"motivation_entry,omitempty"`
	POB             *string    `json:"pob,omitempty"`
	DOB             *time.Time `json:"dob,omitempty"`
	Blood           *string    `json:"blood,omitempty"`

	Address     *string `json:"address,omitempty"`
	SubDistrict *string `json:"sub_district,omitempty"`
	District    *string `json:"district,omitempty"`
	Province    *string `json:"province,omitempty"`
	PostalCode  *string `json:"postal_code,omitempty"`

	PreviousPondokName    *string `json:"previous_pondok_name,omitempty"`
	PreviousPondokAddress *string `json:"previous_pondok_address,omitempty"`
	PreviousPondokDiv     *string `json:"previous_pondok_div,omitempty"`
	PreviousPondokTime    *string `json:"previous_pondok_time,omitempty"`

	NIK   *string `json:"nik,omitempty"`
	NoKK  *string `json:"no_kk,omitempty"`
	NISN  *string `json:"nisn,omitempty"`
	NoKIP *string `json:"no_kip,omitempty"`
	NoKKS *string `json:"no_kks,omitempty"`
	NoPKH *string `json:"no_pkh,omitempty"`

	Workplace  *string `json:"workplace,omitempty"`
	Department *string `json:"department,omitempty"`

	HomeStatus *string `json:"home_status,omitempty"`

	Father         *string `json:"father,omitempty"`
	FatherPN       *string `json:"father_pn,omitempty"`
	FatherNIK      *string `json:"father_nik,omitempty"`
	FatherJob      *string `json:"father_job,omitempty"`
	FatherGraduate *string `json:"father_graduate,omitempty"`
	FatherIncome   *string `json:"father_income,omitempty"`

	Mother         *string `json:"mother,omitempty"`
	MotherPN       *string `json:"mother_pn,omitempty"`
	MotherNIK      *string `json:"mother_nik,omitempty"`
	MotherJob      *string `json:"mother_job,omitempty"`
	MotherGraduate *string `json:"mother_graduate,omitempty"`
	MotherIncome   *string `json:"mother_income,omitempty"`

	GuardianRelationship *string `json:"guardian_relationship,omitempty"`
	Guardian             *string `json:"guardian,omitempty"`
	GuardianPN           *string `json:"guardian_pn,omitempty"`
	GuardianNIK          *string `json:"guardian_nik,omitempty"`
	GuardianJob          *string `json:"guardian_job,omitempty"`
	GuardianGraduate     *string `json:"guardian_graduate,omitempty"`
	GuardianIncome       *string `json:"guardian_income,omitempty"`

	Dokumen []FormulirDokumenItem `json:"dokumen,omitempty"`
}

type FormulirDokumenItem struct {
	Stage string `json:"stage" binding:"required"`
	Kind  string `json:"kind" binding:"required"`
	Key   string `json:"key" binding:"required"`
}

type PendaftarResponse struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	PsbSettingID  string     `json:"psb_setting_id"`
	Gender        string     `json:"gender"`
	Program       *string    `json:"program,omitempty"`

	Nickname        *string    `json:"nickname,omitempty"`
	Hobby           *string    `json:"hobby,omitempty"`
	Purpose         *string    `json:"purpose,omitempty"`
	MotivationEntry *string    `json:"motivation_entry,omitempty"`
	POB             *string    `json:"pob,omitempty"`
	DOB             *time.Time `json:"dob,omitempty"`
	Blood           *string    `json:"blood,omitempty"`

	Address     *string `json:"address,omitempty"`
	SubDistrict *string `json:"sub_district,omitempty"`
	District    *string `json:"district,omitempty"`
	Province    *string `json:"province,omitempty"`
	PostalCode  *string `json:"postal_code,omitempty"`

	PreviousPondokName    *string `json:"previous_pondok_name,omitempty"`
	PreviousPondokAddress *string `json:"previous_pondok_address,omitempty"`
	PreviousPondokDiv     *string `json:"previous_pondok_div,omitempty"`
	PreviousPondokTime    *string `json:"previous_pondok_time,omitempty"`

	NIK   *string `json:"nik,omitempty"`
	NoKK  *string `json:"no_kk,omitempty"`
	NISN  *string `json:"nisn,omitempty"`
	NoKIP *string `json:"no_kip,omitempty"`
	NoKKS *string `json:"no_kks,omitempty"`
	NoPKH *string `json:"no_pkh,omitempty"`

	Workplace  *string `json:"workplace,omitempty"`
	Department *string `json:"department,omitempty"`

	HomeStatus *string `json:"home_status,omitempty"`

	Father         *string `json:"father,omitempty"`
	FatherPN       *string `json:"father_pn,omitempty"`
	FatherNIK      *string `json:"father_nik,omitempty"`
	FatherJob      *string `json:"father_job,omitempty"`
	FatherGraduate *string `json:"father_graduate,omitempty"`
	FatherIncome   *string `json:"father_income,omitempty"`

	Mother         *string `json:"mother,omitempty"`
	MotherPN       *string `json:"mother_pn,omitempty"`
	MotherNIK      *string `json:"mother_nik,omitempty"`
	MotherJob      *string `json:"mother_job,omitempty"`
	MotherGraduate *string `json:"mother_graduate,omitempty"`
	MotherIncome   *string `json:"mother_income,omitempty"`

	GuardianRelationship *string `json:"guardian_relationship,omitempty"`
	Guardian             *string `json:"guardian,omitempty"`
	GuardianPN           *string `json:"guardian_pn,omitempty"`
	GuardianNIK          *string `json:"guardian_nik,omitempty"`
	GuardianJob          *string `json:"guardian_job,omitempty"`
	GuardianGraduate     *string `json:"guardian_graduate,omitempty"`
	GuardianIncome       *string `json:"guardian_income,omitempty"`

	Status     string     `json:"status"`
	AcceptedBy *string    `json:"accepted_by,omitempty"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	SantriID   *string    `json:"santri_id,omitempty"`
	NIS        *string    `json:"nis,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type ListPendaftarQuery struct {
	Status       string `form:"status"`
	PsbSettingID string `form:"psb_setting_id"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type ListPendaftarItem struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	PsbSettingID string  `json:"psb_setting_id"`
	Gender       string  `json:"gender"`
	Program      *string `json:"program,omitempty"`
	Status       string  `json:"status"`
	NIS          *string `json:"nis,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
