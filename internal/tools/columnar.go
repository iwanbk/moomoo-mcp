package tools

import (
	"fmt"
	"strings"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

// Columnar is a header + array-of-arrays representation used for tools that
// can return many rows, so field names are sent once instead of once per
// row.
type Columnar struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

var (
	klineColumns        = []string{"time", "open", "high", "low", "close", "volume", "turnover", "change_rate"}
	klineDefaultColumns = []string{"time", "open", "high", "low", "close", "volume"}

	orderColumns = []string{
		"order_id", "order_id_ex", "code", "name", "trd_side", "order_type", "order_status",
		"qty", "price", "fill_qty", "fill_avg_price", "create_time", "update_time", "remark", "last_err_msg",
	}
	orderDefaultColumns = []string{
		"order_id", "code", "name", "trd_side", "order_type", "order_status",
		"qty", "price", "fill_qty", "fill_avg_price", "create_time", "update_time",
	}

	dealColumns = []string{
		"fill_id", "fill_id_ex", "order_id", "order_id_ex", "code", "name", "trd_side",
		"qty", "price", "create_time", "status", "counter_broker_name",
	}
	dealDefaultColumns = []string{
		"fill_id", "order_id", "code", "name", "trd_side", "qty", "price", "create_time", "status",
	}
)

func klinesToColumnar(kls []moomoo.Kline, fields []string) (Columnar, error) {
	rows := make([][]any, len(kls))
	for i, k := range kls {
		rows[i] = []any{k.Time, k.Open, k.High, k.Low, k.Close, k.Volume, k.Turnover, k.ChangeRate}
	}
	return filterColumns(klineColumns, klineDefaultColumns, rows, fields)
}

func ordersToColumnar(orders []moomoo.Order, fields []string) (Columnar, error) {
	rows := make([][]any, len(orders))
	for i, o := range orders {
		rows[i] = []any{
			o.OrderID, o.OrderIDEx, o.Code, o.Name, o.TrdSide, o.OrderType, o.OrderStatus,
			o.Qty, o.Price, o.FillQty, o.FillAvgPrice, o.CreateTime, o.UpdateTime, o.Remark, o.LastErrMsg,
		}
	}
	return filterColumns(orderColumns, orderDefaultColumns, rows, fields)
}

func dealsToColumnar(deals []moomoo.Deal, fields []string) (Columnar, error) {
	rows := make([][]any, len(deals))
	for i, d := range deals {
		rows[i] = []any{
			d.FillID, d.FillIDEx, d.OrderID, d.OrderIDEx, d.Code, d.Name, d.TrdSide,
			d.Qty, d.Price, d.CreateTime, d.Status, d.CounterBrokerName,
		}
	}
	return filterColumns(dealColumns, dealDefaultColumns, rows, fields)
}

// filterColumns builds a Columnar restricted to the requested fields (in the
// order given), reading each row's values by position from fullColumns. If
// fields is empty, defaultColumns is used instead, so omitting the argument
// yields a lean response rather than every column.
func filterColumns(fullColumns, defaultColumns []string, fullRows [][]any, fields []string) (Columnar, error) {
	cols := defaultColumns
	if len(fields) > 0 {
		cols = fields
	}
	index := make(map[string]int, len(fullColumns))
	for i, c := range fullColumns {
		index[c] = i
	}
	positions := make([]int, len(cols))
	for i, c := range cols {
		p, ok := index[c]
		if !ok {
			return Columnar{}, fmt.Errorf("unknown field %q; valid fields: %s", c, strings.Join(fullColumns, ", "))
		}
		positions[i] = p
	}
	rows := make([][]any, len(fullRows))
	for i, r := range fullRows {
		row := make([]any, len(positions))
		for j, p := range positions {
			row[j] = r[p]
		}
		rows[i] = row
	}
	return Columnar{Columns: cols, Rows: rows}, nil
}
