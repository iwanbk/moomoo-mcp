package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

// ordersMockClient is a configurable mock for order history tests.
type ordersMockClient struct {
	ordersResult []moomoo.Order
	ordersErr    error

	dealsResult []moomoo.Deal
	dealsErr    error

	historyOrdersResult []moomoo.Order
	historyOrdersErr    error

	historyDealsResult []moomoo.Deal
	historyDealsErr    error

	// gotAccountID/gotBeginTime/gotEndTime/gotCodes record the last arguments
	// passed in, so tests can assert they were forwarded correctly.
	gotAccountID uint64
	gotBeginTime string
	gotEndTime   string
	gotCodes     []string
}

func (m *ordersMockClient) Health(_ context.Context) error { return nil }
func (m *ordersMockClient) Close() error                   { return nil }
func (m *ordersMockClient) GetSnapshot(_ context.Context, _ []string) ([]moomoo.MarketSnapshot, error) {
	return nil, nil
}
func (m *ordersMockClient) GetQuote(_ context.Context, _ []string) ([]moomoo.Quote, error) {
	return nil, nil
}
func (m *ordersMockClient) GetKlines(_ context.Context, _ string, _ int32, _, _ string) ([]moomoo.Kline, error) {
	return nil, nil
}
func (m *ordersMockClient) GetOrderBook(_ context.Context, _ string) (*moomoo.OrderBook, error) {
	return nil, nil
}
func (m *ordersMockClient) GetAccounts(_ context.Context) ([]moomoo.Account, error) {
	return nil, nil
}
func (m *ordersMockClient) GetAssets(_ context.Context, _ uint64, _, _ string) (*moomoo.Assets, error) {
	return nil, nil
}
func (m *ordersMockClient) GetPositions(_ context.Context, _ uint64, _, _ string) ([]moomoo.Position, error) {
	return nil, nil
}
func (m *ordersMockClient) GetMaxTradable(_ context.Context, _ uint64, _, _, _, _ string, _ float64) (*moomoo.MaxTradable, error) {
	return nil, nil
}
func (m *ordersMockClient) GetMarginRatio(_ context.Context, _ uint64, _, _ string, _ []string) ([]moomoo.MarginRatio, error) {
	return nil, nil
}
func (m *ordersMockClient) GetCashFlow(_ context.Context, _ uint64, _, _, _ string) ([]moomoo.CashFlow, error) {
	return nil, nil
}
func (m *ordersMockClient) GetUserSecurityGroup(_ context.Context, _ string) ([]moomoo.SecurityGroup, error) {
	return nil, nil
}
func (m *ordersMockClient) GetUserSecurity(_ context.Context, _ string) ([]moomoo.WatchlistSecurity, error) {
	return nil, nil
}
func (m *ordersMockClient) GetOrders(_ context.Context, accountID uint64, _, _ string) ([]moomoo.Order, error) {
	m.gotAccountID = accountID
	return m.ordersResult, m.ordersErr
}
func (m *ordersMockClient) GetDeals(_ context.Context, accountID uint64, _, _ string) ([]moomoo.Deal, error) {
	m.gotAccountID = accountID
	return m.dealsResult, m.dealsErr
}
func (m *ordersMockClient) GetHistoryOrders(_ context.Context, accountID uint64, _, _, beginTime, endTime string, codes []string) ([]moomoo.Order, error) {
	m.gotAccountID = accountID
	m.gotBeginTime = beginTime
	m.gotEndTime = endTime
	m.gotCodes = codes
	return m.historyOrdersResult, m.historyOrdersErr
}
func (m *ordersMockClient) GetHistoryDeals(_ context.Context, accountID uint64, _, _, beginTime, endTime string, codes []string) ([]moomoo.Deal, error) {
	m.gotAccountID = accountID
	m.gotBeginTime = beginTime
	m.gotEndTime = endTime
	m.gotCodes = codes
	return m.historyDealsResult, m.historyDealsErr
}

func newOrdersServer(c moomoo.MoomooClient) (*mcp.ClientSession, func()) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterOrders(s, c)
	ctx := context.Background()
	cs, cleanup := connectTest(ctx, s)
	return cs, cleanup
}

// --- get_orders ---

