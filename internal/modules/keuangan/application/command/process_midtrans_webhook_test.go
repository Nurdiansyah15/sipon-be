package command

import (
	"testing"

	pgConst "sipon-be/internal/modules/keuangan/domain/paymentgateway/constant"
)

func TestMapMidtransStatus(t *testing.T) {
	tests := []struct {
		name              string
		transactionStatus string
		fraudStatus       string
		want              pgConst.PaymentGatewayStatus
	}{
		{"capture accept", "capture", "accept", pgConst.GatewayStatusSettlement},
		{"capture challenge", "capture", "challenge", pgConst.GatewayStatusPendingChallenge},
		{"capture no fraud", "capture", "", pgConst.GatewayStatusSettlement},
		{"settlement", "settlement", "", pgConst.GatewayStatusSettlement},
		{"pending", "pending", "", pgConst.GatewayStatusPending},
		{"deny", "deny", "", pgConst.GatewayStatusDeny},
		{"cancel", "cancel", "", pgConst.GatewayStatusCancel},
		{"expire", "expire", "", pgConst.GatewayStatusExpire},
		{"failure", "failure", "", pgConst.GatewayStatusFailure},
		{"unknown falls back to pending", "weird-status", "", pgConst.GatewayStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapMidtransStatus(tt.transactionStatus, tt.fraudStatus)
			if got != tt.want {
				t.Errorf("mapMidtransStatus(%q, %q) = %s, want %s", tt.transactionStatus, tt.fraudStatus, got, tt.want)
			}
		})
	}
}
