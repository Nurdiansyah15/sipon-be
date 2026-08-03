package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pconstant "sipon-be/internal/modules/psb/domain/pendaftar/constant"
	pentity "sipon-be/internal/modules/psb/domain/pendaftar/entity"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	"sipon-be/internal/shared/kernel"
)

const pendaftarColumns = `
	id, user_id, psb_setting_id, gender, program,
	nickname, hobby, purpose, motivation_entry, pob, dob, blood,
	address, sub_district, district, province, postal_code,
	previous_pondok_name, previous_pondok_address, previous_pondok_div, previous_pondok_time,
	nik, no_kk, nisn, no_kip, no_kks, no_pkh,
	workplace, department,
	home_status,
	father, father_pn, father_nik, father_job, father_graduate, father_income,
	mother, mother_pn, mother_nik, mother_job, mother_graduate, mother_income,
	guardian_relationship, guardian, guardian_pn, guardian_nik, guardian_job, guardian_graduate, guardian_income,
	status, accepted_by, accepted_at, santri_id, nis,
	created_at, updated_at, deleted_at
`

type PostgresPendaftarRepository struct {
	db *sql.DB
}

func NewPostgresPendaftarRepository(db *sql.DB) *PostgresPendaftarRepository {
	return &PostgresPendaftarRepository{db: db}
}

