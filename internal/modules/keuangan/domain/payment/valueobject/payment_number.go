package valueobject

import (
	"fmt"
	"time"
)

type PaymentNumber struct {
	value string
}

func NewPaymentNumber(year, month string, seq int) PaymentNumber {
	return PaymentNumber{
		value: fmt.Sprintf("PAY/%s/%s/%06d", year, month, seq),
	}
}

func NewPaymentNumberNow(seq int) PaymentNumber {
	now := time.Now()
	return NewPaymentNumber(fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", int(now.Month())), seq)
}

func (n PaymentNumber) String() string {
	return n.value
}

func (n PaymentNumber) IsEmpty() bool {
	return n.value == ""
}

func PaymentNumberFromString(s string) PaymentNumber {
	return PaymentNumber{value: s}
}
