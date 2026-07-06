package moomoo

import "testing"

func TestRoundPrice(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{123.45, 123.45},
		{123.45001, 123.45},
		{123.4567, 123.457},
		{0, 0},
	}
	for _, c := range cases {
		if got := roundPrice(c.in); got != c.want {
			t.Errorf("roundPrice(%v) = %v, want %v", c.in, got, c.want)
		}
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
	for _, c := range cases {
		if got := roundRate(c.in); got != c.want {
			t.Errorf("roundRate(%v) = %v, want %v", c.in, got, c.want)
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
	for _, c := range cases {
		if got := roundTurnover(c.in); got != c.want {
			t.Errorf("roundTurnover(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
