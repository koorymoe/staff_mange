package service

import (
	"testing"

	"staffmange-api/internal/model"
)

func TestValidateQuotationAmounts(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	ok := []model.QuotationItemInput{{Quantity: 2, UnitPrice: 1000}}
	cases := []struct {
		name  string
		items []model.QuotationItemInput
		disc  *float64
		bad   bool
	}{
		{"عادي", ok, f(10), false},
		{"بلا خصم", ok, nil, false},
		{"خصم ١٠٠", ok, f(100), false},
		{"خصم ١٥٠", ok, f(150), true},
		{"خصم سالب", ok, f(-5), true},
		{"سعر سالب", []model.QuotationItemInput{{Quantity: 1, UnitPrice: -1}}, nil, true},
		{"كمية سالبة", []model.QuotationItemInput{{Quantity: -3, UnitPrice: 10}}, nil, true},
	}
	for _, c := range cases {
		if err := validateQuotationAmounts(c.items, c.disc); (err != nil) != c.bad {
			t.Errorf("%s: err=%v", c.name, err)
		}
	}
}
