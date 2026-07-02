package moomoo

import "testing"

func TestTradeHeader_defaultsToSimulate(t *testing.T) {
	c := &Client{simulateOnly: true}
	h, err := c.tradeHeader(12345, "", "US")
	if err != nil {
		t.Fatal(err)
	}
	if h.GetTrdEnv() != 0 {
		t.Errorf("want SIMULATE (0), got %d", h.GetTrdEnv())
	}
	if h.GetAccID() != 12345 {
		t.Errorf("want accID 12345, got %d", h.GetAccID())
	}
}

func TestTradeHeader_realBlockedWhenSimulateOnly(t *testing.T) {
	c := &Client{simulateOnly: true}
	if _, err := c.tradeHeader(12345, "REAL", "US"); err == nil {
		t.Fatal("want error when requesting REAL header on a simulate-only client")
	}
}

func TestTradeHeader_realAllowedWhenNotSimulateOnly(t *testing.T) {
	c := &Client{simulateOnly: false}
	h, err := c.tradeHeader(12345, "REAL", "US")
	if err != nil {
		t.Fatal(err)
	}
	if h.GetTrdEnv() != 1 {
		t.Errorf("want REAL (1), got %d", h.GetTrdEnv())
	}
}

func TestTradeHeader_unknownMarket(t *testing.T) {
	c := &Client{simulateOnly: true}
	if _, err := c.tradeHeader(12345, "SIMULATE", "MARS"); err == nil {
		t.Fatal("want error for unknown trd_market")
	}
}

func TestTradeHeader_unknownEnv(t *testing.T) {
	c := &Client{simulateOnly: true}
	if _, err := c.tradeHeader(12345, "BOGUS", "US"); err == nil {
		t.Fatal("want error for unknown trd_env")
	}
}

func TestTradeHeader_accountIDPrecision(t *testing.T) {
	const bigID uint64 = 9223372036854775807 // math.MaxInt64
	c := &Client{simulateOnly: true}
	h, err := c.tradeHeader(bigID, "SIMULATE", "US")
	if err != nil {
		t.Fatal(err)
	}
	if h.GetAccID() != bigID {
		t.Errorf("account ID lost precision: want %d, got %d", bigID, h.GetAccID())
	}
}
