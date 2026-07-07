package tools

import (
	"testing"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

func TestKlinesToColumnar(t *testing.T) {
	got, err := klinesToColumnar([]moomoo.Kline{
		{Time: "2024-01-01", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100, Turnover: 1000, ChangeRate: 0.01},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Default (fields omitted) is the lean 6-column set, dropping turnover/change_rate.
	if len(got.Columns) != 6 {
		t.Fatalf("want 6 default columns, got %d: %v", len(got.Columns), got.Columns)
	}
	if len(got.Rows) != 1 || len(got.Rows[0]) != 6 {
		t.Fatalf("unexpected rows: %+v", got.Rows)
	}
	if got.Rows[0][0] != "2024-01-01" || got.Rows[0][4] != 1.5 {
		t.Errorf("unexpected row values: %+v", got.Rows[0])
	}
}

func TestKlinesToColumnar_explicitFields(t *testing.T) {
	got, err := klinesToColumnar([]moomoo.Kline{
		{Time: "2024-01-01", Close: 1.5, Turnover: 1000, ChangeRate: 0.01},
	}, []string{"close", "turnover", "change_rate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantCols := []string{"close", "turnover", "change_rate"}
	if len(got.Columns) != len(wantCols) {
		t.Fatalf("want columns %v, got %v", wantCols, got.Columns)
	}
	if got.Rows[0][0] != 1.5 || got.Rows[0][1] != 1000.0 || got.Rows[0][2] != 0.01 {
		t.Errorf("unexpected row values: %+v", got.Rows[0])
	}
}

func TestKlinesToColumnar_unknownField(t *testing.T) {
	_, err := klinesToColumnar([]moomoo.Kline{{Time: "2024-01-01"}}, []string{"bogus"})
	if err == nil {
		t.Fatal("want error for unknown field")
	}
}

func TestOrdersToColumnar_empty(t *testing.T) {
	got, err := ordersToColumnar(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Rows) != 0 {
		t.Errorf("want 0 rows, got %d", len(got.Rows))
	}
	if len(got.Columns) == 0 {
		t.Error("want columns present even with no rows")
	}
}

func TestDealsToColumnar(t *testing.T) {
	got, err := dealsToColumnar([]moomoo.Deal{
		{FillID: "1", Code: "US.AAPL", Qty: 10, Price: 150.5},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Rows) != 1 || got.Rows[0][0] != "1" {
		t.Errorf("unexpected rows: %+v", got.Rows)
	}
}
