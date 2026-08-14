package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrientity "sipon-be/internal/modules/kesantrian/domain/santri/entity"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

// GetSantriDetailUseCase menampilkan detail lengkap satu santri berdasarkan ID
// santri — dipakai endpoint admin (GET /api/v1/web/santri/admin/:id).
type GetSantriDetailUseCase struct {
	santriRepo   santrirepo.SantriRepository
	provisioner  ports.AccountProvisioner
	fileUploader ports.FileUploader
}

func NewGetSantriDetailUseCase(santriRepo santrirepo.SantriRepository, provisioner ports.AccountProvisioner, fileUploader ports.FileUploader) *GetSantriDetailUseCase {
	return &GetSantriDetailUseCase{santriRepo: santriRepo, provisioner: provisioner, fileUploader: fileUploader}
}

func (uc *GetSantriDetailUseCase) Execute(ctx context.Context, santriID string) (*dto.GetSantriResponse, error) {
	santri, err := uc.santriRepo.FindByID(ctx, santriID)
	if err != nil {
		return nil, application.WrapRepoErr(err, santriconstant.CodeSantriNotFound)
	}
	return uc.buildResponse(ctx, santri)
}

func (uc *GetSantriDetailUseCase) buildResponse(ctx context.Context, s *santrientity.Santri) (*dto.GetSantriResponse, error) {
	summary, err := uc.provisioner.GetUserSummary(ctx, s.UserID)
	if err != nil {
		return nil, application.WrapRepoErr(err, application.ErrCodeNotFound)
	}

	avatarURL := resolveAvatarURL(uc.fileUploader, summary.AvatarKey)

	return mapSantriToResponse(s, summary.Username, summary.Email, summary.Fullname, avatarURL), nil
}