func (r *PostgresPendaftarRepository) Save(ctx context.Context, p *pentity.Pendaftar) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO pendaftar (` + pendaftarColumns + `) VALUES (` +
		`$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,` +
		`$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,` +
		`$43,$44,$45,$46,$47,$48,$49,$50,$51,$52,$53,$54,$55,$56,$57,$58,$59)`
	_, err := execer.ExecContext(ctx, query,
		p.ID, p.UserID, p.PsbSettingID, p.Gender, nullStr(p.Program),
		nullStr(p.Nickname), nullStr(p.Hobby), nullStr(p.Purpose), nullStr(p.MotivationEntry), nullStr(p.POB), nullTimeVal(p.DOB), nullStr(p.Blood),
		nullStr(p.Address), nullStr(p.SubDistrict), nullStr(p.District), nullStr(p.Province), nullStr(p.PostalCode),
		nullStr(p.PreviousPondokName), nullStr(p.PreviousPondokAddress), nullStr(p.PreviousPondokDiv), nullStr(p.PreviousPondokTime),
		nullStr(p.NIK), nullStr(p.NoKK), nullStr(p.NISN), nullStr(p.NoKIP), nullStr(p.NoKKS), nullStr(p.NoPKH),
		nullStr(p.Workplace), nullStr(p.Department),
		nullStr(p.HomeStatus),
		nullStr(p.Father), nullStr(p.FatherPN), nullStr(p.FatherNIK), nullStr(p.FatherJob), nullStr(p.FatherGraduate), nullStr(p.FatherIncome),
		nullStr(p.Mother), nullStr(p.MotherPN), nullStr(p.MotherNIK), nullStr(p.MotherJob), nullStr(p.MotherGraduate), nullStr(p.MotherIncome),
		nullStr(p.GuardianRelationship), nullStr(p.Guardian), nullStr(p.GuardianPN), nullStr(p.GuardianNIK), nullStr(p.GuardianJob), nullStr(p.GuardianGraduate), nullStr(p.GuardianIncome),
		string(p.Status), nullStr(p.AcceptedBy), nullTimeVal(p.AcceptedAt), nullStr(p.SantriID), nullStr(p.NIS),
		p.CreatedAt, p.UpdatedAt, nullTimeVal(p.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(pconstant.CodePendaftarDuplicate, err)
		}
		return kernel.Wrap(pconstant.CodePendaftarPersistenceFailed, fmt.Errorf("save pendaftar: %w", err))
	}
	return nil
}

func (r *PostgresPendaftarRepository) Update(ctx context.Context, p *pentity.Pendaftar) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE pendaftar SET ` +
		`user_id=$1, psb_setting_id=$2, gender=$3, program=$4,` +
		`nickname=$5, hobby=$6, purpose=$7, motivation_entry=$8, pob=$9, dob=$10, blood=$11,` +
		`address=$12, sub_district=$13, district=$14, province=$15, postal_code=$16,` +
		`previous_pondok_name=$17, previous_pondok_address=$18, previous_pondok_div=$19, previous_pondok_time=$20,` +
		`nik=$21, no_kk=$22, nisn=$23, no_kip=$24, no_kks=$25, no_pkh=$26,` +
		`workplace=$27, department=$28,` +
		`home_status=$29,` +
		`father=$30, father_pn=$31, father_nik=$32, father_job=$33, father_graduate=$34, father_income=$35,` +
		`mother=$36, mother_pn=$37, mother_nik=$38, mother_job=$39, mother_graduate=$40, mother_income=$41,` +
		`guardian_relationship=$42, guardian=$43, guardian_pn=$44, guardian_nik=$45, guardian_job=$46, guardian_graduate=$47, guardian_income=$48,` +
		`status=$49, accepted_by=$50, accepted_at=$51, santri_id=$52, nis=$53,` +
		`updated_at=$54, deleted_at=$55 WHERE id=$56 AND deleted_at IS NULL`
	res, err := execer.ExecContext(ctx, query,
		p.UserID, p.PsbSettingID, p.Gender, nullStr(p.Program),
		nullStr(p.Nickname), nullStr(p.Hobby), nullStr(p.Purpose), nullStr(p.MotivationEntry), nullStr(p.POB), nullTimeVal(p.DOB), nullStr(p.Blood),
		nullStr(p.Address), nullStr(p.SubDistrict), nullStr(p.District), nullStr(p.Province), nullStr(p.PostalCode),
		nullStr(p.PreviousPondokName), nullStr(p.PreviousPondokAddress), nullStr(p.PreviousPondokDiv), nullStr(p.PreviousPondokTime),
		nullStr(p.NIK), nullStr(p.NoKK), nullStr(p.NISN), nullStr(p.NoKIP), nullStr(p.NoKKS), nullStr(p.NoPKH),
		nullStr(p.Workplace), nullStr(p.Department), nullStr(p.HomeStatus),
		nullStr(p.Father), nullStr(p.FatherPN), nullStr(p.FatherNIK), nullStr(p.FatherJob), nullStr(p.FatherGraduate), nullStr(p.FatherIncome),
		nullStr(p.Mother), nullStr(p.MotherPN), nullStr(p.MotherNIK), nullStr(p.MotherJob), nullStr(p.MotherGraduate), nullStr(p.MotherIncome),
		nullStr(p.GuardianRelationship), nullStr(p.Guardian), nullStr(p.GuardianPN), nullStr(p.GuardianNIK), nullStr(p.GuardianJob), nullStr(p.GuardianGraduate), nullStr(p.GuardianIncome),
		string(p.Status), nullStr(p.AcceptedBy), nullTimeVal(p.AcceptedAt), nullStr(p.SantriID), nullStr(p.NIS),
		p.UpdatedAt, nullTimeVal(p.DeletedAt), p.ID,
	)
	if err != nil {
		return kernel.Wrap(pconstant.CodePendaftarPersistenceFailed, fmt.Errorf("update pendaftar: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(pconstant.CodePendaftarNotFound)
	}
	return nil
}

func (r *PostgresPendaftarRepository) FindByID(ctx context.Context, id string) (*pentity.Pendaftar, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+pendaftarColumns+` FROM pendaftar WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanPendaftar(row)
}

func (r *PostgresPendaftarRepository) FindByUserIDAndSetting(ctx context.Context, userID, psbSettingID string) (*pentity.Pendaftar, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+pendaftarColumns+` FROM pendaftar WHERE user_id=$1 AND psb_setting_id=$2 AND deleted_at IS NULL`, userID, psbSettingID)
	return scanPendaftar(row)
}

func (r *PostgresPendaftarRepository) CountBySettingAndProgram(ctx context.Context, psbSettingID, program string) (int64, error) {
	execer := execerFromContext(ctx, r.db)
	var count int64
	err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM pendaftar WHERE psb_setting_id=$1 AND program=$2 AND status='diterima' AND deleted_at IS NULL`, psbSettingID, program).Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func (r *PostgresPendaftarRepository) List(ctx context.Context, q prepo.PendaftarListQuery) (*prepo.PendaftarListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.PsbSettingID != "" {
		where += fmt.Sprintf(` AND psb_setting_id=$%d`, argIdx)
		args = append(args, q.PsbSettingID)
		argIdx++
	}
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM pendaftar `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.Wrap(pconstant.CodePendaftarQueryFailed, fmt.Errorf("count pendaftar: %w", err))
	}

	limit := q.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (q.Page - 1) * q.Limit
	query := fmt.Sprintf(`SELECT %s FROM pendaftar %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		pendaftarColumns, where, argIdx, argIdx+1)
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(pconstant.CodePendaftarQueryFailed, fmt.Errorf("list pendaftar: %w", err))
	}
	defer rows.Close()

	items := make([]*pentity.Pendaftar, 0)
	for rows.Next() {
		p, err := scanPendaftar(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}

	return &prepo.PendaftarListResult{Items: items, Total: total}, rows.Err()
}

func (r *PostgresPendaftarRepository) HardDeleteBySettingID(ctx context.Context, psbSettingID string) (int64, error) {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `DELETE FROM pendaftar WHERE psb_setting_id=$1`, psbSettingID)
	if err != nil {
		return 0, kernel.Wrap(pconstant.CodePendaftarPersistenceFailed, fmt.Errorf("hard delete pendaftar: %w", err))
	}
	return res.RowsAffected()
}

