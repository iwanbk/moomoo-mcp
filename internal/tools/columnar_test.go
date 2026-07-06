package tools

import (
	"testing"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

func TestKlinesToColumnar(t *testing.T) {
	got := klinesToColumnar([]moomoo.Kline{
		{Time: "2024-01-01", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100, Turnover: 1000, ChangeRate: 0.01},
	})
	if len(got.Columns) != 8 {
		t.Fatalf("want 8 columns, got %d: %v", len(got.Columns), got.Columns)
	}
	if len(got.Rows) != 1 || len(got.Rows[0]) != 8 {
		t.Fatalf("unexpected rows: %+v", got.Rows)
	}
	if got.Rows[0][0] != "2024-01-01" || got.Rows[0][4] != 1.5 {
		t.Errorf("unexpected row values: %+v", got.Rows[0])
	}
}

func TestOrdersToColumnar_empty(t *testing.T) {
	got := ordersToColumnar(nil)
	if len(got.Rows) != 0 {
		t.Errorf("want 0 rows, got %d", len(got.Rows))
	}
	if len(got.Columns) == 0 {
		t.Error("want columns present even with no rows")
	}
}

func TestDealsToColumnar(t *testing.T) {
	got := dealsToColumnar([]moomoo.Deal{
		{FillID: "1", Code: "US.AAPL", Qty: 10, Price: 150.5},
	})
	if len(got.Rows) != 1 || got.Rows[0][0] != "1" {
		t.Errorf("unexpected rows: %+v", got.Rows)
	}
}
