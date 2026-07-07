package moomoo

import (
	"context"
	"strconv"

	"github.com/hyperjiang/futu"
	"github.com/hyperjiang/futu/adapt"
	"github.com/hyperjiang/futu/pb/trdcommon"
	"google.golang.org/protobuf/proto"
)

// Order is a single order returned by the open/history order list APIs.
type Order struct {
	OrderID      string  `json:"order_id"`
	OrderIDEx    string  `json:"order_id_ex,omitempty"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	TrdSide      string  `json:"trd_side"`
	OrderType    string  `json:"order_type"`
	OrderStatus  string  `json:"order_status"`
	Qty          float64 `json:"qty"`
	Price        float64 `json:"price,omitempty"`
	FillQty      float64 `json:"fill_qty,omitempty"`
	FillAvgPrice float64 `json:"fill_avg_price,omitempty"`
	CreateTime   string  `json:"create_time"`
	UpdateTime   string  `json:"update_time"`
	Remark       string  `json:"remark,omitempty"`
	LastErrMsg   string  `json:"last_err_msg,omitempty"`
}

// Deal is a single filled order (order fill) returned by the deal/history
// deal list APIs.
type Deal struct {
	FillID            string  `json:"fill_id"`
	FillIDEx          string  `json:"fill_id_ex,omitempty"`
	OrderID           string  `json:"order_id,omitempty"`
	OrderIDEx         string  `json:"order_id_ex,omitempty"`
	Code              string  `json:"code"`
	Name              string  `json:"name"`
	TrdSide           string  `json:"trd_side"`
	Qty               float64 `json:"qty"`
	Price             float64 `json:"price"`
	CreateTime        string  `json:"create_time"`
	Status            string  `json:"status,omitempty"`
	CounterBrokerName string  `json:"counter_broker_name,omitempty"`
}

var trdSideNames = map[int32]string{
	adapt.TrdSide_Unknown: "UNKNOWN",
	adapt.TrdSide_Buy:     "BUY",
	adapt.TrdSide_Sell:    "SELL",
}

var orderStatusNames = map[int32]string{
	adapt.OrderStatus_Unsubmitted:     "UNSUBMITTED",
	adapt.OrderStatus_Unknown:         "UNKNOWN",
	adapt.OrderStatus_WaitingSubmit:   "WAITING_SUBMIT",
	adapt.OrderStatus_Submitting:      "SUBMITTING",
	adapt.OrderStatus_SubmitFailed:    "SUBMIT_FAILED",
	adapt.OrderStatus_TimeOut:         "TIMEOUT",
	adapt.OrderStatus_Submitted:       "SUBMITTED",
	adapt.OrderStatus_Filled_Part:     "FILLED_PART",
	adapt.OrderStatus_Filled_All:      "FILLED_ALL",
	adapt.OrderStatus_Cancelling_Part: "CANCELLING_PART",
	adapt.OrderStatus_Cancelling_All:  "CANCELLING_ALL",
	adapt.OrderStatus_Cancelled_Part:  "CANCELLED_PART",
	adapt.OrderStatus_Cancelled_All:   "CANCELLED_ALL",
	adapt.OrderStatus_Failed:          "FAILED",
	adapt.OrderStatus_Disabled:        "DISABLED",
	adapt.OrderStatus_Deleted:         "DELETED",
	adapt.OrderStatus_FillCancelled:   "FILL_CANCELLED",
}

var orderFillStatusNames = map[int32]string{
	int32(trdcommon.OrderFillStatus_OrderFillStatus_OK):        "OK",
	int32(trdcommon.OrderFillStatus_OrderFillStatus_Cancelled): "CANCELLED",
	int32(trdcommon.OrderFillStatus_OrderFillStatus_Changed):   "CHANGED",
}

// orderTypeNames is the reverse of orderTypeIDs (defined in trade.go), used
// to translate an Order's raw orderType back to the friendly name a caller
// would pass as order_type to get_max_tradable.
var orderTypeNames = func() map[int32]string {
	m := make(map[int32]string, len(orderTypeIDs))
	for name, id := range orderTypeIDs {
		m[id] = name
	}
	return m
}()

func toOrder(o *trdcommon.Order) Order {
	return Order{
		OrderID:      strconv.FormatUint(o.GetOrderID(), 10),
		OrderIDEx:    o.GetOrderIDEx(),
		Code:         o.GetCode(),
		Name:         o.GetName(),
		TrdSide:      trdSideNames[o.GetTrdSide()],
		OrderType:    orderTypeNames[o.GetOrderType()],
		OrderStatus:  orderStatusNames[o.GetOrderStatus()],
		Qty:          o.GetQty(),
		Price:        o.GetPrice(),
		FillQty:      o.GetFillQty(),
		FillAvgPrice: o.GetFillAvgPrice(),
		CreateTime:   o.GetCreateTime(),
		UpdateTime:   o.GetUpdateTime(),
		Remark:       o.GetRemark(),
		LastErrMsg:   o.GetLastErrMsg(),
	}
}

func toDeal(f *trdcommon.OrderFill) Deal {
	return Deal{
		FillID:            strconv.FormatUint(f.GetFillID(), 10),
		FillIDEx:          f.GetFillIDEx(),
		OrderID:           strconv.FormatUint(f.GetOrderID(), 10),
		OrderIDEx:         f.GetOrderIDEx(),
		Code:              f.GetCode(),
		Name:              f.GetName(),
		TrdSide:           trdSideNames[f.GetTrdSide()],
		Qty:               f.GetQty(),
		Price:             f.GetPrice(),
		CreateTime:        f.GetCreateTime(),
		Status:            orderFillStatusNames[f.GetStatus()],
		CounterBrokerName: f.GetCounterBrokerName(),
	}
}

// GetOrders returns today's open orders for one account.
func (c *Client) GetOrders(ctx context.Context, accountID uint64, trdEnv, trdMarket string) ([]Order, error) {
	header, err := c.tradeHeader(accountID, trdEnv, trdMarket)
	if err != nil {
		return nil, err
	}
	return call(c, ctx, func(ctx context.Context, sdk *futu.SDK) ([]Order, error) {
		orders, err := sdk.GetOpenOrderListWithContext(ctx, header)
		if err != nil {
			return nil, err
		}
		out := make([]Order, 0, len(orders))
		for _, o := range orders {
			out = append(out, toOrder(o))
		}
		return out, nil
	})
}

// GetDeals returns today's filled deals (order fills) for one account.
func (c *Client) GetDeals(ctx context.Context, accountID uint64, trdEnv, trdMarket string) ([]Deal, error) {
	header, err := c.tradeHeader(accountID, trdEnv, trdMarket)
	if err != nil {
		return nil, err
	}
	return call(c, ctx, func(ctx context.Context, sdk *futu.SDK) ([]Deal, error) {
		fills, err := sdk.GetOrderFillListWithContext(ctx, header)
		if err != nil {
			return nil, err
		}
		out := make([]Deal, 0, len(fills))
		for _, f := range fills {
			out = append(out, toDeal(f))
		}
		return out, nil
	})
}

// GetHistoryOrders returns historical orders for one account within
// [beginTime, endTime] (format yyyy-MM-dd HH:mm:ss). codes optionally
// restricts the result to specific security codes.
func (c *Client) GetHistoryOrders(ctx context.Context, accountID uint64, trdEnv, trdMarket, beginTime, endTime string, codes []string) ([]Order, error) {
	header, err := c.tradeHeader(accountID, trdEnv, trdMarket)
	if err != nil {
		return nil, err
	}
	fc := &trdcommon.TrdFilterConditions{
		BeginTime: proto.String(beginTime),
		EndTime:   proto.String(endTime),
		CodeList:  codes,
	}
	return call(c, ctx, func(ctx context.Context, sdk *futu.SDK) ([]Order, error) {
		orders, err := sdk.GetHistoryOrderListWithContext(ctx, header, fc)
		if err != nil {
			return nil, err
		}
		out := make([]Order, 0, len(orders))
		for _, o := range orders {
			out = append(out, toOrder(o))
		}
		return out, nil
	})
}

// GetHistoryDeals returns historical filled deals for one account within
// [beginTime, endTime] (format yyyy-MM-dd HH:mm:ss). codes optionally
// restricts the result to specific security codes.
func (c *Client) GetHistoryDeals(ctx context.Context, accountID uint64, trdEnv, trdMarket, beginTime, endTime string, codes []string) ([]Deal, error) {
	header, err := c.tradeHeader(accountID, trdEnv, trdMarket)
	if err != nil {
		return nil, err
	}
	fc := &trdcommon.TrdFilterConditions{
		BeginTime: proto.String(beginTime),
		EndTime:   proto.String(endTime),
		CodeList:  codes,
	}
	return call(c, ctx, func(ctx context.Context, sdk *futu.SDK) ([]Deal, error) {
		fills, err := sdk.GetHistoryOrderFillListWithContext(ctx, header, fc)
		if err != nil {
			return nil, err
		}
		out := make([]Deal, 0, len(fills))
		for _, f := range fills {
			out = append(out, toDeal(f))
		}
		return out, nil
	})
}
