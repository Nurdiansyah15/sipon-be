package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, h *KeuanganHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	santri := router.Group("/api/v1/web/keuangan")
	santri.Use(jwtAuth, principalLoad)
	{
		santri.GET("/invoices", h.MyInvoices)
		santri.GET("/invoices/:id", h.GetInvoice)
		santri.GET("/payments", h.MyPayments)
	}

	admin := router.Group("/api/v1/web/keuangan/admin")
	admin.Use(jwtAuth, principalLoad)
	{
		admin.GET("/components", middleware.RequirePermission("manage_keuangan"), h.ListFeeComponents)
		admin.POST("/components", middleware.RequirePermission("manage_keuangan"), h.CreateFeeComponent)
		admin.PUT("/components/:id", middleware.RequirePermission("manage_keuangan"), h.UpdateFeeComponent)
		admin.DELETE("/components/:id", middleware.RequirePermission("manage_keuangan"), h.DeleteFeeComponent)

		admin.GET("/schemes", middleware.RequirePermission("manage_keuangan"), h.ListBillingSchemes)
		admin.GET("/schemes/:id", middleware.RequirePermission("manage_keuangan"), h.GetBillingScheme)
		admin.POST("/schemes", middleware.RequirePermission("manage_keuangan"), h.CreateBillingScheme)
		admin.PUT("/schemes/:id", middleware.RequirePermission("manage_keuangan"), h.UpdateBillingScheme)
		admin.DELETE("/schemes/:id", middleware.RequirePermission("manage_keuangan"), h.DeleteBillingScheme)
		admin.POST("/schemes/:id/items", middleware.RequirePermission("manage_keuangan"), h.AddSchemeItem)
		admin.DELETE("/schemes/:id/items/:itemId", middleware.RequirePermission("manage_keuangan"), h.RemoveSchemeItem)

		admin.POST("/assignments", middleware.RequirePermission("manage_keuangan"), h.AssignSchemeToSantri)
		admin.GET("/assignments", middleware.RequirePermission("manage_keuangan"), h.ListAssignments)
		admin.PUT("/assignments/:id", middleware.RequirePermission("manage_keuangan"), h.UpdateAssignmentToSantri)

		admin.GET("/invoices", middleware.RequirePermission("manage_keuangan"), h.ListInvoices)
		admin.POST("/invoices", middleware.RequirePermission("manage_keuangan"), h.CreateInvoice)
		admin.POST("/invoices/batch", middleware.RequirePermission("manage_keuangan"), h.CreateInvoiceBatch)
		admin.GET("/invoices/:id", middleware.RequirePermission("manage_keuangan"), h.GetInvoice)
		admin.POST("/invoices/:id/cancel", middleware.RequirePermission("manage_keuangan"), h.CancelInvoice)
		admin.POST("/invoices/:id/adjustment", middleware.RequirePermission("manage_keuangan"), h.ApplyAdjustment)

		admin.GET("/payments", middleware.RequirePermission("manage_keuangan"), h.ListPayments)
		admin.GET("/payments/:id", middleware.RequirePermission("manage_keuangan"), h.GetPayment)
		admin.GET("/payments/:id/receipt", middleware.RequirePermission("manage_keuangan"), h.DownloadReceipt)
		admin.POST("/payments/manual", middleware.RequirePermission("manage_keuangan"), h.CreateManualPayment)
		admin.POST("/payments/:id/verify", middleware.RequirePermission("verify_payment"), h.VerifyPayment)
		admin.POST("/payments/:id/reject", middleware.RequirePermission("verify_payment"), h.RejectPayment)

		admin.GET("/accounts", middleware.RequirePermission("manage_accounts"), h.ListAccounts)
		admin.GET("/accounts/:id", middleware.RequirePermission("manage_accounts"), h.GetAccount)
		admin.POST("/accounts", middleware.RequirePermission("manage_accounts"), h.CreateAccount)
		admin.PUT("/accounts/:id", middleware.RequirePermission("manage_accounts"), h.UpdateAccount)
		admin.DELETE("/accounts/:id", middleware.RequirePermission("manage_accounts"), h.DeleteAccount)

		admin.GET("/journal-entries", middleware.RequirePermission("manage_journal"), h.ListJournalEntries)
		admin.GET("/journal-entries/by-source", middleware.RequirePermission("manage_journal"), h.GetJournalEntryBySource)
		admin.GET("/journal-entries/:id", middleware.RequirePermission("manage_journal"), h.GetJournalEntry)
		admin.POST("/journal-entries", middleware.RequirePermission("manage_journal"), h.CreateJournalEntry)
		admin.POST("/journal-entries/:id/cancel", middleware.RequirePermission("manage_journal"), h.CancelJournalEntry)

		admin.GET("/periods", middleware.RequirePermission("close_period"), h.ListPeriods)
		admin.GET("/periods/active", middleware.RequirePermission("close_period"), h.GetActivePeriod)
		admin.POST("/periods", middleware.RequirePermission("close_period"), h.CreatePeriod)
		admin.POST("/periods/:id/close", middleware.RequirePermission("close_period"), h.ClosePeriod)
		admin.POST("/periods/:id/reopen", middleware.RequirePermission("close_period"), h.ReopenPeriod)
		admin.POST("/periods/:id/lock", middleware.RequirePermission("close_period"), h.LockPeriod)

		admin.GET("/billing-periods", middleware.RequirePermission("manage_keuangan"), h.ListBillingPeriods)
		admin.GET("/billing-periods/:id", middleware.RequirePermission("manage_keuangan"), h.GetBillingPeriod)
		admin.POST("/billing-periods", middleware.RequirePermission("manage_keuangan"), h.CreateBillingPeriod)
		admin.POST("/billing-periods/:id/open", middleware.RequirePermission("manage_keuangan"), h.OpenBillingPeriod)
		admin.POST("/billing-periods/:id/close", middleware.RequirePermission("manage_keuangan"), h.CloseBillingPeriod)

		admin.GET("/billing-batches", middleware.RequirePermission("manage_keuangan"), h.ListBillingBatches)
		admin.GET("/billing-batches/:id", middleware.RequirePermission("manage_keuangan"), h.GetBillingBatch)

		admin.GET("/reports/summary", middleware.RequirePermission("view_keuangan_reports"), h.ReportSummary)
		admin.GET("/reports/outstanding", middleware.RequirePermission("view_keuangan_reports"), h.ReportOutstanding)
		admin.GET("/reports/ledger", middleware.RequirePermission("view_keuangan_reports"), h.ReportLedger)
		admin.GET("/reports/trial-balance", middleware.RequirePermission("view_keuangan_reports"), h.ReportTrialBalance)
		admin.GET("/reports/balance-sheet", middleware.RequirePermission("view_keuangan_reports"), h.ReportBalanceSheet)
		admin.GET("/reports/income-statement", middleware.RequirePermission("view_keuangan_reports"), h.ReportIncomeStatement)

		admin.GET("/reports/summary/pdf", middleware.RequirePermission("view_keuangan_reports"), h.ReportSummaryPDF)
		admin.GET("/reports/outstanding/pdf", middleware.RequirePermission("view_keuangan_reports"), h.ReportOutstandingPDF)
		admin.GET("/reports/ledger/pdf", middleware.RequirePermission("view_keuangan_reports"), h.ReportLedgerPDF)
		admin.GET("/reports/trial-balance/pdf", middleware.RequirePermission("view_keuangan_reports"), h.ReportTrialBalancePDF)
		admin.GET("/reports/balance-sheet/pdf", middleware.RequirePermission("view_keuangan_reports"), h.ReportBalanceSheetPDF)
		admin.GET("/reports/income-statement/pdf", middleware.RequirePermission("view_keuangan_reports"), h.ReportIncomeStatementPDF)
	}
}
