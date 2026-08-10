package service

import (
	"context"
	"fmt"

	"sipon-be/internal/modules/kesantrian/domain/surat/constant"
	"sipon-be/internal/modules/kesantrian/domain/surat/repository"
	"sipon-be/internal/shared/kernel"
)

type NomorGenerator struct {
	suratRepo repository.SuratRepository
}

func NewNomorGenerator(suratRepo repository.SuratRepository) *NomorGenerator {
	return &NomorGenerator{suratRepo: suratRepo}
}

func (g *NomorGenerator) Generate(ctx context.Context, kodeTipe string, bulan, tahun int) (string, int, error) {
	maxSeq, err := g.suratRepo.FindMaxSeqByMonthYear(ctx, bulan, tahun)
	if err != nil {
		return "", 0, kernel.Wrap(constant.CodeSuratNomorFailed, err)
	}

	newSeq := maxSeq + 1
	nomor := fmt.Sprintf("%03d/%s/%s/%s/%d", newSeq, kodeTipe, constant.OrgCode, toRoman(bulan), tahun)
	return nomor, newSeq, nil
}

func toRoman(month int) string {
	romans := []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	if month < 1 || month > 12 {
		return "?"
	}
	return romans[month-1]
}
