package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/academic_period/constant"
	"sipon-be/internal/modules/akademik/domain/academic_period/entity"
	repo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateAcademicPeriodUseCase struct {
	periodRepo repo.AcademicPeriodRepository
}

func NewCreateAcademicPeriodUseCase(periodRepo repo.AcademicPeriodRepository) *CreateAcademicPeriodUseCase {
	return &CreateAcademicPeriodUseCase{periodRepo: periodRepo}
}

func (uc *CreateAcademicPeriodUseCase) Execute(ctx context.Context, req dto.CreateAcademicPeriodRequest) (*dto.AcademicPeriodResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	existing, err := uc.periodRepo.FindByCode(ctx, req.Code)
	if err != nil && !application.IsNotFoundErr(err, constant.CodeAcademicPeriodNotFound) {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if existing != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	period, err := entity.NewAcademicPeriod(uuid.NewString(), req.Code, req.Name, startDate, endDate)
	if err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeAcademicPeriodInvalidRange)
	}
	if err := uc.periodRepo.Save(ctx, period); err != nil {
		return nil, application.WrapConflictErr(err, constant.CodeAcademicPeriodDuplicate)
	}
	return MapAcademicPeriodToResponse(period), nil
}

func MapAcademicPeriodToResponse(p *entity.AcademicPeriod) *dto.AcademicPeriodResponse {
	return &dto.AcademicPeriodResponse{
		ID:        p.ID,
		Code:      p.Code,
		Name:      p.Name,
		StartDate: p.StartDate.Format("2006-01-02"),
		EndDate:   p.EndDate.Format("2006-01-02"),
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
