package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/command"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/query"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsEntity "sipon-be/internal/modules/keuangan/domain/billingscheme/entity"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	"sipon-be/internal/modules/keuangan/infrastructure/external"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type KeuanganHandler struct {
	createFeeComponentUC   *command.CreateFeeComponentUseCase
	updateFeeComponentUC   *command.UpdateFeeComponentUseCase
	createBillingSchemeUC  *command.CreateBillingSchemeUseCase
	updateBillingSchemeUC  *command.UpdateBillingSchemeUseCase
	assignSchemeToSantriUC *command.AssignSchemeToSantriUseCase
	createInvoiceUC        *command.CreateInvoiceUseCase
	createInvoiceBatchUC   *command.CreateInvoiceBatchUseCase
	cancelInvoiceUC        *command.CancelInvoiceUseCase
	applyAdjustmentUC      *command.ApplyAdjustmentUseCase
	createManualPaymentUC  *command.CreateManualPaymentUseCase
	verifyPaymentUC        *command.VerifyPaymentUseCase
	rejectPaymentUC        *command.RejectPaymentUseCase

	createAccountUC       *command.CreateAccountUseCase
	updateAccountUC       *command.UpdateAccountUseCase
	createManualJournalUC *command.CreateManualJournalUseCase
	cancelJournalUC       *command.CancelJournalUseCase
	createPeriodUC        *command.CreatePeriodUseCase
	closePeriodUC         *command.ClosePeriodUseCase
	reopenPeriodUC        *command.ReopenPeriodUseCase
	lockPeriodUC          *command.LockPeriodUseCase

	createBillingPeriodUC *command.CreateBillingPeriodUseCase
	openBillingPeriodUC   *command.OpenBillingPeriodUseCase
	closeBillingPeriodUC  *command.CloseBillingPeriodUseCase

	listFeeComponentsUC  *query.ListFeeComponentsUseCase
	listBillingSchemesUC *query.ListBillingSchemesUseCase
	getBillingSchemeUC   *query.GetBillingSchemeUseCase
	listInvoicesUC       *query.ListInvoicesUseCase
	getInvoiceUC         *query.GetInvoiceUseCase
	myInvoicesUC         *query.MyInvoicesUseCase
	listPaymentsUC       *query.ListPaymentsUseCase
	getPaymentUC         *query.GetPaymentUseCase
	myPaymentsUC         *query.MyPaymentsUseCase
	listAccountsUC       *query.ListAccountsUseCase
	getAccountUC         *query.GetAccountUseCase
	listJournalEntriesUC *query.ListJournalEntriesUseCase
	getJournalEntryUC    *query.GetJournalEntryUseCase
	listPeriodsUC        *query.ListPeriodsUseCase
	getActivePeriodUC    *query.GetActivePeriodUseCase
	listAssignmentsUC    *query.ListAssignmentsUseCase
	listBillingPeriodsUC *query.ListBillingPeriodsUseCase
	getBillingPeriodUC   *query.GetBillingPeriodUseCase
	listBillingBatchesUC *query.ListBillingBatchesUseCase
	getBillingBatchUC    *query.GetBillingBatchUseCase

	reportSummaryUC         *query.ReportSummaryUseCase
	reportOutstandingUC     *query.ReportOutstandingUseCase
	reportLedgerUC          *query.ReportLedgerUseCase
	reportTrialBalanceUC    *query.ReportTrialBalanceUseCase
	reportBalanceSheetUC    *query.ReportBalanceSheetUseCase
	reportIncomeStatementUC *query.ReportIncomeStatementUseCase

	feeComponentRepo  feeRepo.FeeComponentRepository
	billingSchemeRepo bsRepo.BillingSchemeRepository
	billingPeriodRepo bpRepo.BillingPeriodRepository
	accountRepo       accRepo.AccountRepository
	invoiceRepo       invRepo.InvoiceRepository
}

