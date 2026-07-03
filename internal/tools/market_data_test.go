package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

// marketMockClient is a configurable mock for market data tests.
type marketMockClient struct {
	snapshotResult  []moomoo.MarketSnapshot
	snapshotErr     error
	quoteResult     []moomoo.Quote
	quoteErr        error
	klinesResult    []moomoo.Kline
	klinesErr       error
	orderBookResult *moomoo.OrderBook
	orderBookErr    error
}

func (m *marketMockClient) Health(_ context.Context) error { return nil }
func (m *marketMockClient) Close() error                   { return nil }
func (m *marketMockClient) GetSnapshot(_ context.Context, _ []string) ([]moomoo.MarketSnapshot, error) {
	return m.snapshotResult, m.snapshotErr
}
func (m *marketMockClient) GetQuote(_ context.Context, _ []string) ([]moomoo.Quote, error) {
	return m.quoteResult, m.quoteErr
}
func (m *marketMockClient) GetKlines(_ context.Context, _ string, _ int32, _, _ string) ([]moomoo.Kline, error) {
	return m.klinesResult, m.klinesErr
}
func (m *marketMockClient) GetOrderBook(_ context.Context, _ string) (*moomoo.OrderBook, error) {
	return m.orderBookResult, m.orderBookErr
}
func (m *marketMockClient) GetAccounts(_ context.Context) ([]moomoo.Account, error) {
	return nil, nil
}
func (m *marketMockClient) GetAssets(_ context.Context, _ uint64, _, _ string) (*moomoo.Assets, error) {
	return nil, nil
}
func (m *marketMockClient) GetPositions(_ context.Context, _ uint64, _, _ string) ([]moomoo.Position, error) {
	return nil, nil
}
func (m *marketMockClient) GetMaxTradable(_ context.Context, _ uint64, _, _, _, _ string, _ float64) (*moomoo.MaxTradable, error) {
	return nil, nil
}
func (m *marketMockClient) GetMarginRatio(_ context.Context, _ uint64, _, _ string, _ []string) ([]moomoo.MarginRatio, error) {
	return nil, nil
}
func (m *marketMockClient) GetCashFlow(_ context.Context, _ uint64, _, _, _ string) ([]moomoo.CashFlow, error) {
	return nil, nil
}
func (m *marketMockClient) GetUserSecurityGroup(_ context.Context, _ string) ([]moomoo.SecurityGroup, error) {
	return nil, nil
}
func (m *marketMockClient) GetUserSecurity(_ context.Context, _ string) ([]moomoo.WatchlistSecurity, error) {
	return nil, nil
}
func (m *marketMockClient) GetOrders(_ context.Context, _ uint64, _, _ string) ([]moomoo.Order, error) {
	return nil, nil
}
func (m *marketMockClient) GetDeals(_ context.Context, _ uint64, _, _ string) ([]moomoo.Deal, error) {
	return nil, nil
}
func (m *marketMockClient) GetHistoryOrders(_ context.Context, _ uint64, _, _, _, _ string, _ []string) ([]moomoo.Order, error) {
	return nil, nil
}
func (m *marketMockClient) GetHistoryDeals(_ context.Context, _ uint64, _, _, _, _ string, _ []string) ([]moomoo.Deal, error) {
	return nil, nil
}

func newMarketServer(c moomoo.MoomooClient) (*mcp.ClientSession, func()) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterMarketData(s, c)
	ctx := context.Background()
	cs, cleanup := connectTest(ctx, s)
	return cs, cleanup
}

// --- get_market_snapshot ---

