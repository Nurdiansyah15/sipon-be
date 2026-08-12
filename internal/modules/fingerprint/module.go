package fingerprint

import (
	"context"
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/fingerprint/application/command"
	"sipon-be/internal/modules/fingerprint/application/query"
	"sipon-be/internal/modules/fingerprint/infrastructure/persistence"
	fingerprintHTTP "sipon-be/internal/modules/fingerprint/interfaces/http"
	"sipon-be/internal/shared/config"
)

type Module struct {
	handler            *fingerprintHTTP.FingerprintHandler
	listDistinctPinsUC *query.ListDistinctPinsUseCase
	sandboxEnabled     bool
	jwtAuth            gin.HandlerFunc
	principalLoad      gin.HandlerFunc
}

func NewModule(
	db *sql.DB,
	cfg *config.Config,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	scanLogRepo := persistence.NewPostgresScanLogRepository(db)

	simulateScanUC := command.NewSimulateScanUseCase(scanLogRepo)
	listScansUC := query.NewListScanLogsUseCase(scanLogRepo)
	listDistinctPinsUC := query.NewListDistinctPinsUseCase(scanLogRepo)

	handler := fingerprintHTTP.NewFingerprintHandler(simulateScanUC, listScansUC)

	return &Module{
		handler:            handler,
		listDistinctPinsUC: listDistinctPinsUC,
		sandboxEnabled:     cfg.Fingerprint.SandboxEnabled,
		jwtAuth:            jwtAuth,
		principalLoad:      principalLoad,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	fingerprintHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad, m.sandboxEnabled)
}

func (m *Module) ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]ScanPin, error) {
	pins, err := m.listDistinctPinsUC.Execute(ctx, from, to)
	if err != nil {
		return nil, err
	}
	result := make([]ScanPin, 0, len(pins))
	for _, p := range pins {
		result = append(result, ScanPin{PIN: p.PIN, FirstScanAt: p.FirstScanAt})
	}
	return result, nil
}
