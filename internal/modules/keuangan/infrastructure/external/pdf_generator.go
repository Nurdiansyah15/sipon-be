package external

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type ReceiptData struct {
	ReceiptNumber   string
	PaymentDate     string
	SantriName      string
	NIS             string
	InvoiceNumber   string
	FeeComponent    string
	Periode         string
	TahunAjaran     string
	Amount          float64
	PaymentMethod   string
	ReferenceNumber string
	VerifiedBy      string
}

func GenerateReceiptPDF(data ReceiptData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A5", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, "PONDOK PESANTREN SIPON", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 5, "Jl. Pesantren No. 123", "", 1, "C", false, 0, "")
	pdf.Ln(3)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "KWITANSI PEMBAYARAN", "", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "", 10)
	addRow(pdf, "No. Kwitansi", data.ReceiptNumber)
	addRow(pdf, "Tanggal", data.PaymentDate)
	pdf.Ln(2)

	addRow(pdf, "Nama Santri", data.SantriName)
	if data.NIS != "" {
		addRow(pdf, "NIS", data.NIS)
	}
	pdf.Ln(2)

	addRow(pdf, "No. Invoice", data.InvoiceNumber)
	addRow(pdf, "Komponen", data.FeeComponent)
	addRow(pdf, "Periode", data.Periode)
	addRow(pdf, "Tahun Ajaran", data.TahunAjaran)
	pdf.Ln(2)

	addRow(pdf, "Jumlah", formatRupiah(data.Amount))
	addRow(pdf, "Metode", data.PaymentMethod)
	if data.ReferenceNumber != "" {
		addRow(pdf, "Referensi", data.ReferenceNumber)
	}
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Dicetak: %s", time.Now().Format("02/01/2006 15:04")), "", 1, "R", false, 0, "")
	pdf.Ln(10)
	pdf.CellFormat(0, 5, "Bendahara,", "", 1, "R", false, 0, "")
	pdf.Ln(15)
	if data.VerifiedBy != "" {
		pdf.CellFormat(0, 5, data.VerifiedBy, "", 1, "R", false, 0, "")
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func addRow(pdf *gofpdf.Fpdf, label, value string) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(40, 6, label, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(5, 6, ":", "", 0, "C", false, 0, "")
	pdf.CellFormat(0, 6, value, "", 1, "L", false, 0, "")
}

func formatRupiah(amount float64) string {
	return fmt.Sprintf("Rp %s", commaFormat(int64(amount)))
}

func commaFormat(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return commaFormat(n/1000) + "." + fmt.Sprintf("%03d", n%1000)
}
