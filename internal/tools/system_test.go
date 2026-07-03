package tools

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

type mockClient struct {
	healthErr error
}

func (m *mockClient) Health(_ context.Context) error { return m.healthErr }
func (m *mockClient) Close() error                   { return nil }

func (m *mockClient) GetSnapshot(_ context.Context, _ []string) ([]moomoo.MarketSnapshot, error) {
	return nil, nil
}
func (m *mockClient) GetQuote(_ context.Context, _ []string) ([]moomoo.Quote, error) {
	return nil, nil
}
func (m *mockClient) GetKlines(_ context.Context, _ string, _ int32, _, _ string) ([]moomoo.Kline, error) {
	return nil, nil
}
func (m *mockClient) GetOrderBook(_ context.Context, _ string) (*moomoo.OrderBook, error) {
	return nil, nil
}
func (m *mockClient) GetAccounts(_ context.Context) ([]moomoo.Account, error) {
	return nil, nil
}
func (m *mockClient) GetAssets(_ context.Context, _ uint64, _, _ string) (*moomoo.Assets, error) {
	return nil, nil
}
func (m *mockClient) GetPositions(_ context.Context, _ uint64, _, _ string) ([]moomoo.Position, error) {
	return nil, nil
}
func (m *mockClient) GetMaxTradable(_ context.Context, _ uint64, _, _, _, _ string, _ float64) (*moomoo.MaxTradable, error) {
	return nil, nil
}
func (m *mockClient) GetMarginRatio(_ context.Context, _ uint64, _, _ string, _ []string) ([]moomoo.MarginRatio, error) {
	return nil, nil
}
func (m *mockClient) GetCashFlow(_ context.Context, _ uint64, _, _, _ string) ([]moomoo.CashFlow, error) {
	return nil, nil
}
func (m *mockClient) GetUserSecurityGroup(_ context.Context, _ string) ([]moomoo.SecurityGroup, error) {
	return nil, nil
}
func (m *mockClient) GetUserSecurity(_ context.Context, _ string) ([]moomoo.WatchlistSecurity, error) {
	return nil, nil
}
func (m *mockClient) GetOrders(_ context.Context, _ uint64, _, _ string) ([]moomoo.Order, error) {
	return nil, nil
}
func (m *mockClient) GetDeals(_ context.Context, _ uint64, _, _ string) ([]moomoo.Deal, error) {
	return nil, nil
}
func (m *mockClient) GetHistoryOrders(_ context.Context, _ uint64, _, _, _, _ string, _ []string) ([]moomoo.Order, error) {
	return nil, nil
}
func (m *mockClient) GetHistoryDeals(_ context.Context, _ uint64, _, _, _, _ string, _ []string) ([]moomoo.Deal, error) {
	return nil, nil
}

// connect creates an in-memory MCP client session connected to the given server.
func connectTest(ctx context.Context, s *mcp.Server) (*mcp.ClientSession, func()) {
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := s.Connect(ctx, t1, nil); err != nil {
		log.Fatalf("server connect: %v", err)
	}
	c := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	cs, err := c.Connect(ctx, t2, nil)
	if err != nil {
		log.Fatalf("client connect: %v", err)
	}
	return cs, func() { cs.Close() }
}

func TestCheckHealth_connected(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterSystem(s, &mockClient{})

	ctx := context.Background()
	cs, cleanup := connectTest(ctx, s)
	defer cleanup()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "check_health"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}

	text := res.Content[0].(*mcp.TextContent).Text
	var h healthResult
	if err := json.Unmarshal([]byte(text), &h); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if !h.Connected {
		t.Errorf("want connected=true, got %+v", h)
	}
	if h.Error != "" {
		t.Errorf("want empty error, got %q", h.Error)
	}
}

func TestCheckHealth_disconnected(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterSystem(s, &mockClient{healthErr: errors.New("dial failed")})

	ctx := context.Background()
	cs, cleanup := connectTest(ctx, s)
	defer cleanup()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "check_health"})
	if err != nil {
		t.Fatal(err)
	}

	text := res.Content[0].(*mcp.TextContent).Text
	var h healthResult
	if err := json.Unmarshal([]byte(text), &h); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if h.Connected {
		t.Error("want connected=false")
	}
	if h.Error == "" {
		t.Error("want non-empty error message")
	}
}
