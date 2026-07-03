package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

// accountMockClient is a configurable mock for account tests.
type accountMockClient struct {
	accountsResult []moomoo.Account
	accountsErr    error

	assetsResult *moomoo.Assets
	assetsErr    error

	positionsResult []moomoo.Position
	positionsErr    error

	maxTradableResult *moomoo.MaxTradable
	maxTradableErr    error

	marginRatioResult []moomoo.MarginRatio
	marginRatioErr    error

	cashFlowResult []moomoo.CashFlow
	cashFlowErr    error

	// gotAccountID records the accountID last passed in, so tests can assert
	// it kept full uint64 precision through the JSON string round-trip.
	gotAccountID uint64
}

func (m *accountMockClient) Health(_ context.Context) error { return nil }
func (m *accountMockClient) Close() error                   { return nil }
func (m *accountMockClient) GetSnapshot(_ context.Context, _ []string) ([]moomoo.MarketSnapshot, error) {
	return nil, nil
}
func (m *accountMockClient) GetQuote(_ context.Context, _ []string) ([]moomoo.Quote, error) {
	return nil, nil
}
func (m *accountMockClient) GetKlines(_ context.Context, _ string, _ int32, _, _ string) ([]moomoo.Kline, error) {
	return nil, nil
}
func (m *accountMockClient) GetOrderBook(_ context.Context, _ string) (*moomoo.OrderBook, error) {
	return nil, nil
}
func (m *accountMockClient) GetAccounts(_ context.Context) ([]moomoo.Account, error) {
	return m.accountsResult, m.accountsErr
}
func (m *accountMockClient) GetAssets(_ context.Context, accountID uint64, _, _ string) (*moomoo.Assets, error) {
	m.gotAccountID = accountID
	return m.assetsResult, m.assetsErr
}
func (m *accountMockClient) GetPositions(_ context.Context, accountID uint64, _, _ string) ([]moomoo.Position, error) {
	m.gotAccountID = accountID
	return m.positionsResult, m.positionsErr
}
func (m *accountMockClient) GetMaxTradable(_ context.Context, accountID uint64, _, _, _, _ string, _ float64) (*moomoo.MaxTradable, error) {
	m.gotAccountID = accountID
	return m.maxTradableResult, m.maxTradableErr
}
func (m *accountMockClient) GetMarginRatio(_ context.Context, accountID uint64, _, _ string, _ []string) ([]moomoo.MarginRatio, error) {
	m.gotAccountID = accountID
	return m.marginRatioResult, m.marginRatioErr
}
func (m *accountMockClient) GetCashFlow(_ context.Context, accountID uint64, _, _, _ string) ([]moomoo.CashFlow, error) {
	m.gotAccountID = accountID
	return m.cashFlowResult, m.cashFlowErr
}
func (m *accountMockClient) GetUserSecurityGroup(_ context.Context, _ string) ([]moomoo.SecurityGroup, error) {
	return nil, nil
}
func (m *accountMockClient) GetUserSecurity(_ context.Context, _ string) ([]moomoo.WatchlistSecurity, error) {
	return nil, nil
}

func newAccountServer(c moomoo.MoomooClient) (*mcp.ClientSession, func()) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterAccount(s, c)
	ctx := context.Background()
	cs, cleanup := connectTest(ctx, s)
	return cs, cleanup
}

// bigAccountID exceeds 2^53, the largest integer a float64 (and therefore a
// JSON number decoded by most non-Go clients) can represent exactly.
const bigAccountID = "9223372036854775807" // math.MaxInt64

// --- get_accounts ---

