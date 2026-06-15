package realip

import (
	"errors"
	"net"
	"testing"
)

func TestCheckTrustedSubnet(t *testing.T) {
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		subnet  *net.IPNet
		realIP  string
		wantErr error
	}{
		{name: "disabled", subnet: nil, wantErr: nil},
		{name: "allowed", subnet: subnet, realIP: " 192.168.1.10 ", wantErr: nil},
		{name: "missing", subnet: subnet, wantErr: ErrMissing},
		{name: "invalid", subnet: subnet, realIP: "not-an-ip", wantErr: ErrUntrusted},
		{name: "denied", subnet: subnet, realIP: "10.0.0.1", wantErr: ErrUntrusted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckTrustedSubnet(tt.subnet, tt.realIP)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
