package valueobject

import (
	"regexp"
	"strings"

	"sipon-be/internal/modules/kesantrian/domain/santri/constant"
	"sipon-be/internal/shared/kernel"
)

// nisPattern is the SINGLE source of the NIS format rule in this codebase:
// fixed "1000" prefix, gender digit ('1'=laki-laki/'2'=perempuan), then 5
// digits. Identity's side never re-validates this — it only stores the
// value it's handed as a login identifier.
var nisPattern = regexp.MustCompile(`^1000[12][0-9]{5}$`)

type NIS struct {
	value string
}

func NewNIS(raw string) (NIS, error) {
	raw = strings.TrimSpace(raw)
	if !nisPattern.MatchString(raw) {
		return NIS{}, kernel.New(constant.CodeInvalidNISFormat)
	}
	return NIS{value: raw}, nil
}

func (n NIS) String() string {
	return n.value
}

func (n NIS) IsEmpty() bool {
	return n.value == ""
}

// Gender returns the NIS's 5th character: '1' = laki-laki, '2' = perempuan.
// Mirrors the "option" column, which is derived from this at creation time.
func (n NIS) Gender() string {
	if len(n.value) < 5 {
		return ""
	}
	return string(n.value[4])
}