func TestGetMarketSnapshot_success(t *testing.T) {
	want := []moomoo.MarketSnapshot{
		{Code: "US.AAPL", Name: "Apple", CurPrice: 200.5, Volume: 1000},
	}
	cs, cleanup := newMarketServer(&marketMockClient{snapshotResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_market_snapshot",
		Arguments: map[string]any{"codes": []string{"US.AAPL"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.MarketSnapshot
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].Code != "US.AAPL" || got[0].CurPrice != 200.5 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestGetMarketSnapshot_extendedSessions(t *testing.T) {
	want := []moomoo.MarketSnapshot{
		{
			Code:        "US.AAPL",
			CurPrice:    200.0,
			PreMarket:   &moomoo.SessionData{Price: 201.5, ChangeRate: 0.75},
			AfterMarket: &moomoo.SessionData{Price: 199.0, ChangeRate: -0.50},
			Overnight:   &moomoo.SessionData{Price: 198.5, ChangeRate: -0.75},
		},
	}
	cs, cleanup := newMarketServer(&marketMockClient{snapshotResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_market_snapshot",
		Arguments: map[string]any{"codes": []string{"US.AAPL"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.MarketSnapshot
	if err := json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &got); err != nil {
		t.Fatal(err)
	}
	g := got[0]
	if g.PreMarket == nil || g.PreMarket.Price != 201.5 {
		t.Errorf("unexpected pre_market: %+v", g.PreMarket)
	}
	if g.AfterMarket == nil || g.AfterMarket.Price != 199.0 {
		t.Errorf("unexpected after_market: %+v", g.AfterMarket)
	}
	if g.Overnight == nil || g.Overnight.Price != 198.5 {
		t.Errorf("unexpected overnight: %+v", g.Overnight)
	}
}

func TestGetMarketSnapshot_error(t *testing.T) {
	cs, cleanup := newMarketServer(&marketMockClient{snapshotErr: errors.New("network error")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_market_snapshot",
		Arguments: map[string]any{"codes": []string{"US.AAPL"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_stock_quote ---

func TestGetStockQuote_success(t *testing.T) {
	want := []moomoo.Quote{
		{Code: "HK.00700", Name: "Tencent", CurPrice: 400.0, Volume: 5000},
	}
	cs, cleanup := newMarketServer(&marketMockClient{quoteResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_stock_quote",
		Arguments: map[string]any{"codes": []string{"HK.00700"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.Quote
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].Code != "HK.00700" || got[0].CurPrice != 400.0 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestGetStockQuote_error(t *testing.T) {
	cs, cleanup := newMarketServer(&marketMockClient{quoteErr: errors.New("subscribe failed")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_stock_quote",
		Arguments: map[string]any{"codes": []string{"US.AAPL"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_historical_klines ---

func TestGetHistoricalKlines_success(t *testing.T) {
	want := []moomoo.Kline{
		{Time: "2024-01-02", Open: 185.0, High: 187.0, Low: 184.0, Close: 186.5, Volume: 80000000},
	}
	cs, cleanup := newMarketServer(&marketMockClient{klinesResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_historical_klines",
		Arguments: map[string]any{
			"code":       "US.AAPL",
			"kl_type":    "day",
			"begin_time": "2024-01-01",
			"end_time":   "2024-01-31",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.Kline
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].Time != "2024-01-02" || got[0].Close != 186.5 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestGetHistoricalKlines_invalidKLType(t *testing.T) {
	cs, cleanup := newMarketServer(&marketMockClient{})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_historical_klines",
		Arguments: map[string]any{
			"code":       "US.AAPL",
			"kl_type":    "invalid",
			"begin_time": "2024-01-01",
			"end_time":   "2024-01-31",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true for unknown kl_type")
	}
}

func TestGetHistoricalKlines_error(t *testing.T) {
	cs, cleanup := newMarketServer(&marketMockClient{klinesErr: errors.New("quota exceeded")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_historical_klines",
		Arguments: map[string]any{
			"code":       "US.AAPL",
			"kl_type":    "day",
			"begin_time": "2024-01-01",
			"end_time":   "2024-01-31",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_order_book ---

func TestGetOrderBook_success(t *testing.T) {
	want := &moomoo.OrderBook{
		Code: "US.AAPL",
		Ask:  []moomoo.OrderBookEntry{{Price: 200.5, Volume: 100, Count: 3}},
		Bid:  []moomoo.OrderBookEntry{{Price: 200.4, Volume: 200, Count: 5}},
	}
	cs, cleanup := newMarketServer(&marketMockClient{orderBookResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_order_book",
		Arguments: map[string]any{"code": "US.AAPL"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got moomoo.OrderBook
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if got.Code != "US.AAPL" {
		t.Errorf("want code US.AAPL, got %q", got.Code)
	}
	if len(got.Ask) != 1 || got.Ask[0].Price != 200.5 {
		t.Errorf("unexpected ask: %+v", got.Ask)
	}
	if len(got.Bid) != 1 || got.Bid[0].Price != 200.4 {
		t.Errorf("unexpected bid: %+v", got.Bid)
	}
}

func TestGetOrderBook_error(t *testing.T) {
	cs, cleanup := newMarketServer(&marketMockClient{orderBookErr: errors.New("subscribe failed")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_order_book",
		Arguments: map[string]any{"code": "US.AAPL"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}
