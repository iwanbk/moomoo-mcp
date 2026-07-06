package config

import (
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("MOOMOO_OPEND_HOST", "")
	t.Setenv("MOOMOO_OPEND_PORT", "")
	t.Setenv("MOOMOO_TRADE_PASSWORD", "")
	t.Setenv("MOOMOO_TRADE_PASSWORD_MD5", "")
	t.Setenv("MOOMOO_SECURITY_FIRM", "")
	t.Setenv("MOOMOO_DISABLE_ROUNDING", "")

	cfg := Load()

	if cfg.OpendHost != defaultHost {
		t.Errorf("OpendHost = %q, want %q", cfg.OpendHost, defaultHost)
	}
	if cfg.OpendPort != defaultPort {
		t.Errorf("OpendPort = %d, want %d", cfg.OpendPort, defaultPort)
	}
	if !cfg.SimulateOnly {
		t.Error("SimulateOnly should be true when no trade password is set")
	}
	if cfg.DisableRounding {
		t.Error("DisableRounding should default to false")
	}
}

func TestLoad_custom(t *testing.T) {
	t.Setenv("MOOMOO_OPEND_HOST", "192.168.1.100")
	t.Setenv("MOOMOO_OPEND_PORT", "22222")
	t.Setenv("MOOMOO_TRADE_PASSWORD", "secret")
	t.Setenv("MOOMOO_TRADE_PASSWORD_MD5", "")
	t.Setenv("MOOMOO_SECURITY_FIRM", "FUTUSG")
	t.Setenv("MOOMOO_DISABLE_ROUNDING", "true")

	cfg := Load()

	if cfg.OpendHost != "192.168.1.100" {
		t.Errorf("OpendHost = %q", cfg.OpendHost)
	}
	if cfg.OpendPort != 22222 {
		t.Errorf("OpendPort = %d", cfg.OpendPort)
	}
	if cfg.SimulateOnly {
		t.Error("SimulateOnly should be false when trade password is set")
	}
	if cfg.SecurityFirm != "FUTUSG" {
		t.Errorf("SecurityFirm = %q", cfg.SecurityFirm)
	}
	if !cfg.DisableRounding {
		t.Error("DisableRounding should be true when MOOMOO_DISABLE_ROUNDING=true")
	}
}

func TestLoad_simulateOnlyWithMD5(t *testing.T) {
	t.Setenv("MOOMOO_TRADE_PASSWORD", "")
	t.Setenv("MOOMOO_TRADE_PASSWORD_MD5", "abc123md5hash")

	cfg := Load()

	if cfg.SimulateOnly {
		t.Error("SimulateOnly should be false when MD5 password is set")
	}
	if cfg.TradePasswordMD5 != "abc123md5hash" {
		t.Errorf("TradePasswordMD5 = %q", cfg.TradePasswordMD5)
	}
}

func TestParsePort_fallback(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", defaultPort},
		{"abc", defaultPort},
		{"0", defaultPort},
		{"-1", defaultPort},
		{"99999", defaultPort},
		{"8888", 8888},
		{"11111", 11111},
		{"65535", 65535},
	}
	for _, c := range cases {
		got := parsePort(c.in)
		if got != c.want {
			t.Errorf("parsePort(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