func TestGetAccounts_success(t *testing.T) {
	want := []moomoo.Account{
		{AccountID: bigAccountID, TrdEnv: "SIMULATE", TrdMarkets: []string{"US"}, AccType: "CASH"},
	}
	cs, cleanup := newAccountServer(&accountMockClient{accountsResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_accounts"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.Account
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].AccountID != bigAccountID {
		t.Errorf("unexpected result: %+v", got)
	}
	// account_id must be a JSON string, not a number, to preserve precision.
	if !json.Valid([]byte(text)) {
		t.Fatal("invalid JSON")
	}
	if want := `"account_id":"` + bigAccountID + `"`; !strings.Contains(text, want) {
		t.Errorf("expected account_id to be encoded as a string %q in raw JSON, got: %s", want, text)
	}
}

func TestGetAccounts_error(t *testing.T) {
	cs, cleanup := newAccountServer(&accountMockClient{accountsErr: errors.New("not connected")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_accounts"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_assets ---

func TestGetAssets_success(t *testing.T) {
	want := &moomoo.Assets{Power: 1000, TotalAssets: 5000, Cash: 3000}
	mock := &accountMockClient{assetsResult: want}
	cs, cleanup := newAccountServer(mock)
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_assets",
		Arguments: map[string]any{
			"account_id": bigAccountID,
			"trd_market": "US",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got moomoo.Assets
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if got.Cash != 3000 {
		t.Errorf("unexpected result: %+v", got)
	}
	if mock.gotAccountID != 9223372036854775807 {
		t.Errorf("account_id lost precision: got %d", mock.gotAccountID)
	}
}

func TestGetAssets_invalidAccountID(t *testing.T) {
	cs, cleanup := newAccountServer(&accountMockClient{})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_assets",
		Arguments: map[string]any{
			"account_id": "not-a-number",
			"trd_market": "US",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true for invalid account_id")
	}
}

func TestGetAssets_error(t *testing.T) {
	cs, cleanup := newAccountServer(&accountMockClient{assetsErr: errors.New("permission denied")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_assets",
		Arguments: map[string]any{
			"account_id": "12345",
			"trd_market": "US",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_positions ---

func TestGetPositions_success(t *testing.T) {
	want := []moomoo.Position{
		{PositionID: bigAccountID, PositionSide: "LONG", Code: "US.AAPL", Qty: 10},
	}
	cs, cleanup := newAccountServer(&accountMockClient{positionsResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_positions",
		Arguments: map[string]any{
			"account_id": "12345",
			"trd_market": "US",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.Position
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].Code != "US.AAPL" || got[0].PositionID != bigAccountID {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestGetPositions_error(t *testing.T) {
	cs, cleanup := newAccountServer(&accountMockClient{positionsErr: errors.New("timeout")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_positions",
		Arguments: map[string]any{
			"account_id": "12345",
			"trd_market": "US",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_account_summary ---

func TestGetAccountSummary_success(t *testing.T) {
	mock := &accountMockClient{
		assetsResult:    &moomoo.Assets{Cash: 100},
		positionsResult: []moomoo.Position{{Code: "US.AAPL", Qty: 1}},
	}
	cs, cleanup := newAccountServer(mock)
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_account_summary",
		Arguments: map[string]any{
			"account_id": "12345",
			"trd_market": "US",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got moomoo.AccountSummary
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if got.Assets == nil || got.Assets.Cash != 100 {
		t.Errorf("unexpected assets: %+v", got.Assets)
	}
	if len(got.Positions) != 1 || got.Positions[0].Code != "US.AAPL" {
		t.Errorf("unexpected positions: %+v", got.Positions)
	}
}

func TestGetAccountSummary_assetsError(t *testing.T) {
	cs, cleanup := newAccountServer(&accountMockClient{assetsErr: errors.New("boom")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_account_summary",
		Arguments: map[string]any{
			"account_id": "12345",
			"trd_market": "US",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_max_tradable ---

func TestGetMaxTradable_success(t *testing.T) {
	want := &moomoo.MaxTradable{MaxCashBuy: 100, MaxPositionSell: 10}
	cs, cleanup := newAccountServer(&accountMockClient{maxTradableResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_max_tradable",
		Arguments: map[string]any{
			"account_id": "12345",
			"trd_market": "US",
			"order_type": "NORMAL",
			"code":       "US.AAPL",
			"price":      200.0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got moomoo.MaxTradable
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if got.MaxCashBuy != 100 || got.MaxPositionSell != 10 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestGetMaxTradable_error(t *testing.T) {
	cs, cleanup := newAccountServer(&accountMockClient{maxTradableErr: errors.New("no quote")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_max_tradable",
		Arguments: map[string]any{
			"account_id": "12345",
			"trd_market": "US",
			"order_type": "NORMAL",
			"code":       "US.AAPL",
			"price":      200.0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_margin_ratio ---

func TestGetMarginRatio_success(t *testing.T) {
	want := []moomoo.MarginRatio{
		{Code: "US.AAPL", IsLongPermit: true, IsShortPermit: false},
	}
	cs, cleanup := newAccountServer(&accountMockClient{marginRatioResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_margin_ratio",
		Arguments: map[string]any{
			"account_id": "12345",
			"trd_market": "US",
			"codes":      []string{"US.AAPL"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.MarginRatio
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || !got[0].IsLongPermit {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestGetMarginRatio_error(t *testing.T) {
	cs, cleanup := newAccountServer(&accountMockClient{marginRatioErr: errors.New("no permission")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_margin_ratio",
		Arguments: map[string]any{
			"account_id": "12345",
			"trd_market": "US",
			"codes":      []string{"US.AAPL"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}

// --- get_cash_flow ---

func TestGetCashFlow_success(t *testing.T) {
	want := []moomoo.CashFlow{
		{ClearingDate: "20240101", Direction: "IN", Amount: 500},
	}
	cs, cleanup := newAccountServer(&accountMockClient{cashFlowResult: want})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_cash_flow",
		Arguments: map[string]any{
			"account_id":    "12345",
			"trd_market":    "US",
			"clearing_date": "20240101",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	var got []moomoo.CashFlow
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if len(got) != 1 || got[0].Amount != 500 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestGetCashFlow_error(t *testing.T) {
	cs, cleanup := newAccountServer(&accountMockClient{cashFlowErr: errors.New("bad date")})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_cash_flow",
		Arguments: map[string]any{
			"account_id":    "12345",
			"trd_market":    "US",
			"clearing_date": "20240101",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("want IsError=true")
	}
}
