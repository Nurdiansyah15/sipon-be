package valueobject

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TransactionID adalah order_id yang dikirim ke Midtrans. Harus unik dan
// panjangnya tidak melebihi 50 karakter (batasan Midtrans). Nilai asing hanya
// dipakai untuk debugging; aplikasi selalu menemukan transaksi lewat
// TransactionID yang tersimpan di database.
type TransactionID struct {
	value string
}

func NewTransactionID() TransactionID {
	return TransactionID{
		value: fmt.Sprintf("SIPON-%s-%s", time.Now().Format("20060102150405"), strings.ReplaceAll(uuid.NewString(), "-", "")[:8]),
	}
}

func (n TransactionID) String() string {
	return n.value
}

func (n TransactionID) IsEmpty() bool {
	return n.value == ""
}

func TransactionIDFromString(s string) TransactionID {
	return TransactionID{value: s}
}
