package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/kesantrian/domain/santri/constant"
	"sipon-be/internal/modules/kesantrian/domain/santri/entity"
	"sipon-be/internal/modules/kesantrian/domain/santri/repository"
	"sipon-be/internal/modules/kesantrian/domain/santri/valueobject"
	"sipon-be/internal/shared/kernel"
)

// santriColumns is the SINGLE explicit column list used by every SELECT in
// this repo — never `SELECT *`. This is a deliberate house rule to avoid
// the class of bug sipon-api hit (a column added later via ALTER TABLE
// physically reordering after `deleted_at`, silently breaking a `SELECT *`
// scan that assumed the original CREATE TABLE order).
const santriColumns = `
	id, user_id, nis, nickname, program, "option", hobby, purpose, motivation_entry, pob, dob, blood,
	address, sub_district, district, province, postal_code,
	previous_pondok_name, previous_pondok_address, previous_pondok_div, previous_pondok_time,
	nik, no_kk, nisn, no_kip, no_kks, no_pkh,
	workplace, department,
	home_status,
	father, father_pn, father_nik, father_job, father_graduate, father_income,
	mother, mother_pn, mother_nik, mother_job, mother_graduate, mother_income,
	guardian_relationship, guardian, guardian_pn, guardian_nik, guardian_job, guardian_graduate, guardian_income,
	created_at, updated_at, deleted_at,
	status, status_changed_by, status_changed_at, status_notes
`

var santriSortColumns = map[string]string{
	"created_at": "created_at",
	"nickname":   "nickname",
	"nis":        "nis",
}

type PostgresSantriRepository struct {
	db *sql.DB
}

func NewPostgresSantriRepository(db *sql.DB) *PostgresSantriRepository {
	return &PostgresSantriRepository{db: db}
}

func (r *PostgresSantriRepository) Save(ctx context.Context, s *entity.Santri) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO santri (` + santriColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,
		$13,$14,$15,$16,$17,
		$18,$19,$20,$21,
		$22,$23,$24,$25,$26,$27,
		$28,$29,
		$30,
		$31,$32,$33,$34,$35,$36,
		$37,$38,$39,$40,$41,$42,
		$43,$44,$45,$46,$47,$48,$49,
		$50,$51,$52,
		$53,$54,$55,$56
	)`

	_, err := execer.ExecContext(ctx, query, r.args(s)...)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeSantriDuplicate, err)
		}
		return kernel.Wrap(constant.CodeSantriPersistenceFailed, fmt.Errorf("save santri: %w", err))
	}
	return nil
}

