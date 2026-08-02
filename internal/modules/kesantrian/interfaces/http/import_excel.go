package http

import (
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"sipon-be/internal/modules/kesantrian/application/command"
)

// importColumn pairs a spreadsheet header (matched case/space-insensitively)
// with a setter that writes the cell's value onto a row being built. Order
// here is also the column order used when generating the template — but
// parsing itself looks columns up BY HEADER NAME, so an admin reordering
// columns in their own copy still works.
type importColumn struct {
	header string
	set    func(row *command.ImportSantriRow, value string) error
}

var importColumns = []importColumn{
	{"NIS", func(r *command.ImportSantriRow, v string) error { r.NIS = strings.TrimSpace(v); return nil }},
	{"Nama Lengkap", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Fullname })},
	{"Nama Panggilan", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Nickname })},
	{"Program", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Program })},
	{"Golongan Darah", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Blood })},
	{"Tempat Lahir", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.POB })},
	{"Tanggal Lahir (YYYY-MM-DD)", setDate(func(r *command.ImportSantriRow) **time.Time { return &r.Profile.DOB })},
	{"Hobi", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Hobby })},
	{"Tujuan/Cita-cita", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Purpose })},
	{"Motivasi Masuk", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.MotivationEntry })},
	{"Alamat", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Address })},
	{"Kecamatan", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.SubDistrict })},
	{"Kabupaten/Kota", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.District })},
	{"Provinsi", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Province })},
	{"Kode Pos", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.PostalCode })},
	{"Nama Pondok Sebelumnya", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.PreviousPondokName })},
	{"Alamat Pondok Sebelumnya", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.PreviousPondokAddress })},
	{"Jenjang/Divisi Pondok", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.PreviousPondokDiv })},
	{"Lama Belajar Di Pondok Sebelumnya", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.PreviousPondokTime })},
	{"NIK", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.NIK })},
	{"No. Kartu Keluarga", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.NoKK })},
	{"NISN", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.NISN })},
	{"No. KIP", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.NoKIP })},
	{"No. KKS", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.NoKKS })},
	{"No. PKH", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.NoPKH })},
	{"Tempat Bekerja", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Workplace })},
	{"Departemen/Bagian", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Department })},
	{"Status Rumah", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.HomeStatus })},
	{"Nama Ayah", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Father })},
	{"No. Telepon Ayah", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.FatherPN })},
	{"NIK Ayah", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.FatherNIK })},
	{"Pekerjaan Ayah", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.FatherJob })},
	{"Pendidikan Terakhir Ayah", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.FatherGraduate })},
	{"Penghasilan Ayah", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.FatherIncome })},
	{"Nama Ibu", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Mother })},
	{"No. Telepon Ibu", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.MotherPN })},
	{"NIK Ibu", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.MotherNIK })},
	{"Pekerjaan Ibu", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.MotherJob })},
	{"Pendidikan Terakhir Ibu", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.MotherGraduate })},
	{"Penghasilan Ibu", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.MotherIncome })},
	{"Hubungan Wali", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.GuardianRelationship })},
	{"Nama Wali", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.Guardian })},
	{"No. Telepon Wali", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.GuardianPN })},
	{"NIK Wali", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.GuardianNIK })},
	{"Pekerjaan Wali", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.GuardianJob })},
	{"Pendidikan Terakhir Wali", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.GuardianGraduate })},
	{"Penghasilan Wali", setStr(func(r *command.ImportSantriRow) **string { return &r.Profile.GuardianIncome })},
}

// setStr builds a column setter for a **string field: empty cells are left
// nil (so applySantriUpdate's partial-update semantics skip them) instead of
// being set to a pointer-to-empty-string.
func setStr(field func(r *command.ImportSantriRow) **string) func(r *command.ImportSantriRow, v string) error {
	return func(r *command.ImportSantriRow, v string) error {
		v = strings.TrimSpace(v)
		if v == "" {
			return nil
		}
		*field(r) = &v
		return nil
	}
}

func setDate(field func(r *command.ImportSantriRow) **time.Time) func(r *command.ImportSantriRow, v string) error {
	return func(r *command.ImportSantriRow, v string) error {
		v = strings.TrimSpace(v)
		if v == "" {
			return nil
		}
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return fmt.Errorf("format tanggal lahir harus YYYY-MM-DD, dapat %q", v)
		}
		*field(r) = &t
		return nil
	}
}

func normalizeHeader(h string) string {
	return strings.ToLower(strings.TrimSpace(h))
}

// parseSantriImportExcel reads the first sheet of an uploaded .xlsx file,
// matches its header row against importColumns (case/space-insensitive), and
// builds one ImportSantriRow per subsequent non-empty row.
func parseSantriImportExcel(file multipart.File) ([]command.ImportSantriRow, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("file bukan spreadsheet .xlsx yang valid: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("file tidak memiliki sheet")
	}

	allRows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("gagal membaca sheet: %w", err)
	}
	if len(allRows) < 2 {
		return nil, fmt.Errorf("file kosong — isi minimal satu baris data di bawah header")
	}

	header := allRows[0]
	colIndexToSetter := make(map[int]func(row *command.ImportSantriRow, value string) error, len(importColumns))
	for i, cell := range header {
		norm := normalizeHeader(cell)
		for _, col := range importColumns {
			if normalizeHeader(col.header) == norm {
				colIndexToSetter[i] = col.set
				break
			}
		}
	}
	if _, ok := findNISColumn(header); !ok {
		return nil, fmt.Errorf("kolom NIS tidak ditemukan — gunakan template resmi")
	}

	rows := make([]command.ImportSantriRow, 0, len(allRows)-1)
	for i, rawRow := range allRows[1:] {
		rowNumber := i + 2 // +1 for 0-index, +1 for header row
		if isBlankRow(rawRow) {
			continue
		}

		row := command.ImportSantriRow{RowNumber: rowNumber}
		for colIdx, setter := range colIndexToSetter {
			if colIdx >= len(rawRow) {
				continue
			}
			if err := setter(&row, rawRow[colIdx]); err != nil {
				return nil, fmt.Errorf("baris %d: %w", rowNumber, err)
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func findNISColumn(header []string) (int, bool) {
	for i, cell := range header {
		if normalizeHeader(cell) == "nis" {
			return i, true
		}
	}
	return 0, false
}

func isBlankRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// buildSantriImportTemplate generates a fresh .xlsx with just the header
// row, matching importColumns exactly, for the admin to fill in and
// re-upload.
func buildSantriImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := f.GetSheetName(0)
	for i, col := range importColumns {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheet, cell, col.header); err != nil {
			return nil, err
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