func NewKeuanganHandler(
	createFeeComponentUC *command.CreateFeeComponentUseCase,
	updateFeeComponentUC *command.UpdateFeeComponentUseCase,
	createBillingSchemeUC *command.CreateBillingSchemeUseCase,
	updateBillingSchemeUC *command.UpdateBillingSchemeUseCase,
	assignSchemeToSantriUC *command.AssignSchemeToSantriUseCase,
	createInvoiceUC *command.CreateInvoiceUseCase,
	createInvoiceBatchUC *command.CreateInvoiceBatchUseCase,
	cancelInvoiceUC *command.CancelInvoiceUseCase,
	applyAdjustmentUC *command.ApplyAdjustmentUseCase,
	createManualPaymentUC *command.CreateManualPaymentUseCase,
	verifyPaymentUC *command.VerifyPaymentUseCase,
	rejectPaymentUC *command.RejectPaymentUseCase,
	createAccountUC *command.CreateAccountUseCase,
	updateAccountUC *command.UpdateAccountUseCase,
	createManualJournalUC *command.CreateManualJournalUseCase,
	cancelJournalUC *command.CancelJournalUseCase,
	createPeriodUC *command.CreatePeriodUseCase,
	closePeriodUC *command.ClosePeriodUseCase,
	reopenPeriodUC *command.ReopenPeriodUseCase,
	lockPeriodUC *command.LockPeriodUseCase,
	createBillingPeriodUC *command.CreateBillingPeriodUseCase,
	openBillingPeriodUC *command.OpenBillingPeriodUseCase,
	closeBillingPeriodUC *command.CloseBillingPeriodUseCase,
	listFeeComponentsUC *query.ListFeeComponentsUseCase,
	listBillingSchemesUC *query.ListBillingSchemesUseCase,
	getBillingSchemeUC *query.GetBillingSchemeUseCase,
	listInvoicesUC *query.ListInvoicesUseCase,
	getInvoiceUC *query.GetInvoiceUseCase,
	myInvoicesUC *query.MyInvoicesUseCase,
	listPaymentsUC *query.ListPaymentsUseCase,
	getPaymentUC *query.GetPaymentUseCase,
	myPaymentsUC *query.MyPaymentsUseCase,
	listAccountsUC *query.ListAccountsUseCase,
	getAccountUC *query.GetAccountUseCase,
	listJournalEntriesUC *query.ListJournalEntriesUseCase,
	getJournalEntryUC *query.GetJournalEntryUseCase,
	listPeriodsUC *query.ListPeriodsUseCase,
	getActivePeriodUC *query.GetActivePeriodUseCase,
	listAssignmentsUC *query.ListAssignmentsUseCase,
	listBillingPeriodsUC *query.ListBillingPeriodsUseCase,
	getBillingPeriodUC *query.GetBillingPeriodUseCase,
	listBillingBatchesUC *query.ListBillingBatchesUseCase,
	getBillingBatchUC *query.GetBillingBatchUseCase,
	reportSummaryUC *query.ReportSummaryUseCase,
	reportOutstandingUC *query.ReportOutstandingUseCase,
	reportLedgerUC *query.ReportLedgerUseCase,
	reportTrialBalanceUC *query.ReportTrialBalanceUseCase,
	reportBalanceSheetUC *query.ReportBalanceSheetUseCase,
	reportIncomeStatementUC *query.ReportIncomeStatementUseCase,
	feeComponentRepo feeRepo.FeeComponentRepository,
	billingSchemeRepo bsRepo.BillingSchemeRepository,
	billingPeriodRepo bpRepo.BillingPeriodRepository,
	accountRepo accRepo.AccountRepository,
	invoiceRepo invRepo.InvoiceRepository,
) *KeuanganHandler {
	return &KeuanganHandler{
		createFeeComponentUC:   createFeeComponentUC,
		updateFeeComponentUC:   updateFeeComponentUC,
		createBillingSchemeUC:  createBillingSchemeUC,
		updateBillingSchemeUC:  updateBillingSchemeUC,
		assignSchemeToSantriUC: assignSchemeToSantriUC,
		createInvoiceUC:        createInvoiceUC,
		createInvoiceBatchUC:   createInvoiceBatchUC,
		cancelInvoiceUC:        cancelInvoiceUC,
		applyAdjustmentUC:      applyAdjustmentUC,
		createManualPaymentUC:  createManualPaymentUC,
		verifyPaymentUC:        verifyPaymentUC,
		rejectPaymentUC:        rejectPaymentUC,
		createAccountUC:        createAccountUC,
		updateAccountUC:        updateAccountUC,
		createManualJournalUC:  createManualJournalUC,
		cancelJournalUC:        cancelJournalUC,
		createPeriodUC:         createPeriodUC,
		closePeriodUC:          closePeriodUC,
		reopenPeriodUC:         reopenPeriodUC,
		lockPeriodUC:           lockPeriodUC,
		createBillingPeriodUC:  createBillingPeriodUC,
		openBillingPeriodUC:    openBillingPeriodUC,
		closeBillingPeriodUC:   closeBillingPeriodUC,
		listFeeComponentsUC:    listFeeComponentsUC,
		listBillingSchemesUC:   listBillingSchemesUC,
		getBillingSchemeUC:     getBillingSchemeUC,
		listInvoicesUC:         listInvoicesUC,
		getInvoiceUC:           getInvoiceUC,
		myInvoicesUC:           myInvoicesUC,
		listPaymentsUC:         listPaymentsUC,
		getPaymentUC:           getPaymentUC,
		myPaymentsUC:           myPaymentsUC,
		listAccountsUC:         listAccountsUC,
		getAccountUC:           getAccountUC,
		listJournalEntriesUC:   listJournalEntriesUC,
		getJournalEntryUC:      getJournalEntryUC,
		listPeriodsUC:          listPeriodsUC,
		getActivePeriodUC:      getActivePeriodUC,
		listAssignmentsUC:      listAssignmentsUC,
		listBillingPeriodsUC:   listBillingPeriodsUC,
		getBillingPeriodUC:     getBillingPeriodUC,
		listBillingBatchesUC:   listBillingBatchesUC,
		getBillingBatchUC:      getBillingBatchUC,
		reportSummaryUC:         reportSummaryUC,
		reportOutstandingUC:     reportOutstandingUC,
		reportLedgerUC:          reportLedgerUC,
		reportTrialBalanceUC:    reportTrialBalanceUC,
		reportBalanceSheetUC:    reportBalanceSheetUC,
		reportIncomeStatementUC: reportIncomeStatementUC,
		feeComponentRepo:        feeComponentRepo,
		billingSchemeRepo:      billingSchemeRepo,
		billingPeriodRepo:      billingPeriodRepo,
		accountRepo:            accountRepo,
		invoiceRepo:            invoiceRepo,
	}
}

