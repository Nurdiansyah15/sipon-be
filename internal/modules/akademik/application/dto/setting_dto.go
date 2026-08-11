package dto

type AkademikSettingResponse struct {
	DefaultProgramID   *string          `json:"default_program_id,omitempty"`
	DefaultProgram     *ProgramResponse `json:"default_program,omitempty"`
}

type UpdateAkademikSettingRequest struct {
	DefaultProgramID *string `json:"default_program_id"`
}