func TestGetOrders_success(t *testing.T) {
	want := []moomoo.Order{
		{OrderID: "123", Code: "US.AAPL", TrdSide: "BUY", OrderStatus: "SUBMITTED", Qty: 10},
	}
	mock := &ordersMockClient{ordersResult: want}
	cs, cleanup := newOrdersServer(mock)
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_orders",
		Arguments: map[string]any{"account_id": "123456789012345678", "trd_market": "US"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.Order
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].OrderID != "123" || got[0].Code != "US.AAPL" {
		t.Errorf("unexpected result: %+v", got)
	}
	if mock.gotAccountID != 123456789012345678 {
		t.Errorf("want account_id forwarded as 123456789012345678, got %d", mock.gotAccountID)
	}
}

func TestGetOrders_error(t *testing.T) {
	cs, cleanup := newOrdersServer(&ordersMockClient{ordersErr: errors.New("network error")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_orders",
		Arguments: map[string]any{"account_id": "123", "trd_market": "US"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

func TestGetOrders_invalidAccountID(t *testing.T) {
	cs, cleanup := newOrdersServer(&ordersMockClient{})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_orders",
		Arguments: map[string]any{"account_id": "not-a-number", "trd_market": "US"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_deals ---

func TestGetDeals_success(t *testing.T) {
	want := []moomoo.Deal{
		{FillID: "789", Code: "US.AAPL", TrdSide: "BUY", Qty: 10, Price: 150.5},
	}
	mock := &ordersMockClient{dealsResult: want}
	cs, cleanup := newOrdersServer(mock)
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_deals",
		Arguments: map[string]any{"account_id": "123", "trd_market": "US"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.Deal
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].FillID != "789" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestGetDeals_error(t *testing.T) {
	cs, cleanup := newOrdersServer(&ordersMockClient{dealsErr: errors.New("network error")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_deals",
		Arguments: map[string]any{"account_id": "123", "trd_market": "US"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_history_orders ---

func TestGetHistoryOrders_success(t *testing.T) {
	want := []moomoo.Order{
		{OrderID: "123", Code: "US.AAPL", OrderStatus: "FILLED_ALL", Qty: 10},
	}
	mock := &ordersMockClient{historyOrdersResult: want}
	cs, cleanup := newOrdersServer(mock)
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_history_orders",
		Arguments: map[string]any{
			"account_id": "123",
			"trd_market": "US",
			"begin_time": "2026-06-01 00:00:00",
			"end_time":   "2026-07-01 00:00:00",
			"codes":      []string{"US.AAPL"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got Columnar
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got.Rows) != 1 || got.Rows[0][0] != "123" {
		t.Errorf("unexpected rows: %+v", got.Rows)
	}
	if mock.gotBeginTime != "2026-06-01 00:00:00" || mock.gotEndTime != "2026-07-01 00:00:00" {
		t.Errorf("want begin/end time forwarded, got %q/%q", mock.gotBeginTime, mock.gotEndTime)
	}
	if len(mock.gotCodes) != 1 || mock.gotCodes[0] != "US.AAPL" {
		t.Errorf("want codes forwarded as [US.AAPL], got %v", mock.gotCodes)
	}
}

func TestGetHistoryOrders_error(t *testing.T) {
	cs, cleanup := newOrdersServer(&ordersMockClient{historyOrdersErr: errors.New("bad time range")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_history_orders",
		Arguments: map[string]any{
			"account_id": "123",
			"trd_market": "US",
			"begin_time": "2026-06-01 00:00:00",
			"end_time":   "2026-07-01 00:00:00",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_history_deals ---

func TestGetHistoryDeals_success(t *testing.T) {
	want := []moomoo.Deal{
		{FillID: "789", Code: "US.AAPL", Qty: 10, Price: 150.5},
	}
	mock := &ordersMockClient{historyDealsResult: want}
	cs, cleanup := newOrdersServer(mock)
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_history_deals",
		Arguments: map[string]any{
			"account_id": "123",
			"trd_market": "US",
			"begin_time": "2026-06-01 00:00:00",
			"end_time":   "2026-07-01 00:00:00",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got Columnar
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got.Rows) != 1 || got.Rows[0][0] != "789" {
		t.Errorf("unexpected rows: %+v", got.Rows)
	}
}

func TestGetHistoryDeals_error(t *testing.T) {
	cs, cleanup := newOrdersServer(&ordersMockClient{historyDealsErr: errors.New("bad time range")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_history_deals",
		Arguments: map[string]any{
			"account_id": "123",
			"trd_market": "US",
			"begin_time": "2026-06-01 00:00:00",
			"end_time":   "2026-07-01 00:00:00",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}
