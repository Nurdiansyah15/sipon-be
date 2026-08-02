package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrientity "sipon-be/internal/modules/kesantrian/domain/santri/entity"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	santrivo "sipon-be/internal/modules/kesantrian/domain/santri/valueobject"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

// ImportSantriRow is one row parsed from an uploaded spreadsheet — NIS is
// mandatory, Profile carries every other santri field (reusing
// dto.UpdateSantriRequest's shape, since it's already the exact "everything
// except NIS/username/email/phone" set). Document upload is intentionally
// NOT part of this row — documents can't be represented in a spreadsheet
// cell, so they stay a separate, per-santri upload flow.
type ImportSantriRow struct {
	RowNumber int
	NIS       string
	Profile   dto.UpdateSantriRequest
}

// ImportSantriUseCase processes a batch of rows independently — one row
// failing (bad NIS, duplicate, account-provisioning error) does not abort
// the rest. Every row gets its own result entry so the admin can see exactly
// which rows succeeded and why others didn't.
type ImportSantriUseCase struct {
	santriRepo  santrirepo.SantriRepository
	provisioner ports.AccountProvisioner
	transactor  ports.Transactor
}

func NewImportSantriUseCase(santriRepo santrirepo.SantriRepository, provisioner ports.AccountProvisioner, transactor ports.Transactor) *ImportSantriUseCase {
	return &ImportSantriUseCase{santriRepo: santriRepo, provisioner: provisioner, transactor: transactor}
}

func (uc *ImportSantriUseCase) Execute(ctx context.Context, rows []ImportSantriRow) (*dto.ImportSantriResponse, error) {
	resp := &dto.ImportSantriResponse{Items: make([]dto.ImportSantriResultItem, 0, len(rows))}

	seenInFile := make(map[string]bool)

	for _, row := range rows {
		item := dto.ImportSantriResultItem{RowNumber: row.RowNumber, NIS: row.NIS}

		nis, err := santrivo.NewNIS(row.NIS)
		if err != nil {
			item.Status = "error"
			item.Message = "format NIS tidak valid"
			resp.Items = append(resp.Items, item)
			continue
		}

		if seenInFile[nis.String()] {
			item.Status = "error"
			item.Message = "NIS duplikat di dalam file"
			resp.Items = append(resp.Items, item)
			continue
		}

		if _, err := uc.santriRepo.FindByNIS(ctx, nis.String()); err == nil {
			item.Status = "error"
			item.Message = "NIS sudah terdaftar"
			resp.Items = append(resp.Items, item)
			continue
		} else {
			var ke *kernel.AppError
			if !errors.As(err, &ke) || ke.Code != santriconstant.CodeSantriNotFound {
				item.Status = "error"
				item.Message = "gagal memeriksa NIS: " + err.Error()
				resp.Items = append(resp.Items, item)
				continue
			}
		}

		email := nis.String() + "@santri.sipon"
		acc, err := uc.provisioner.CreateAccountWithNIS(ctx, identity.CreateAccountInput{
			Username: nis.String(),
			Email:    email,
			Fullname: row.Profile.Fullname,
			NISValue: nis.String(),
		})
		if err != nil {
			item.Status = "error"
			item.Message = "gagal membuat akun: " + err.Error()
			resp.Items = append(resp.Items, item)
			continue
		}

		santri, err := santrientity.NewSantri(uuid.NewString(), acc.UserID)
		if err != nil {
			item.Status = "error"
			item.Message = "akun berhasil dibuat tapi profil santri gagal disiapkan: " + err.Error()
			resp.Items = append(resp.Items, item)
			continue
		}
		santri.SetNIS(nis)
		applySantriUpdate(santri, row.Profile)

		if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
			return uc.santriRepo.Save(txCtx, santri)
		}); err != nil {
			item.Status = "error"
			item.Message = "akun berhasil dibuat tapi profil santri gagal disimpan: " + err.Error()
			resp.Items = append(resp.Items, item)
			continue
		}

		seenInFile[nis.String()] = true
		item.Status = "success"
		item.UserID = acc.UserID
		item.SantriID = santri.ID
		item.GeneratedPassword = acc.GeneratedPassword
		resp.Items = append(resp.Items, item)
	}

	for _, item := range resp.Items {
		if item.Status == "success" {
			resp.SuccessCount++
		} else {
			resp.ErrorCount++
		}
	}

	return resp, nil
}