func (h *KeuanganHandler) ListFeeComponents(c *gin.Context) {
	var req dto.FeeComponentListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listFeeComponentsUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar komponen biaya berhasil diambil", items, meta)
}

func (h *KeuanganHandler) ListAssignments(c *gin.Context) {
	items, err := h.listAssignmentsUC.Execute(c.Request.Context())
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar penetapan skema berhasil diambil", items, nil)
}

func (h *KeuanganHandler) CreateFeeComponent(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateFeeComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createFeeComponentUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "komponen biaya berhasil dibuat", resp)
}

func (h *KeuanganHandler) UpdateFeeComponent(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateFeeComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateFeeComponentUC.Execute(c.Request.Context(), id, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "komponen biaya berhasil diubah", resp)
}

func (h *KeuanganHandler) DeleteFeeComponent(c *gin.Context) {
	id := c.Param("id")
	fc, err := h.feeComponentRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, kernel.New("ERR_NOT_FOUND"))
		return
	}
	fc.Deactivate()
	if err := h.feeComponentRepo.Update(c.Request.Context(), fc); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "komponen biaya berhasil dihapus", nil)
}

func (h *KeuanganHandler) ListBillingSchemes(c *gin.Context) {
	var req dto.BillingSchemeListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listBillingSchemesUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar skema billing berhasil diambil", items, meta)
}

