package valueobject

import (
	"fmt"
	"time"
)

type InvoiceNumber struct {
	value string
}

func NewInvoiceNumber(year, month string, seq int) InvoiceNumber {
	return InvoiceNumber{
		value: fmt.Sprintf("INV/%s/%s/%06d", year, month, seq),
	}
}

func NewInvoiceNumberNow(seq int) InvoiceNumber {
	now := time.Now()
	return NewInvoiceNumber(fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", int(now.Month())), seq)
}

func (n InvoiceNumber) String() string {
	return n.value
}

func (n InvoiceNumber) IsEmpty() bool {
	return n.value == ""
}

func InvoiceNumberFromString(s string) InvoiceNumber {
	return InvoiceNumber{value: s}
}
