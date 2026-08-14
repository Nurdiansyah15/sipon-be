package command

import (
	"sipon-be/internal/modules/kesantrian/application"
	santrivo "sipon-be/internal/modules/kesantrian/domain/santri/valueobject"
	"sipon-be/internal/shared/kernel"
)

// validateGenderMatchesNIS memastikan gender yang dikirim ('1' laki-laki /
// '2' perempuan) konsisten dengan digit gender yang sudah ter-encode di NIS.
// Gender opsional (nil) berarti mengandalkan digit gender NIS sebagai sumber
// kebenaran — pola yang dipakai semua alur pembuatan santri.
func validateGenderMatchesNIS(gender *string, nis santrivo.NIS) error {
	if gender == nil {
		return nil
	}
	if *gender != "1" && *gender != "2" {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "jenis kelamin tidak valid (harus '1' atau '2')", nil)
	}
	if *gender != nis.Gender() {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "jenis kelamin tidak sesuai dengan digit gender pada NIS", nil)
	}
	return nil
}
