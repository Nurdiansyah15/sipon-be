package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
)

type ActivityScheduleRepository interface {
	Save(ctx context.Context, schedule *entity.ActivitySchedule) error
	Update(ctx context.Context, schedule *entity.ActivitySchedule) error
	FindByID(ctx context.Context, id string) (*entity.ActivitySchedule, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.ActivitySchedule, error)
	ListByActivityPeriod(ctx context.Context, activityPeriodID string) ([]*entity.ActivitySchedule, error)
	// ListByActivityPeriodIDs returns schedules for the given activity period ids.
	ListByActivityPeriodIDs(ctx context.Context, activityPeriodIDs []string) ([]*entity.ActivitySchedule, error)

	ReplaceWeeklies(ctx context.Context, scheduleID string, days []constant.DayOfWeek) error
	ReplaceMonthlies(ctx context.Context, scheduleID string, days []int) error
	ReplaceYearlies(ctx context.Context, scheduleID string, dates []entity.YearlyDate) error

	ListWeeklies(ctx context.Context, scheduleID string) ([]entity.ActivityScheduleWeekly, error)
	ListMonthlies(ctx context.Context, scheduleID string) ([]entity.ActivityScheduleMonthly, error)
	ListYearlies(ctx context.Context, scheduleID string) ([]entity.ActivityScheduleYearly, error)
}
