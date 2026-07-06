package config

import (
	"os"
	"strconv"
)

const (
	defaultHost = "127.0.0.1"
	defaultPort = 11111
)

// Config holds all runtime configuration parsed from environment variables.
type Config struct {
	OpendHost        string
	OpendPort        int
	TradePassword    string
	TradePasswordMD5 string
	SecurityFirm     string
	// SimulateOnly is true when no trade password is configured, preventing
	// accidental real-account operations.
	SimulateOnly bool
	// DisableRounding turns off number rounding on outgoing float fields
	// (prices, rates, turnover), returning raw SDK values instead.
	DisableRounding bool
}

// Load reads configuration from environment variables with safe defaults.
func Load() *Config {
	tradePass := os.Getenv("MOOMOO_TRADE_PASSWORD")
	tradePassMD5 := os.Getenv("MOOMOO_TRADE_PASSWORD_MD5")

	return &Config{
		OpendHost:        getenv("MOOMOO_OPEND_HOST", defaultHost),
		OpendPort:        parsePort(os.Getenv("MOOMOO_OPEND_PORT")),
		TradePassword:    tradePass,
		TradePasswordMD5: tradePassMD5,
		SecurityFirm:     os.Getenv("MOOMOO_SECURITY_FIRM"),
		SimulateOnly:     tradePass == "" && tradePassMD5 == "",
		DisableRounding:  parseBool(os.Getenv("MOOMOO_DISABLE_ROUNDING")),
	}
}

// parseBool parses a boolean env var, treating any unparseable or empty
// value as false.
func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// parsePort parses a port string, falling back to defaultPort for any invalid input.
func parsePort(s string) int {
	if s == "" {
		return defaultPort
	}
	p, err := strconv.Atoi(s)
	if err != nil || p <= 0 || p > 65535 {
		return defaultPort
	}
	return p
}