func (h *KeuanganHandler) CreateBillingScheme(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateBillingSchemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createBillingSchemeUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "skema billing berhasil dibuat", resp)
}

func (h *KeuanganHandler) UpdateBillingScheme(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateBillingSchemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateBillingSchemeUC.Execute(c.Request.Context(), id, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "skema billing berhasil diubah", resp)
}

func (h *KeuanganHandler) DeleteBillingScheme(c *gin.Context) {
	id := c.Param("id")
	scheme, err := h.billingSchemeRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, kernel.New("ERR_NOT_FOUND"))
		return
	}
	scheme.Deactivate()
	if err := h.billingSchemeRepo.Update(c.Request.Context(), scheme); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "skema billing berhasil dihapus", nil)
}

func (h *KeuanganHandler) GetBillingScheme(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getBillingSchemeUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "detail skema billing berhasil diambil", resp)
}

func (h *KeuanganHandler) AddSchemeItem(c *gin.Context) {
	schemeID := c.Param("id")
	var req dto.AddSchemeItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	item, err := bsEntity.NewBillingSchemeItem(
		uuid.New().String(), schemeID, req.FeeComponentID,
		req.AmountOverride, req.IsRequired, req.SortOrder,
	)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.billingSchemeRepo.AddItems(c.Request.Context(), schemeID, []*bsEntity.BillingSchemeItem{item}); err != nil {
		httperror.Handle(c, application.WrapConflictErr(err, bsConst.CodeSchemeItemDuplicate))
		return
	}
	resp, err := h.getBillingSchemeUC.Execute(c.Request.Context(), schemeID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "item skema berhasil ditambahkan", resp)
}

func (h *KeuanganHandler) RemoveSchemeItem(c *gin.Context) {
	schemeID := c.Param("id")
	itemID := c.Param("itemId")
	if err := h.billingSchemeRepo.RemoveItem(c.Request.Context(), schemeID, itemID); err != nil {
		httperror.Handle(c, application.WrapRepoErr(err, bsConst.CodeSchemeItemNotFound))
		return
	}
	resp, err := h.getBillingSchemeUC.Execute(c.Request.Context(), schemeID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "item skema berhasil dihapus", resp)
}

func (h *KeuanganHandler) AssignSchemeToSantri(c *gin.Context) {
	var req dto.AssignSchemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	userID := middleware.GetUserID(c)
	cmd := command.AssignSchemeCmd{
		SantriID:        req.SantriID,
		BillingSchemeID: req.BillingSchemeID,
		EffectiveFrom:   req.EffectiveFrom,
		EffectiveUntil:  req.EffectiveUntil,
		AssignedBy:      userID,
	}
	resp, err := h.assignSchemeToSantriUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, resp.Message, resp)
}

func (h *KeuanganHandler) ListInvoices(c *gin.Context) {
	var req dto.InvoiceListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listInvoicesUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar invoice berhasil diambil", items, meta)
}

func (h *KeuanganHandler) CreateInvoice(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	cmd := command.CreateInvoiceCmd{
		SantriID:        req.SantriID,
		FeeComponentID:  req.FeeComponentID,
		BillingPeriodID: req.BillingPeriodID,
		Amount:          req.Amount,
		DueDate:         req.DueDate,
		Notes:           req.Notes,
		CreatedBy:       userID,
		Issue:           true,
	}
	resp, err := h.createInvoiceUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "invoice berhasil dibuat", resp)
}

func (h *KeuanganHandler) CreateInvoiceBatch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateInvoiceBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	cmd := command.CreateInvoiceBatchCmd{
		BillingSchemeID: req.BillingSchemeID,
		BillingPeriodID: req.BillingPeriodID,
		DueDate:         req.DueDate,
		CreatedBy:       userID,
	}
	resp, err := h.createInvoiceBatchUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "batch invoice berhasil dibuat", resp)
}

