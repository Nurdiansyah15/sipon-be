package http

import (
	"time"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/fingerprint/application/command"
	"sipon-be/internal/modules/fingerprint/application/dto"
	"sipon-be/internal/modules/fingerprint/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/respond"
	"sipon-be/internal/shared/timeutil"
)

type FingerprintHandler struct {
	simulateScanUC *command.SimulateScanUseCase
	listScansUC    *query.ListScanLogsUseCase
}

func NewFingerprintHandler(
	simulateScanUC *command.SimulateScanUseCase,
	listScansUC *query.ListScanLogsUseCase,
) *FingerprintHandler {
	return &FingerprintHandler{simulateScanUC: simulateScanUC, listScansUC: listScansUC}
}

func (h *FingerprintHandler) SimulateScan(c *gin.Context) {
	var req dto.SimulateScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.simulateScanUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "scan simulasi berhasil dicatat", resp)
}

func (h *FingerprintHandler) ListScans(c *gin.Context) {
	var q dto.ScanLogListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	from, to, err := parseScanRange(q)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	items, err := h.listScansUC.Execute(c.Request.Context(), from, to)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar scan berhasil diambil", items)
}

// parseScanRange mem-parse query from/to (RFC3339). Saat kosong, from default
// ke awal hari ini dan to ke sekarang.
func parseScanRange(q dto.ScanLogListQuery) (time.Time, time.Time, error) {
	from := timeutil.DateOnly(timeutil.Now())
	if q.From != "" {
		t, err := time.Parse(time.RFC3339, q.From)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = t
	}
	to := timeutil.Now()
	if q.To != "" {
		t, err := time.Parse(time.RFC3339, q.To)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = t
	}
	return from, to, nil
}