func (r *PostgresSantriRepository) Update(ctx context.Context, s *entity.Santri) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE santri SET
		user_id=$1, nis=$2, nickname=$3, program=$4, "option"=$5, hobby=$6, purpose=$7, motivation_entry=$8, pob=$9, dob=$10, blood=$11,
		address=$12, sub_district=$13, district=$14, province=$15, postal_code=$16,
		previous_pondok_name=$17, previous_pondok_address=$18, previous_pondok_div=$19, previous_pondok_time=$20,
		nik=$21, no_kk=$22, nisn=$23, no_kip=$24, no_kks=$25, no_pkh=$26,
		workplace=$27, department=$28,
		home_status=$29,
		father=$30, father_pn=$31, father_nik=$32, father_job=$33, father_graduate=$34, father_income=$35,
		mother=$36, mother_pn=$37, mother_nik=$38, mother_job=$39, mother_graduate=$40, mother_income=$41,
		guardian_relationship=$42, guardian=$43, guardian_pn=$44, guardian_nik=$45, guardian_job=$46, guardian_graduate=$47, guardian_income=$48,
		updated_at=$49, deleted_at=$50,
		status=$51, status_changed_by=$52, status_changed_at=$53, status_notes=$54
		WHERE id=$55 AND deleted_at IS NULL`

	args := []interface{}{
		s.UserID, nullStr(nisString(s.NIS)), nullStr(s.Nickname), nullStr(s.Program), nullStr(s.Option), nullStr(s.Hobby), nullStr(s.Purpose), nullStr(s.MotivationEntry), nullStr(s.POB), nullTimeVal(s.DOB), nullStr(s.Blood),
		nullStr(s.Address), nullStr(s.SubDistrict), nullStr(s.District), nullStr(s.Province), nullStr(s.PostalCode),
		nullStr(s.PreviousPondokName), nullStr(s.PreviousPondokAddress), nullStr(s.PreviousPondokDiv), nullStr(s.PreviousPondokTime),
		nullStr(s.NIK), nullStr(s.NoKK), nullStr(s.NISN), nullStr(s.NoKIP), nullStr(s.NoKKS), nullStr(s.NoPKH),
		nullStr(s.Workplace), nullStr(s.Department),
		nullStr(s.HomeStatus),
		nullStr(s.Father), nullStr(s.FatherPN), nullStr(s.FatherNIK), nullStr(s.FatherJob), nullStr(s.FatherGraduate), nullStr(s.FatherIncome),
		nullStr(s.Mother), nullStr(s.MotherPN), nullStr(s.MotherNIK), nullStr(s.MotherJob), nullStr(s.MotherGraduate), nullStr(s.MotherIncome),
		nullStr(s.GuardianRelationship), nullStr(s.Guardian), nullStr(s.GuardianPN), nullStr(s.GuardianNIK), nullStr(s.GuardianJob), nullStr(s.GuardianGraduate), nullStr(s.GuardianIncome),
		s.UpdatedAt, nullTimeVal(s.DeletedAt),
		string(s.Status), nullStr(s.StatusChangedBy), nullTimeVal(s.StatusChangedAt), nullStr(s.StatusNotes),
		s.ID,
	}

	res, err := execer.ExecContext(ctx, query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeSantriDuplicate, err)
		}
		return kernel.Wrap(constant.CodeSantriPersistenceFailed, fmt.Errorf("update santri: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeSantriNotFound)
	}
	return nil
}

func (r *PostgresSantriRepository) FindByID(ctx context.Context, id string) (*entity.Santri, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+santriColumns+` FROM santri WHERE id=$1 AND deleted_at IS NULL`, id)
	return r.scan(row)
}

func (r *PostgresSantriRepository) FindByUserID(ctx context.Context, userID string) (*entity.Santri, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+santriColumns+` FROM santri WHERE user_id=$1 AND deleted_at IS NULL`, userID)
	return r.scan(row)
}

func (r *PostgresSantriRepository) FindByNIS(ctx context.Context, nis string) (*entity.Santri, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+santriColumns+` FROM santri WHERE nis=$1 AND deleted_at IS NULL`, nis)
	return r.scan(row)
}

func (r *PostgresSantriRepository) List(ctx context.Context, q repository.SantriListQuery) (*repository.SantriListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.NIS != nil && *q.NIS != "" {
		where += fmt.Sprintf(` AND nis ILIKE $%d`, argIdx)
		args = append(args, "%"+*q.NIS+"%")
		argIdx++
	}

	sortCol, ok := santriSortColumns[q.SortBy]
	if !ok {
		sortCol = "created_at"
	}
	sortDir := "DESC"
	if q.SortType == "asc" {
		sortDir = "ASC"
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM santri `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeSantriQueryFailed, fmt.Errorf("count santri: %w", err))
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM santri %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		santriColumns, where, sortCol, sortDir, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeSantriQueryFailed, fmt.Errorf("list santri: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.Santri, 0)
	for rows.Next() {
		s, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeSantriQueryFailed, fmt.Errorf("iterate santri rows: %w", err))
	}

	return &repository.SantriListResult{Items: items, Total: total}, nil
}

