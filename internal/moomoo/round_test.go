package moomoo

import "testing"

func TestRoundPrice(t *testing.T) {
	cases := []struct {
		name   string
		in     float64
		market string
		want   float64
	}{
		{"plain rounding", 123.45, "US", 123.45},
		{"rounds down noise", 123.45001, "US", 123.45},
		{"rounds to 3 decimals", 123.4567, "US", 123.457},
		{"zero", 0, "US", 0},
		{"US sub-penny below threshold is untouched", 0.12345, "US", 0.12345},
		{"HK sub-penny below threshold is untouched", 0.00001, "HK", 0.00001},
		{"SG sub-penny below threshold is untouched", 0.9999, "SG", 0.9999},
		{"US at or above threshold still rounds", 1.23456, "US", 1.235},
		{"JP has no sub-penny exception below 1", 0.12345, "JP", 0.123},
		{"SH has no sub-penny exception below 1", 0.12345, "SH", 0.123},
		{"unknown market has no sub-penny exception", 0.12345, "", 0.123},
	}
	c := &Client{}
	for _, tc := range cases {
		if got := c.roundPrice(tc.in, tc.market); got != tc.want {
			t.Errorf("%s: roundPrice(%v, %q) = %v, want %v", tc.name, tc.in, tc.market, got, tc.want)
		}
	}
}

func TestRoundPrice_disabled(t *testing.T) {
	c := &Client{disableRounding: true}
	in := 123.456789
	if got := c.roundPrice(in, "US"); got != in {
		t.Errorf("roundPrice with disableRounding = %v, want unrounded %v", got, in)
	}
}

func TestRoundRate(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{0.023, 0.023},
		{0.023456, 0.0235},
		{-0.5, -0.5},
	}
	c := &Client{}
	for _, tc := range cases {
		if got := c.roundRate(tc.in); got != tc.want {
			t.Errorf("roundRate(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestRoundTurnover(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{123456789.12, 123456789},
		{123456789.5, 123456790},
		{0.4, 0},
	}
	c := &Client{}
	for _, tc := range cases {
		if got := c.roundTurnover(tc.in); got != tc.want {
			t.Errorf("roundTurnover(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestMarketFromCode(t *testing.T) {
	cases := []struct {
		code, want string
	}{
		{"US.AAPL", "US"},
		{"HK.00700", "HK"},
		{"", ""},
		{"NODOT", ""},
	}
	for _, tc := range cases {
		if got := marketFromCode(tc.code); got != tc.want {
			t.Errorf("marketFromCode(%q) = %q, want %q", tc.code, got, tc.want)
		}
	}
}