func scanPendaftar(sc scanner) (*pentity.Pendaftar, error) {
	var (
		id, userID, psbSettingID, gender                               string
		program                                                         sql.NullString
		nickname, hobby, purpose, motivationEntry, pob, blood          sql.NullString
		dob                                                            sql.NullTime
		address, subDistrict, district, province, postalCode           sql.NullString
		prevName, prevAddress, prevDiv, prevTime                       sql.NullString
		nik, noKK, nisn, noKIP, noKKS, noPKH                           sql.NullString
		workplace, department                                          sql.NullString
		homeStatus                                                     sql.NullString
		father, fatherPN, fatherNIK, fatherJob, fatherGraduate, fatherIncome sql.NullString
		mother, motherPN, motherNIK, motherJob, motherGraduate, motherIncome sql.NullString
		guardianRel, guardian, guardianPN, guardianNIK, guardianJob, guardianGraduate, guardianIncome sql.NullString
		status                                                         sql.NullString
		acceptedBy, santriID, nisVal                                   sql.NullString
		acceptedAt                                                     sql.NullTime
		createdAt, updatedAt                                           time.Time
		deletedAt                                                      sql.NullTime
	)

	err := sc.Scan(
		&id, &userID, &psbSettingID, &gender, &program,
		&nickname, &hobby, &purpose, &motivationEntry, &pob, &dob, &blood,
		&address, &subDistrict, &district, &province, &postalCode,
		&prevName, &prevAddress, &prevDiv, &prevTime,
		&nik, &noKK, &nisn, &noKIP, &noKKS, &noPKH,
		&workplace, &department,
		&homeStatus,
		&father, &fatherPN, &fatherNIK, &fatherJob, &fatherGraduate, &fatherIncome,
		&mother, &motherPN, &motherNIK, &motherJob, &motherGraduate, &motherIncome,
		&guardianRel, &guardian, &guardianPN, &guardianNIK, &guardianJob, &guardianGraduate, &guardianIncome,
		&status, &acceptedBy, &acceptedAt, &santriID, &nisVal,
		&createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(pconstant.CodePendaftarNotFound)
		}
		return nil, kernel.Wrap(pconstant.CodePendaftarQueryFailed, fmt.Errorf("scan pendaftar: %w", err))
	}

	return &pentity.Pendaftar{
		ID:                   id,
		UserID:               userID,
		PsbSettingID:         psbSettingID,
		Gender:               gender,
		Program:              strFromNull(program),
		Nickname:             strFromNull(nickname),
		Hobby:                strFromNull(hobby),
		Purpose:              strFromNull(purpose),
		MotivationEntry:      strFromNull(motivationEntry),
		POB:                  strFromNull(pob),
		DOB:                  timeFromNull(dob),
		Blood:                strFromNull(blood),
		Address:              strFromNull(address),
		SubDistrict:          strFromNull(subDistrict),
		District:             strFromNull(district),
		Province:             strFromNull(province),
		PostalCode:           strFromNull(postalCode),
		PreviousPondokName:   strFromNull(prevName),
		PreviousPondokAddress: strFromNull(prevAddress),
		PreviousPondokDiv:    strFromNull(prevDiv),
		PreviousPondokTime:   strFromNull(prevTime),
		NIK:                  strFromNull(nik),
		NoKK:                 strFromNull(noKK),
		NISN:                 strFromNull(nisn),
		NoKIP:                strFromNull(noKIP),
		NoKKS:                strFromNull(noKKS),
		NoPKH:                strFromNull(noPKH),
		Workplace:            strFromNull(workplace),
		Department:           strFromNull(department),
		HomeStatus:           strFromNull(homeStatus),
		Father:               strFromNull(father),
		FatherPN:             strFromNull(fatherPN),
		FatherNIK:            strFromNull(fatherNIK),
		FatherJob:            strFromNull(fatherJob),
		FatherGraduate:       strFromNull(fatherGraduate),
		FatherIncome:         strFromNull(fatherIncome),
		Mother:               strFromNull(mother),
		MotherPN:             strFromNull(motherPN),
		MotherNIK:            strFromNull(motherNIK),
		MotherJob:            strFromNull(motherJob),
		MotherGraduate:       strFromNull(motherGraduate),
		MotherIncome:         strFromNull(motherIncome),
		GuardianRelationship: strFromNull(guardianRel),
		Guardian:             strFromNull(guardian),
		GuardianPN:           strFromNull(guardianPN),
		GuardianNIK:          strFromNull(guardianNIK),
		GuardianJob:          strFromNull(guardianJob),
		GuardianGraduate:     strFromNull(guardianGraduate),
		GuardianIncome:       strFromNull(guardianIncome),
		Status:               pconstant.PendaftarStatus(blankToDraft(status)),
		AcceptedBy:           strFromNull(acceptedBy),
		AcceptedAt:           timeFromNull(acceptedAt),
		SantriID:             strFromNull(santriID),
		NIS:                  strFromNull(nisVal),
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
		DeletedAt:            timeFromNull(deletedAt),
	}, nil
}

func blankToDraft(s sql.NullString) string {
	if !s.Valid || s.String == "" {
		return "draft"
	}
	return s.String
}