func (r *PostgresSantriRepository) args(s *entity.Santri) []interface{} {
	return []interface{}{
		s.ID, s.UserID, nullStr(nisString(s.NIS)), nullStr(s.Nickname), nullStr(s.Program), nullStr(s.Option), nullStr(s.Hobby), nullStr(s.Purpose), nullStr(s.MotivationEntry), nullStr(s.POB), nullTimeVal(s.DOB), nullStr(s.Blood),
		nullStr(s.Address), nullStr(s.SubDistrict), nullStr(s.District), nullStr(s.Province), nullStr(s.PostalCode),
		nullStr(s.PreviousPondokName), nullStr(s.PreviousPondokAddress), nullStr(s.PreviousPondokDiv), nullStr(s.PreviousPondokTime),
		nullStr(s.NIK), nullStr(s.NoKK), nullStr(s.NISN), nullStr(s.NoKIP), nullStr(s.NoKKS), nullStr(s.NoPKH),
		nullStr(s.Workplace), nullStr(s.Department),
		nullStr(s.HomeStatus),
		nullStr(s.Father), nullStr(s.FatherPN), nullStr(s.FatherNIK), nullStr(s.FatherJob), nullStr(s.FatherGraduate), nullStr(s.FatherIncome),
		nullStr(s.Mother), nullStr(s.MotherPN), nullStr(s.MotherNIK), nullStr(s.MotherJob), nullStr(s.MotherGraduate), nullStr(s.MotherIncome),
		nullStr(s.GuardianRelationship), nullStr(s.Guardian), nullStr(s.GuardianPN), nullStr(s.GuardianNIK), nullStr(s.GuardianJob), nullStr(s.GuardianGraduate), nullStr(s.GuardianIncome),
		s.CreatedAt, s.UpdatedAt, nullTimeVal(s.DeletedAt),
		string(s.Status), nullStr(s.StatusChangedBy), nullTimeVal(s.StatusChangedAt), nullStr(s.StatusNotes),
	}
}

