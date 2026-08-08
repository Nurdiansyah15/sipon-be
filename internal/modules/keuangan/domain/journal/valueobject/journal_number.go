package valueobject

import (
	"fmt"
	"time"
)

type JournalNumber struct {
	value string
}

func NewJournalNumber(year, month string, seq int) JournalNumber {
	return JournalNumber{
		value: fmt.Sprintf("JRN/%s/%s/%06d", year, month, seq),
	}
}

func NewJournalNumberNow(seq int) JournalNumber {
	now := time.Now()
	return NewJournalNumber(fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", int(now.Month())), seq)
}

func (n JournalNumber) String() string {
	return n.value
}

func (n JournalNumber) IsEmpty() bool {
	return n.value == ""
}

func JournalNumberFromString(s string) JournalNumber {
	return JournalNumber{value: s}
}
