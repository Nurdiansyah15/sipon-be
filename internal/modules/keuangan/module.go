package keuangan

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/keuangan/application/command"
	"sipon-be/internal/modules/keuangan/application/query"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	"sipon-be/internal/modules/keuangan/infrastructure/kesantriangateway"
	"sipon-be/internal/modules/keuangan/infrastructure/persistence"
	keuanganHTTP "sipon-be/internal/modules/keuangan/interfaces/http"
	"sipon-be/internal/modules/kesantrian"
	"sipon-be/internal/shared/config"
)

type Module struct {
	handler        *keuanganHTTP.KeuanganHandler
	transactor     *persistence.PostgresTransactor
	invoiceRepo    *persistence.PostgresInvoiceRepository
	jwtAuth        gin.HandlerFunc
	principalLoad  gin.HandlerFunc
}

func NewModule(
	db *sql.DB,
	cfg *config.Config,
	kesantrianContract kesantrian.Contract,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	kesantrianReader := kesantriangateway.New(kesantrianContract)
	feeComponentRepo := persistence.NewPostgresFeeComponentRepository(db)
	billingSchemeRepo := persistence.NewPostgresBillingSchemeRepository(db)
	billingPeriodRepo := persistence.NewPostgresBillingPeriodRepository(db)
	billingBatchRepo := persistence.NewPostgresBillingBatchRepository(db)
	billingBatchTargetRepo := persistence.NewPostgresBillingBatchTargetRepository(db)
	assignmentRepo := persistence.NewPostgresSantriBillingAssignmentRepository(db)
	invoiceRepo := persistence.NewPostgresInvoiceRepository(db)
	paymentRepo := persistence.NewPostgresPaymentRepository(db)
	adjustmentRepo := persistence.NewPostgresAdjustmentRepository(db)
	accountRepo := persistence.NewPostgresAccountRepository(db)
	journalRepo := persistence.NewPostgresJournalRepository(db)
	periodRepo := persistence.NewPostgresAccountingPeriodRepository(db)
	transactor := persistence.NewPostgresTransactor(db)
	assignmentReader := persistence.NewPostgresAssignmentReader(db)
	reportReader := persistence.NewPostgresReportReader(db)

	autoPostingService := journalService.NewAutoPostingService(journalRepo, accountRepo, periodRepo)

	createFeeComponentUC := command.NewCreateFeeComponentUseCase(feeComponentRepo, accountRepo)
	updateFeeComponentUC := command.NewUpdateFeeComponentUseCase(feeComponentRepo, accountRepo)
	createBillingSchemeUC := command.NewCreateBillingSchemeUseCase(billingSchemeRepo)
	updateBillingSchemeUC := command.NewUpdateBillingSchemeUseCase(billingSchemeRepo)
	assignSchemeToSantriUC := command.NewAssignSchemeToSantriUseCase(assignmentRepo, billingSchemeRepo)
	updateAssignmentUC := command.NewUpdateAssignmentUseCase(assignmentRepo, billingSchemeRepo)
	createInvoiceUC := command.NewCreateInvoiceUseCase(invoiceRepo, feeComponentRepo, assignmentRepo, billingPeriodRepo, kesantrianReader, transactor, autoPostingService)
	createInvoiceBatchUC := command.NewCreateInvoiceBatchUseCase(invoiceRepo, feeComponentRepo, billingSchemeRepo, assignmentRepo, billingPeriodRepo, billingBatchRepo, billingBatchTargetRepo, kesantrianReader, transactor, autoPostingService)
	cancelInvoiceUC := command.NewCancelInvoiceUseCase(invoiceRepo, feeComponentRepo, transactor, autoPostingService)
	applyAdjustmentUC := command.NewApplyAdjustmentUseCase(adjustmentRepo, invoiceRepo, feeComponentRepo, transactor, autoPostingService)
	createManualPaymentUC := command.NewCreateManualPaymentUseCase(paymentRepo, invoiceRepo, accountRepo)
	verifyPaymentUC := command.NewVerifyPaymentUseCase(paymentRepo, invoiceRepo, feeComponentRepo, transactor, autoPostingService)
	rejectPaymentUC := command.NewRejectPaymentUseCase(paymentRepo)
	createAccountUC := command.NewCreateAccountUseCase(accountRepo)
	updateAccountUC := command.NewUpdateAccountUseCase(accountRepo)
	createManualJournalUC := command.NewCreateManualJournalUseCase(journalRepo, accountRepo, periodRepo, transactor)
	cancelJournalUC := command.NewCancelJournalUseCase(journalRepo, periodRepo)
	createPeriodUC := command.NewCreatePeriodUseCase(periodRepo)
	closePeriodUC := command.NewClosePeriodUseCase(periodRepo, accountRepo, journalRepo, transactor)
	reopenPeriodUC := command.NewReopenPeriodUseCase(periodRepo, journalRepo, transactor)
	lockPeriodUC := command.NewLockPeriodUseCase(periodRepo)
	createBillingPeriodUC := command.NewCreateBillingPeriodUseCase(billingPeriodRepo)
	openBillingPeriodUC := command.NewOpenBillingPeriodUseCase(billingPeriodRepo)
	closeBillingPeriodUC := command.NewCloseBillingPeriodUseCase(billingPeriodRepo)

	listFeeComponentsUC := query.NewListFeeComponentsUseCase(feeComponentRepo, accountRepo)
	listBillingSchemesUC := query.NewListBillingSchemesUseCase(billingSchemeRepo)
	getBillingSchemeUC := query.NewGetBillingSchemeUseCase(billingSchemeRepo, feeComponentRepo)
	listInvoicesUC := query.NewListInvoicesUseCase(invoiceRepo, feeComponentRepo, billingPeriodRepo)
	getInvoiceUC := query.NewGetInvoiceUseCase(invoiceRepo, feeComponentRepo, billingPeriodRepo, paymentRepo, adjustmentRepo)
	myInvoicesUC := query.NewMyInvoicesUseCase(invoiceRepo, feeComponentRepo, billingPeriodRepo)
	mySummaryUC := query.NewMyInvoiceSummaryUseCase(invoiceRepo)
	listBillingPeriodsUC := query.NewListBillingPeriodsUseCase(billingPeriodRepo)
	getBillingPeriodUC := query.NewGetBillingPeriodUseCase(billingPeriodRepo)
	listBillingBatchesUC := query.NewListBillingBatchesUseCase(billingBatchRepo)
	getBillingBatchUC := query.NewGetBillingBatchUseCase(billingBatchRepo, billingBatchTargetRepo)
	listPaymentsUC := query.NewListPaymentsUseCase(paymentRepo, invoiceRepo, accountRepo)
	getPaymentUC := query.NewGetPaymentUseCase(paymentRepo, invoiceRepo, accountRepo)
	myPaymentsUC := query.NewMyPaymentsUseCase(paymentRepo, invoiceRepo, accountRepo)
	listAccountsUC := query.NewListAccountsUseCase(accountRepo)
	getAccountUC := query.NewGetAccountUseCase(accountRepo)
	listJournalEntriesUC := query.NewListJournalEntriesUseCase(journalRepo)
	getJournalEntryUC := query.NewGetJournalEntryUseCase(journalRepo)
	getJournalEntryBySourceUC := query.NewGetJournalEntryBySourceUseCase(journalRepo)
	listPeriodsUC := query.NewListPeriodsUseCase(periodRepo)
	getActivePeriodUC := query.NewGetActivePeriodUseCase(periodRepo)
	listAssignmentsUC := query.NewListAssignmentsUseCase(assignmentReader, billingSchemeRepo)
	reportSummaryUC := query.NewReportSummaryUseCase(reportReader)
	reportOutstandingUC := query.NewReportOutstandingUseCase(reportReader)
	reportLedgerUC := query.NewReportLedgerUseCase(reportReader, accountRepo, periodRepo)
	reportTrialBalanceUC := query.NewReportTrialBalanceUseCase(reportReader, accountRepo, periodRepo)
	reportBalanceSheetUC := query.NewReportBalanceSheetUseCase(reportReader, accountRepo, periodRepo)
	reportIncomeStatementUC := query.NewReportIncomeStatementUseCase(reportReader, accountRepo, periodRepo)

	handler := keuanganHTTP.NewKeuanganHandler(
		createFeeComponentUC,
		updateFeeComponentUC,
		createBillingSchemeUC,
		updateBillingSchemeUC,
		assignSchemeToSantriUC,
		updateAssignmentUC,
		createInvoiceUC,
		createInvoiceBatchUC,
		cancelInvoiceUC,
		applyAdjustmentUC,
		createManualPaymentUC,
		verifyPaymentUC,
		rejectPaymentUC,
		createAccountUC,
		updateAccountUC,
		createManualJournalUC,
		cancelJournalUC,
		createPeriodUC,
		closePeriodUC,
		reopenPeriodUC,
		lockPeriodUC,
		createBillingPeriodUC,
		openBillingPeriodUC,
		closeBillingPeriodUC,
		listFeeComponentsUC,
		listBillingSchemesUC,
		getBillingSchemeUC,
		listInvoicesUC,
		getInvoiceUC,
		myInvoicesUC,
		mySummaryUC,
		listPaymentsUC,
		getPaymentUC,
		myPaymentsUC,
		listAccountsUC,
		getAccountUC,
		listJournalEntriesUC,
		getJournalEntryUC,
		getJournalEntryBySourceUC,
		listPeriodsUC,
		getActivePeriodUC,
		listAssignmentsUC,
		listBillingPeriodsUC,
		getBillingPeriodUC,
		listBillingBatchesUC,
		getBillingBatchUC,
		reportSummaryUC,
		reportOutstandingUC,
		reportLedgerUC,
		reportTrialBalanceUC,
		reportBalanceSheetUC,
		reportIncomeStatementUC,
		feeComponentRepo,
		billingSchemeRepo,
		billingPeriodRepo,
		accountRepo,
		invoiceRepo,
	)

	return &Module{
		handler:       handler,
		transactor:    transactor,
		invoiceRepo:   invoiceRepo,
		jwtAuth:       jwtAuth,
		principalLoad: principalLoad,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	keuanganHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}

func (m *Module) GetOutstandingInvoices(ctx context.Context, santriID string) (*OutstandingSummary, error) {
	invoices, err := m.invoiceRepo.FindOutstandingBySantriID(ctx, santriID)
	if err != nil {
		return nil, err
	}

	total := 0.0
	count := 0
	for _, inv := range invoices {
		if inv.HasOutstanding() {
			total += inv.RemainingAmount()
			count++
		}
	}

	return &OutstandingSummary{
		HasOutstanding:   count > 0,
		TotalOutstanding: total,
		Count:            count,
	}, nil
}

func (m *Module) HasPaidComponent(ctx context.Context, santriID, componentCode, billingPeriodID string) (bool, error) {
	return m.invoiceRepo.HasPaidComponent(ctx, santriID, componentCode, billingPeriodID)
}
