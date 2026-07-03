package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

// watchlistMockClient is a configurable mock for watchlist tests.
type watchlistMockClient struct {
	groupsResult []moomoo.SecurityGroup
	groupsErr    error

	securitiesResult []moomoo.WatchlistSecurity
	securitiesErr    error

	// gotGroupType/gotGroupName record the last arguments passed in, so tests
	// can assert they were forwarded correctly.
	gotGroupType string
	gotGroupName string
}

func (m *watchlistMockClient) Health(_ context.Context) error { return nil }
func (m *watchlistMockClient) Close() error                   { return nil }
func (m *watchlistMockClient) GetSnapshot(_ context.Context, _ []string) ([]moomoo.MarketSnapshot, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetQuote(_ context.Context, _ []string) ([]moomoo.Quote, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetKlines(_ context.Context, _ string, _ int32, _, _ string) ([]moomoo.Kline, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetOrderBook(_ context.Context, _ string) (*moomoo.OrderBook, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetAccounts(_ context.Context) ([]moomoo.Account, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetAssets(_ context.Context, _ uint64, _, _ string) (*moomoo.Assets, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetPositions(_ context.Context, _ uint64, _, _ string) ([]moomoo.Position, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetMaxTradable(_ context.Context, _ uint64, _, _, _, _ string, _ float64) (*moomoo.MaxTradable, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetMarginRatio(_ context.Context, _ uint64, _, _ string, _ []string) ([]moomoo.MarginRatio, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetCashFlow(_ context.Context, _ uint64, _, _, _ string) ([]moomoo.CashFlow, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetUserSecurityGroup(_ context.Context, groupType string) ([]moomoo.SecurityGroup, error) {
	m.gotGroupType = groupType
	return m.groupsResult, m.groupsErr
}
func (m *watchlistMockClient) GetUserSecurity(_ context.Context, groupName string) ([]moomoo.WatchlistSecurity, error) {
	m.gotGroupName = groupName
	return m.securitiesResult, m.securitiesErr
}
func (m *watchlistMockClient) GetOrders(_ context.Context, _ uint64, _, _ string) ([]moomoo.Order, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetDeals(_ context.Context, _ uint64, _, _ string) ([]moomoo.Deal, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetHistoryOrders(_ context.Context, _ uint64, _, _, _, _ string, _ []string) ([]moomoo.Order, error) {
	return nil, nil
}
func (m *watchlistMockClient) GetHistoryDeals(_ context.Context, _ uint64, _, _, _, _ string, _ []string) ([]moomoo.Deal, error) {
	return nil, nil
}

func newWatchlistServer(c moomoo.MoomooClient) (*mcp.ClientSession, func()) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterWatchlist(s, c)
	ctx := context.Background()
	cs, cleanup := connectTest(ctx, s)
	return cs, cleanup
}

// --- get_user_security_group ---

func TestGetUserSecurityGroup_success(t *testing.T) {
	want := []moomoo.SecurityGroup{
		{GroupName: "My Watchlist", GroupType: "CUSTOM"},
	}
	mock := &watchlistMockClient{groupsResult: want}
	cs, cleanup := newWatchlistServer(mock)
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_user_security_group",
		Arguments: map[string]any{"group_type": "CUSTOM"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.SecurityGroup
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].GroupName != "My Watchlist" || got[0].GroupType != "CUSTOM" {
		t.Errorf("unexpected result: %+v", got)
	}
	if mock.gotGroupType != "CUSTOM" {
		t.Errorf("want group_type forwarded as CUSTOM, got %q", mock.gotGroupType)
	}
}

func TestGetUserSecurityGroup_error(t *testing.T) {
	cs, cleanup := newWatchlistServer(&watchlistMockClient{groupsErr: errors.New("network error")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_user_security_group"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_user_security ---

func TestGetUserSecurity_success(t *testing.T) {
	want := []moomoo.WatchlistSecurity{
		{Code: "US.AAPL", Name: "Apple", LotSize: 1, SecType: "EQUITY"},
	}
	mock := &watchlistMockClient{securitiesResult: want}
	cs, cleanup := newWatchlistServer(mock)
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_user_security",
		Arguments: map[string]any{"group_name": "My Watchlist"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.WatchlistSecurity
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].Code != "US.AAPL" || got[0].SecType != "EQUITY" {
		t.Errorf("unexpected result: %+v", got)
	}
	if mock.gotGroupName != "My Watchlist" {
		t.Errorf("want group_name forwarded as %q, got %q", "My Watchlist", mock.gotGroupName)
	}
}

func TestGetUserSecurity_error(t *testing.T) {
	cs, cleanup := newWatchlistServer(&watchlistMockClient{securitiesErr: errors.New("group not found")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_user_security",
		Arguments: map[string]any{"group_name": "Nonexistent"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}
