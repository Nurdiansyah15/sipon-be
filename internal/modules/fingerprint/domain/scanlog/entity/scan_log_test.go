package entity

import (
	"testing"
	"time"
)

func TestNewScanLogValid(t *testing.T) {
	log, err := NewScanLog("id-1", "SN-1", "1000", "192.168.1.10", time.Now(), 1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.ID != "id-1" || log.SN != "SN-1" || log.PIN != "1000" {
		t.Fatalf("unexpected scan log: %+v", log)
	}
}

func TestNewScanLogRejectsEmptyFields(t *testing.T) {
	cases := []struct {
		name string
		id   string
		sn   string
		pin  string
	}{
		{"empty id", "", "SN-1", "1000"},
		{"empty sn", "id-1", "", "1000"},
		{"empty pin", "id-1", "SN-1", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewScanLog(tc.id, tc.sn, tc.pin, "", time.Now(), 0, 0); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