func (h *KeuanganHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getInvoiceUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "detail invoice berhasil diambil", resp)
}

func (h *KeuanganHandler) CancelInvoice(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.cancelInvoiceUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "invoice berhasil dibatalkan", resp)
}

func (h *KeuanganHandler) ApplyAdjustment(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	var req dto.ApplyAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.applyAdjustmentUC.Execute(c.Request.Context(), id, req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "penyesuaian berhasil diterapkan", resp)
}

func (h *KeuanganHandler) ListPayments(c *gin.Context) {
	var req dto.PaymentListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listPaymentsUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar pembayaran berhasil diambil", items, meta)
}

func (h *KeuanganHandler) GetPayment(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getPaymentUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "detail pembayaran berhasil diambil", resp)
}

func (h *KeuanganHandler) CreateManualPayment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateManualPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createManualPaymentUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "pembayaran manual berhasil dibuat", resp)
}

func (h *KeuanganHandler) VerifyPayment(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	resp, err := h.verifyPaymentUC.Execute(c.Request.Context(), id, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "pembayaran berhasil diverifikasi", resp)
}

func (h *KeuanganHandler) RejectPayment(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.rejectPaymentUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "pembayaran berhasil ditolak", resp)
}

func (h *KeuanganHandler) MyInvoices(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.InvoiceListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.myInvoicesUC.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar invoice berhasil diambil", items, meta)
}

func (h *KeuanganHandler) MyPayments(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.PaymentListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.myPaymentsUC.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar pembayaran berhasil diambil", items, meta)
}

func (h *KeuanganHandler) ListAccounts(c *gin.Context) {
	var req dto.AccountListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listAccountsUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar akun berhasil diambil", items, meta)
}

func (h *KeuanganHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getAccountUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "detail akun berhasil diambil", resp)
}

func (h *KeuanganHandler) CreateAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createAccountUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "akun berhasil dibuat", resp)
}

func (h *KeuanganHandler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateAccountUC.Execute(c.Request.Context(), id, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "akun berhasil diubah", resp)
}

func (h *KeuanganHandler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	acc, err := h.accountRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, kernel.New("ERR_NOT_FOUND"))
		return
	}
	if err := acc.Deactivate(); err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.accountRepo.Update(c.Request.Context(), acc); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "akun berhasil dihapus", nil)
}

func (h *KeuanganHandler) ListJournalEntries(c *gin.Context) {
	var req dto.JournalListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listJournalEntriesUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar jurnal berhasil diambil", items, meta)
}

func (h *KeuanganHandler) GetJournalEntry(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getJournalEntryUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "detail jurnal berhasil diambil", resp)
}

func (h *KeuanganHandler) CreateJournalEntry(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateJournalEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createManualJournalUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "jurnal berhasil dibuat", resp)
}

func (h *KeuanganHandler) CancelJournalEntry(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.cancelJournalUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *KeuanganHandler) ListPeriods(c *gin.Context) {
	var req dto.PeriodListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listPeriodsUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar periode berhasil diambil", items, meta)
}

func (h *KeuanganHandler) GetActivePeriod(c *gin.Context) {
	resp, err := h.getActivePeriodUC.Execute(c.Request.Context())
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode aktif berhasil diambil", resp)
}

func (h *KeuanganHandler) CreatePeriod(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreatePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createPeriodUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "periode berhasil dibuat", resp)
}

