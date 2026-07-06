package moomoo

import "math"

// Decimal precision applied to outgoing float fields to cut token-costly
// noise digits that no caller acts on.
const (
	pricePrecision    = 3 // prices: covers sub-cent tick sizes on some markets
	ratePrecision     = 4 // ratios such as change_rate
	turnoverPrecision = 0 // turnover: drop cents on these large figures
)

func roundTo(v float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals))
	return math.Round(v*p) / p
}

func roundPrice(v float64) float64    { return roundTo(v, pricePrecision) }
func roundRate(v float64) float64     { return roundTo(v, ratePrecision) }
func roundTurnover(v float64) float64 { return roundTo(v, turnoverPrecision) }
