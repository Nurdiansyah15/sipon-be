package query

import (
	"context"
	"log/slog"
	"strings"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

// ListSantriProgramsUseCase menampilkan daftar santri aktif beserta program
// aktifnya untuk kelola program dari sisi module akademik.
type ListSantriProgramsUseCase struct {
	kesantrianReader  ports.KesantrianReader
	santriProgramRepo spRepo.SantriProgramRepository
	programRepo       progRepo.ProgramRepository
}

func NewListSantriProgramsUseCase(
	kesantrianReader ports.KesantrianReader,
	santriProgramRepo spRepo.SantriProgramRepository,
	programRepo progRepo.ProgramRepository,
) *ListSantriProgramsUseCase {
	return &ListSantriProgramsUseCase{
		kesantrianReader:  kesantrianReader,
		santriProgramRepo: santriProgramRepo,
		programRepo:       programRepo,
	}
}

func (uc *ListSantriProgramsUseCase) Execute(ctx context.Context, q dto.SantriProgramListQuery) ([]dto.SantriProgramListItem, *dto.Meta, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 10
	}

	santris, err := uc.kesantrianReader.ListActiveSantriWithUserID(ctx)
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	programMap := make(map[string]*dto.ProgramBrief)
	activePrograms, err := uc.santriProgramRepo.ListActive(ctx)
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	santriProgramBySantri := make(map[string]string, len(activePrograms))
	programIDs := make([]string, 0)
	for _, sp := range activePrograms {
		santriProgramBySantri[sp.SantriID] = sp.ProgramID
		programIDs = append(programIDs, sp.ProgramID)
	}
	if len(programIDs) > 0 {
		programs, err := uc.programRepo.FindByIDs(ctx, programIDs)
		if err != nil {
			slog.Warn("akademik: enrich programs for santri list failed", "error", err)
		} else {
			for _, p := range programs {
				programMap[p.ID] = &dto.ProgramBrief{ID: p.ID, Code: p.Code, Name: p.Name}
			}
		}
	}

	search := ""
	if q.Search != nil {
		search = strings.ToLower(strings.TrimSpace(*q.Search))
	}

	items := make([]dto.SantriProgramListItem, 0)
	for _, s := range santris {
		fullname := ""
		if s.Fullname != nil {
			fullname = *s.Fullname
		}
		nis := ""
		if s.NIS != nil {
			nis = *s.NIS
		}
		if search != "" && !strings.Contains(strings.ToLower(fullname), search) && !strings.Contains(strings.ToLower(nis), search) {
			continue
		}
		programID, hasProgram := santriProgramBySantri[s.SantriID]
		item := dto.SantriProgramListItem{
			SantriID: s.SantriID,
			NIS:      s.NIS,
			Fullname: s.Fullname,
		}
		if hasProgram {
			item.ProgramID = programID
			item.Program = programMap[programID]
		}
		items = append(items, item)
	}

	total := int64(len(items))
	start := (q.Page - 1) * q.Limit
	end := start + q.Limit
	if start > len(items) {
		start = len(items)
	}
	if end > len(items) {
		end = len(items)
	}
	pageItems := items[start:end]

	return pageItems, dto.NewMeta(q.Page, q.Limit, total), nil
}