func (h *KeuanganHandler) ClosePeriod(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	resp, err := h.closePeriodUC.Execute(c.Request.Context(), id, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode berhasil ditutup", resp)
}

func (h *KeuanganHandler) ReopenPeriod(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.reopenPeriodUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode berhasil dibuka kembali", resp)
}

func (h *KeuanganHandler) LockPeriod(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	resp, err := h.lockPeriodUC.Execute(c.Request.Context(), id, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode berhasil dikunci permanen", resp)
}

func (h *KeuanganHandler) ListBillingPeriods(c *gin.Context) {
	var req dto.BillingPeriodListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listBillingPeriodsUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar periode tagihan berhasil diambil", items, meta)
}

func (h *KeuanganHandler) GetBillingPeriod(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getBillingPeriodUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "detail periode tagihan berhasil diambil", resp)
}

func (h *KeuanganHandler) CreateBillingPeriod(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateBillingPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createBillingPeriodUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "periode tagihan berhasil dibuat", resp)
}

func (h *KeuanganHandler) OpenBillingPeriod(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.openBillingPeriodUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode tagihan berhasil dibuka", resp)
}

func (h *KeuanganHandler) CloseBillingPeriod(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.closeBillingPeriodUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode tagihan berhasil ditutup", resp)
}

func (h *KeuanganHandler) ListBillingBatches(c *gin.Context) {
	var req dto.BillingBatchListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listBillingBatchesUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar batch tagihan berhasil diambil", items, meta)
}

func (h *KeuanganHandler) GetBillingBatch(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getBillingBatchUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "detail batch tagihan berhasil diambil", resp)
}

func (h *KeuanganHandler) DownloadReceipt(c *gin.Context) {
	id := c.Param("id")

	payment, err := h.getPaymentUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	inv, err := h.invoiceRepo.FindByID(c.Request.Context(), payment.InvoiceID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	fc, err := h.feeComponentRepo.FindByID(c.Request.Context(), inv.FeeComponentID)
	feeName := inv.FeeComponentID
	if err == nil {
		feeName = fc.Name
	}

	bp, err := h.billingPeriodRepo.FindByID(c.Request.Context(), inv.BillingPeriodID)
	billingPeriodName := inv.BillingPeriodID
	if err == nil {
		billingPeriodName = bp.Name
	}

	pdfData := external.ReceiptData{
		ReceiptNumber:   payment.PaymentNumber,
		PaymentDate:     payment.PaymentDate,
		InvoiceNumber:   inv.InvoiceNumber,
		FeeComponent:    feeName,
		BillingPeriod:   billingPeriodName,
		Amount:          payment.Amount,
		PaymentMethod:   payment.Method,
		ReferenceNumber: func() string {
			if payment.ReferenceNumber != nil {
				return *payment.ReferenceNumber
			}
			return ""
		}(),
		SantriName: inv.SantriID,
		NIS:        "",
		VerifiedBy: func() string {
			if payment.VerifiedBy != nil {
				return *payment.VerifiedBy
			}
			return ""
		}(),
	}

	pdf, err := external.GenerateReceiptPDF(pdfData)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	filename := fmt.Sprintf("kwitansi-%s.pdf", payment.PaymentNumber)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(200, "application/pdf", pdf)
}

func (h *KeuanganHandler) ReportSummary(c *gin.Context) {
	var req dto.InvoiceSummaryQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, err := h.reportSummaryUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "rekap tagihan berhasil diambil", items)
}

func (h *KeuanganHandler) ReportOutstanding(c *gin.Context) {
	var req dto.OutstandingListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.reportOutstandingUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar tunggakan berhasil diambil", items, meta)
}

func (h *KeuanganHandler) ReportLedger(c *gin.Context) {
	var req dto.LedgerQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.reportLedgerUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "buku besar berhasil diambil", resp)
}

func (h *KeuanganHandler) ReportTrialBalance(c *gin.Context) {
	var req dto.TrialBalanceQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.reportTrialBalanceUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "neraca saldo berhasil diambil", resp)
}

func (h *KeuanganHandler) ReportBalanceSheet(c *gin.Context) {
	var req dto.BalanceSheetQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.reportBalanceSheetUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "neraca berhasil diambil", resp)
}

func (h *KeuanganHandler) ReportIncomeStatement(c *gin.Context) {
	var req dto.IncomeStatementQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.reportIncomeStatementUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "laporan laba rugi berhasil diambil", resp)
}
