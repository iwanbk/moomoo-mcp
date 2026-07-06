package tools

import "github.com/iwanbk/moomoo-mcp/internal/moomoo"

// Columnar is a header + array-of-arrays representation used for tools that
// can return many rows, so field names are sent once instead of once per
// row.
type Columnar struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

func klinesToColumnar(kls []moomoo.Kline) Columnar {
	rows := make([][]any, len(kls))
	for i, k := range kls {
		rows[i] = []any{k.Time, k.Open, k.High, k.Low, k.Close, k.Volume, k.Turnover, k.ChangeRate}
	}
	return Columnar{
		Columns: []string{"time", "open", "high", "low", "close", "volume", "turnover", "change_rate"},
		Rows:    rows,
	}
}

func ordersToColumnar(orders []moomoo.Order) Columnar {
	rows := make([][]any, len(orders))
	for i, o := range orders {
		rows[i] = []any{
			o.OrderID, o.OrderIDEx, o.Code, o.Name, o.TrdSide, o.OrderType, o.OrderStatus,
			o.Qty, o.Price, o.FillQty, o.FillAvgPrice, o.CreateTime, o.UpdateTime, o.Remark, o.LastErrMsg,
		}
	}
	return Columnar{
		Columns: []string{
			"order_id", "order_id_ex", "code", "name", "trd_side", "order_type", "order_status",
			"qty", "price", "fill_qty", "fill_avg_price", "create_time", "update_time", "remark", "last_err_msg",
		},
		Rows: rows,
	}
}

func dealsToColumnar(deals []moomoo.Deal) Columnar {
	rows := make([][]any, len(deals))
	for i, d := range deals {
		rows[i] = []any{
			d.FillID, d.FillIDEx, d.OrderID, d.OrderIDEx, d.Code, d.Name, d.TrdSide,
			d.Qty, d.Price, d.CreateTime, d.Status, d.CounterBrokerName,
		}
	}
	return Columnar{
		Columns: []string{
			"fill_id", "fill_id_ex", "order_id", "order_id_ex", "code", "name", "trd_side",
			"qty", "price", "create_time", "status", "counter_broker_name",
		},
		Rows: rows,
	}
}
