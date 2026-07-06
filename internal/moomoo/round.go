package moomoo

import (
	"math"
	"strings"
)

// Decimal precision applied to outgoing float fields to cut token-costly
// noise digits that no caller acts on.
const (
	pricePrecision    = 3 // prices: covers sub-cent tick sizes on most markets
	ratePrecision     = 4 // ratios such as change_rate
	turnoverPrecision = 0 // turnover: drop cents on these large figures

	// subPennyThreshold is the price below which markets in
	// subPennyTickMarkets may quote finer than pricePrecision allows, so
	// rounding is skipped entirely below it.
	subPennyThreshold = 1.0
)

// subPennyTickMarkets are quote markets whose exchanges allow tick sizes
// finer than pricePrecision for very low priced securities (e.g. the US
// SEC's sub-penny rule, which permits $0.0001 ticks for stocks under $1).
// Other supported markets don't have this problem at any price: JPY has no
// sub-unit so JP prices are always whole numbers, and SH/SZ use a flat 0.01
// CNY tick.
var subPennyTickMarkets = map[string]bool{
	"US": true,
	"HK": true,
	"SG": true,
}

func roundTo(v float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals))
	return math.Round(v*p) / p
}

// roundPrice rounds a price to pricePrecision, unless market is a
// sub-penny-tick market (subPennyTickMarkets) and v is below
// subPennyThreshold, in which case v is returned unrounded to avoid losing
// real tick precision. Returns v unrounded when the client was constructed
// with rounding disabled.
func (c *Client) roundPrice(v float64, market string) float64 {
	if c.disableRounding {
		return v
	}
	if subPennyTickMarkets[market] && v > 0 && v < subPennyThreshold {
		return v
	}
	return roundTo(v, pricePrecision)
}

func (c *Client) roundRate(v float64) float64 {
	if c.disableRounding {
		return v
	}
	return roundTo(v, ratePrecision)
}

func (c *Client) roundTurnover(v float64) float64 {
	if c.disableRounding {
		return v
	}
	return roundTo(v, turnoverPrecision)
}

// marketFromCode extracts the market prefix from a security code such as
// "US.AAPL" or "HK.00700".
func marketFromCode(code string) string {
	if i := strings.IndexByte(code, '.'); i >= 0 {
		return code[:i]
	}
	return ""
}
