package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalEntity "sipon-be/internal/modules/keuangan/domain/journal/entity"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type CreateManualJournalUseCase struct {
	journalRepo journalRepo.JournalRepository
	accountRepo accRepo.AccountRepository
	periodRepo  periodRepo.AccountingPeriodRepository
}

func NewCreateManualJournalUseCase(journalRepo journalRepo.JournalRepository, accountRepo accRepo.AccountRepository, periodRepo periodRepo.AccountingPeriodRepository) *CreateManualJournalUseCase {
	return &CreateManualJournalUseCase{journalRepo: journalRepo, accountRepo: accountRepo, periodRepo: periodRepo}
}

func (uc *CreateManualJournalUseCase) Execute(ctx context.Context, req dto.CreateJournalEntryRequest, postedBy string) (*dto.JournalEntryResponse, error) {
	period, err := uc.periodRepo.FindActive(ctx)
	if err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
	}
	if req.PeriodID != "" {
		period, err = uc.periodRepo.FindByID(ctx, req.PeriodID)
		if err != nil {
			return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
		}
	}
	if !period.CanPost() {
		return nil, application.WrapRepoErr(fmt.Errorf("period is not open"), journalConst.CodeJournalPeriodClosed)
	}

	entryDate, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
	}

	jn := fmt.Sprintf("JRN/%d/%02d/%06d", time.Now().Year(), time.Now().Month(), 1)
	entry, err := journalEntity.NewJournalEntry(
		uuid.New().String(), jn, entryDate,
		req.Description, period.ID, postedBy,
	)
	if err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
	}
	entry.SetSource(journalConst.SourceManual, entry.ID)

	for _, line := range req.Lines {
		acc, err := uc.accountRepo.FindByID(ctx, line.AccountID)
		if err != nil {
			return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
		}
		if err := acc.EnsurePostable(); err != nil {
			return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotBalanced)
		}

		entryLine := journalEntity.NewJournalEntryLine(
			uuid.New().String(), entry.ID,
			acc.ID, acc.Code,
			line.Debit, line.Credit, line.Description,
		)
		entry.AddLine(entryLine)
	}

	if err := entry.Post(); err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalInvalidStatus)
	}

	if err := uc.journalRepo.Save(ctx, entry); err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
	}

	return toJournalEntryResponse(entry), nil
}

func toJournalEntryResponse(entry *journalEntity.JournalEntry) *dto.JournalEntryResponse {
	resp := &dto.JournalEntryResponse{
		ID:            entry.ID,
		JournalNumber: entry.JournalNumber,
		EntryDate:     entry.EntryDate.Format("2006-01-02"),
		Description:   entry.Description,
		PeriodID:      entry.PeriodID,
		TotalDebit:    entry.TotalDebit,
		TotalCredit:   entry.TotalCredit,
		Status:        string(entry.Status),
		PostedBy:      entry.PostedBy,
		CreatedAt:     entry.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     entry.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if entry.SourceType != nil {
		s := string(*entry.SourceType)
		resp.SourceType = &s
	}
	if entry.SourceID != nil {
		resp.SourceID = entry.SourceID
	}
	if entry.PostedAt != nil {
		s := entry.PostedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.PostedAt = &s
	}
	if entry.Lines != nil {
		resp.Lines = make([]dto.JournalLineResponse, len(entry.Lines))
		for i, l := range entry.Lines {
			resp.Lines[i] = dto.JournalLineResponse{
				ID:          l.ID,
				AccountID:   l.AccountID,
				AccountCode: l.AccountCode,
				Description: l.Description,
				Debit:       l.Debit,
				Credit:      l.Credit,
			}
		}
	}
	return resp
}
