package service

import "testing"

func TestInvoiceNet(t *testing.T) {
	cases := []struct {
		name              string
		exec, mat, disc   float64
		free              bool
		wantDisc, wantNet float64
	}{
		{"عادي", 100000, 50000, 20000, false, 20000, 130000},
		{"بلا خصم", 100000, 0, 0, false, 0, 100000},
		{"خصم سالب ينصفّر", 100000, 0, -5000, false, 0, 100000},
		{"خصم أكبر من المبلغ", 100000, 0, 250000, false, 250000, 0},
		{"مجاني: الكلفة تبقى", 100000, 50000, 999, true, 150000, 0},
	}
	for _, c := range cases {
		d, n := invoiceNet(c.exec, c.mat, c.disc, c.free)
		if d != c.wantDisc || n != c.wantNet {
			t.Errorf("%s: خصم %v صافي %v، المتوقع %v و %v", c.name, d, n, c.wantDisc, c.wantNet)
		}
	}
}
