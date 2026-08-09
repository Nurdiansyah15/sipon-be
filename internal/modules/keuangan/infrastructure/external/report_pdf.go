package external

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// ReportData adalah data mentah yang dibutuhkan generator PDF laporan.
// Semua field opsional — generator hanya menampilkan yang terisi.

type ReportTableRow struct {
	Cells []string
}

func newReportPDF(orientation string, title string, subtitle string) *gofpdf.Fpdf {
	pdf := gofpdf.New(orientation, "mm", "A4", "")
	pdf.SetMargins(10, 12, 10)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, "PONDOK PESANTREN SIPON", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, title, "", 1, "C", false, 0, "")
	if subtitle != "" {
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(0, 5, subtitle, "", 1, "C", false, 0, "")
	}
	pdf.Ln(3)
	return pdf
}

// reportTable menampilkan tabel dengan kolom lebar mm. header non-nil berarti
// baris pertama adalah header. Kolom kanan-rata bila rightAlign[i] = true.
func reportTable(pdf *gofpdf.Fpdf, widths []float64, header []string, rows [][]string, rightAlign []bool) {
	drawRow := func(cells []string, isHeader bool) {
		for i, cell := range cells {
			if i >= len(widths) {
				break
			}
			align := "L"
			if isHeader {
				align = "C"
			} else if i < len(rightAlign) && rightAlign[i] {
				align = "R"
			}
			if isHeader {
				pdf.SetFont("Helvetica", "B", 8)
			} else {
				pdf.SetFont("Helvetica", "", 8)
			}
			pdf.CellFormat(widths[i], 5, cell, "1", 0, align, false, 0, "")
		}
		pdf.Ln(5)
	}

	if len(header) > 0 {
		drawRow(header, true)
	}
	for _, row := range rows {
		drawRow(row, false)
	}
}

func reportGeneratedBy(pdf *gofpdf.Fpdf) {
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(0, 5, fmt.Sprintf("Dicetak: %s", time.Now().Format("02/01/2006 15:04")), "", 1, "R", false, 0, "")
}

func reportOutput(pdf *gofpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// ── Laporan-laporan ─────────────────────────────────────────────────────────

type SummaryReportData struct {
	Title      string
	Rows       [][]string
	TotalTagihan   float64
	TotalTerbayar  float64
	TotalTunggakan float64
}

func GenerateSummaryPDF(data SummaryReportData) ([]byte, error) {
	pdf := newReportPDF("P", "LAPORAN REKAP TAGIHAN", data.Title)
	reportTable(pdf, []float64{60, 32, 32, 32, 25, 25, 25},
		[]string{"Periode", "Tagihan", "Terbayar", "Tunggakan", "Invoice", "Lunas", "Belum"},
		data.Rows, []bool{false, true, true, true, true, true, true})
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(60, 5, "TOTAL", "", 0, "L", false, 0, "")
	pdf.CellFormat(32, 5, FormatRupiah(data.TotalTagihan), "", 1, "R", false, 0, "")
	reportGeneratedBy(pdf)
	return reportOutput(pdf)
}

type OutstandingReportData struct {
	Title string
	Rows  [][]string
}

func GenerateOutstandingPDF(data OutstandingReportData) ([]byte, error) {
	pdf := newReportPDF("P", "LAPORAN TUNGGAKAN PER SANTRI", data.Title)
	reportTable(pdf, []float64{50, 55, 45},
		[]string{"No.", "Santri", "Total Tunggakan"},
		data.Rows, []bool{true, false, true})
	reportGeneratedBy(pdf)
	return reportOutput(pdf)
}

type LedgerReportData struct {
	Title string
	Rows  [][]string
}

func GenerateLedgerPDF(data LedgerReportData) ([]byte, error) {
	pdf := newReportPDF("L", "BUKU BESAR", data.Title)
	reportTable(pdf, []float64{28, 45, 70, 35, 35, 35},
		[]string{"Tanggal", "No. Jurnal", "Keterangan", "Debit", "Kredit", "Saldo"},
		data.Rows, []bool{false, false, false, true, true, true})
	reportGeneratedBy(pdf)
	return reportOutput(pdf)
}

type TrialBalanceReportData struct {
	Title      string
	Rows       [][]string
	TotalDebit float64
	TotalCredit float64
}

func GenerateTrialBalancePDF(data TrialBalanceReportData) ([]byte, error) {
	pdf := newReportPDF("P", "NERACA SALDO", data.Title)
	reportTable(pdf, []float64{30, 70, 35, 35},
		[]string{"Kode", "Akun", "Debit", "Kredit"},
		data.Rows, []bool{false, false, true, true})
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(100, 5, "TOTAL", "", 0, "L", false, 0, "")
	pdf.CellFormat(35, 5, FormatRupiah(data.TotalDebit), "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 5, FormatRupiah(data.TotalCredit), "", 1, "R", false, 0, "")
	reportGeneratedBy(pdf)
	return reportOutput(pdf)
}

type BalanceSheetReportData struct {
	Title           string
	Assets          [][]string
	TotalAssets     float64
	Liabilities     [][]string
	TotalLiabilities float64
	Equities        [][]string
	TotalEquities   float64
}

func GenerateBalanceSheetPDF(data BalanceSheetReportData) ([]byte, error) {
	pdf := newReportPDF("P", "NERACA", data.Title)
	widths := []float64{30, 70, 40}
	rightAlign := []bool{false, false, true}

	section := func(label string, rows [][]string, total float64) {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(230, 230, 230)
		pdf.CellFormat(140, 6, label, "", 1, "L", true, 0, "")
		reportTable(pdf, widths, nil, rows, rightAlign)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.CellFormat(100, 5, fmt.Sprintf("Total %s", label), "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, FormatRupiah(total), "", 1, "R", false, 0, "")
		pdf.Ln(2)
	}

	section("ASET", data.Assets, data.TotalAssets)
	section("KEWAJIBAN", data.Liabilities, data.TotalLiabilities)
	section("EKUITAS", data.Equities, data.TotalEquities)
	reportGeneratedBy(pdf)
	return reportOutput(pdf)
}

type IncomeStatementReportData struct {
	Title      string
	Revenues   [][]string
	TotalRevenue float64
	Expenses   [][]string
	TotalExpense float64
	NetIncome  float64
}

func GenerateIncomeStatementPDF(data IncomeStatementReportData) ([]byte, error) {
	pdf := newReportPDF("P", "LAPORAN LABA RUGI", data.Title)
	widths := []float64{30, 70, 40}
	rightAlign := []bool{false, false, true}

	section := func(label string, rows [][]string, total float64) {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(230, 230, 230)
		pdf.CellFormat(140, 6, label, "", 1, "L", true, 0, "")
		reportTable(pdf, widths, nil, rows, rightAlign)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.CellFormat(100, 5, fmt.Sprintf("Total %s", label), "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, FormatRupiah(total), "", 1, "R", false, 0, "")
		pdf.Ln(2)
	}

	section("PENDAPATAN", data.Revenues, data.TotalRevenue)
	section("BEBAN", data.Expenses, data.TotalExpense)

	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(100, 6, "LABA (RUGI) BERSIH", "", 0, "L", false, 0, "")
	pdf.CellFormat(40, 6, FormatRupiah(data.NetIncome), "", 1, "R", false, 0, "")
	reportGeneratedBy(pdf)
	return reportOutput(pdf)
}