func (r *PostgresSantriRepository) scan(sc scanner) (*entity.Santri, error) {
	var (
		id, userID                                                                                    string
		nis                                                                                           sql.NullString
		nickname, program, option, hobby, purpose, motivationEntry, pob, blood                        sql.NullString
		dob                                                                                           sql.NullTime
		address, subDistrict, district, province, postalCode                                          sql.NullString
		prevName, prevAddress, prevDiv, prevTime                                                      sql.NullString
		nik, noKK, nisn, noKIP, noKKS, noPKH                                                          sql.NullString
		workplace, department                                                                         sql.NullString
		homeStatus                                                                                    sql.NullString
		father, fatherPN, fatherNIK, fatherJob, fatherGraduate, fatherIncome                          sql.NullString
		mother, motherPN, motherNIK, motherJob, motherGraduate, motherIncome                          sql.NullString
		guardianRel, guardian, guardianPN, guardianNIK, guardianJob, guardianGraduate, guardianIncome sql.NullString
		createdAt, updatedAt                                                                          time.Time
		deletedAt                                                                                     sql.NullTime
		status, statusChangedBy                                                                       sql.NullString
		statusChangedAt                                                                               sql.NullTime
		statusNotes                                                                                   sql.NullString
	)

	err := sc.Scan(
		&id, &userID, &nis, &nickname, &program, &option, &hobby, &purpose, &motivationEntry, &pob, &dob, &blood,
		&address, &subDistrict, &district, &province, &postalCode,
		&prevName, &prevAddress, &prevDiv, &prevTime,
		&nik, &noKK, &nisn, &noKIP, &noKKS, &noPKH,
		&workplace, &department,
		&homeStatus,
		&father, &fatherPN, &fatherNIK, &fatherJob, &fatherGraduate, &fatherIncome,
		&mother, &motherPN, &motherNIK, &motherJob, &motherGraduate, &motherIncome,
		&guardianRel, &guardian, &guardianPN, &guardianNIK, &guardianJob, &guardianGraduate, &guardianIncome,
		&createdAt, &updatedAt, &deletedAt,
		&status, &statusChangedBy, &statusChangedAt, &statusNotes,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeSantriNotFound)
		}
		return nil, kernel.Wrap(constant.CodeSantriQueryFailed, fmt.Errorf("scan santri: %w", err))
	}

	s := &entity.Santri{
		ID:     id,
		UserID: userID,

		Nickname:        strFromNull(nickname),
		Program:         strFromNull(program),
		Option:          strFromNull(option),
		Hobby:           strFromNull(hobby),
		Purpose:         strFromNull(purpose),
		MotivationEntry: strFromNull(motivationEntry),
		POB:             strFromNull(pob),
		DOB:             timeFromNull(dob),
		Blood:           strFromNull(blood),

		Address:     strFromNull(address),
		SubDistrict: strFromNull(subDistrict),
		District:    strFromNull(district),
		Province:    strFromNull(province),
		PostalCode:  strFromNull(postalCode),

		PreviousPondokName:    strFromNull(prevName),
		PreviousPondokAddress: strFromNull(prevAddress),
		PreviousPondokDiv:     strFromNull(prevDiv),
		PreviousPondokTime:    strFromNull(prevTime),

		NIK:   strFromNull(nik),
		NoKK:  strFromNull(noKK),
		NISN:  strFromNull(nisn),
		NoKIP: strFromNull(noKIP),
		NoKKS: strFromNull(noKKS),
		NoPKH: strFromNull(noPKH),

		Workplace:  strFromNull(workplace),
		Department: strFromNull(department),

		HomeStatus: strFromNull(homeStatus),

		Father:         strFromNull(father),
		FatherPN:       strFromNull(fatherPN),
		FatherNIK:      strFromNull(fatherNIK),
		FatherJob:      strFromNull(fatherJob),
		FatherGraduate: strFromNull(fatherGraduate),
		FatherIncome:   strFromNull(fatherIncome),

		Mother:         strFromNull(mother),
		MotherPN:       strFromNull(motherPN),
		MotherNIK:      strFromNull(motherNIK),
		MotherJob:      strFromNull(motherJob),
		MotherGraduate: strFromNull(motherGraduate),
		MotherIncome:   strFromNull(motherIncome),

		GuardianRelationship: strFromNull(guardianRel),
		Guardian:             strFromNull(guardian),
		GuardianPN:           strFromNull(guardianPN),
		GuardianNIK:          strFromNull(guardianNIK),
		GuardianJob:          strFromNull(guardianJob),
		GuardianGraduate:     strFromNull(guardianGraduate),
		GuardianIncome:       strFromNull(guardianIncome),

		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		DeletedAt:      timeFromNull(deletedAt),
		Status:         constant.SantriStatus(blankToSantri(status)),
		StatusChangedBy: strFromNull(statusChangedBy),
		StatusChangedAt: timeFromNull(statusChangedAt),
		StatusNotes:     strFromNull(statusNotes),
	}

	if nis.Valid && nis.String != "" {
		n, err := valueobject.NewNIS(nis.String)
		if err == nil {
			s.NIS = &n
		}
	}

	return s, nil
}

func blankToSantri(s sql.NullString) string {
	if !s.Valid || s.String == "" {
		return "SANTRI"
	}
	return s.String
}

func (r *PostgresSantriRepository) FindMaxSequence(ctx context.Context, prefix string) (int, error) {
	execer := execerFromContext(ctx, r.db)

	if _, err := execer.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, prefix); err != nil {
		return 0, kernel.Wrap(constant.CodeSantriQueryFailed, fmt.Errorf("advisory lock: %w", err))
	}

	var maxSeq sql.NullInt64
	row := execer.QueryRowContext(ctx,
		`SELECT MAX(CAST(SUBSTRING(nis FROM 8) AS INTEGER)) FROM santri WHERE nis LIKE $1 AND deleted_at IS NULL`,
		prefix+"%",
	)
	if err := row.Scan(&maxSeq); err != nil {
		return 0, kernel.Wrap(constant.CodeSantriQueryFailed, fmt.Errorf("find max sequence: %w", err))
	}

	if !maxSeq.Valid {
		return 0, nil
	}
	return int(maxSeq.Int64), nil
}

func nisString(nis *valueobject.NIS) *string {
	if nis == nil {
		return nil
	}
	v := nis.String()
	return &v
}
