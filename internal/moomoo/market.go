package moomoo

import (
	"context"
	"fmt"

	"github.com/hyperjiang/futu/adapt"
	"github.com/hyperjiang/futu/pb/qotcommon"
)

// MarketSnapshot is a compact snapshot of a security's current market state.
type MarketSnapshot struct {
	Code       string  `json:"code"`
	Name       string  `json:"name,omitempty"`
	CurPrice   float64 `json:"cur_price"`
	OpenPrice  float64 `json:"open_price"`
	HighPrice  float64 `json:"high_price"`
	LowPrice   float64 `json:"low_price"`
	LastClose  float64 `json:"last_close"`
	Volume     int64   `json:"volume"`
	Turnover   float64 `json:"turnover"`
	UpdateTime string  `json:"update_time,omitempty"`
}

// Quote holds real-time basic quote data for a security.
type Quote struct {
	Code       string  `json:"code"`
	Name       string  `json:"name,omitempty"`
	CurPrice   float64 `json:"cur_price"`
	OpenPrice  float64 `json:"open_price"`
	HighPrice  float64 `json:"high_price"`
	LowPrice   float64 `json:"low_price"`
	LastClose  float64 `json:"last_close"`
	Volume     int64   `json:"volume"`
	Turnover   float64 `json:"turnover"`
	UpdateTime string  `json:"update_time,omitempty"`
}

// Kline is a single OHLCV candle.
type Kline struct {
	Time       string  `json:"time"`
	Open       float64 `json:"open"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Close      float64 `json:"close"`
	Volume     int64   `json:"volume"`
	Turnover   float64 `json:"turnover"`
	ChangeRate float64 `json:"change_rate"`
}

// OrderBookEntry is one price level in the order book.
type OrderBookEntry struct {
	Price  float64 `json:"price"`
	Volume int64   `json:"volume"`
	Count  int32   `json:"order_count"`
}

// OrderBook holds bid/ask levels for a security.
type OrderBook struct {
	Code string           `json:"code"`
	Ask  []OrderBookEntry `json:"ask"`
	Bid  []OrderBookEntry `json:"bid"`
}

// GetSnapshot returns a compact snapshot for each requested security code.
func (c *Client) GetSnapshot(ctx context.Context, codes []string) ([]MarketSnapshot, error) {
	snaps, err := c.sdk.GetSecuritySnapshotWithContext(ctx, codes)
	if err != nil {
		return nil, err
	}
	out := make([]MarketSnapshot, 0, len(snaps))
	for _, s := range snaps {
		b := s.GetBasic()
		if b == nil {
			continue
		}
		out = append(out, MarketSnapshot{
			Code:       adapt.SecurityToCode(b.GetSecurity()),
			Name:       b.GetName(),
			CurPrice:   b.GetCurPrice(),
			OpenPrice:  b.GetOpenPrice(),
			HighPrice:  b.GetHighPrice(),
			LowPrice:   b.GetLowPrice(),
			LastClose:  b.GetLastClosePrice(),
			Volume:     b.GetVolume(),
			Turnover:   b.GetTurnover(),
			UpdateTime: b.GetUpdateTime(),
		})
	}
	return out, nil
}

// GetQuote subscribes to basic quotes then returns the current quote for each security.
// Subscribe is required before GetBasicQot will return live data.
func (c *Client) GetQuote(ctx context.Context, codes []string) ([]Quote, error) {
	if err := c.sdk.SubscribeWithContext(ctx, codes, []int32{adapt.SubType_Basic}, true); err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	qots, err := c.sdk.GetBasicQotWithContext(ctx, codes)
	if err != nil {
		return nil, err
	}
	out := make([]Quote, 0, len(qots))
	for _, q := range qots {
		out = append(out, Quote{
			Code:       adapt.SecurityToCode(q.GetSecurity()),
			Name:       q.GetName(),
			CurPrice:   q.GetCurPrice(),
			OpenPrice:  q.GetOpenPrice(),
			HighPrice:  q.GetHighPrice(),
			LowPrice:   q.GetLowPrice(),
			LastClose:  q.GetLastClosePrice(),
			Volume:     q.GetVolume(),
			Turnover:   q.GetTurnover(),
			UpdateTime: q.GetUpdateTime(),
		})
	}
	return out, nil
}

// GetKlines requests historical K-line data for the given security.
// klType must be one of the adapt.KLType_* constants.
func (c *Client) GetKlines(ctx context.Context, code string, klType int32, beginTime, endTime string) ([]Kline, error) {
	s2c, err := c.sdk.RequestHistoryKLWithContext(ctx, code, klType, beginTime, endTime)
	if err != nil {
		return nil, err
	}
	kls := s2c.GetKlList()
	out := make([]Kline, 0, len(kls))
	for _, kl := range kls {
		if kl.GetIsBlank() {
			continue
		}
		out = append(out, Kline{
			Time:       kl.GetTime(),
			Open:       kl.GetOpenPrice(),
			High:       kl.GetHighPrice(),
			Low:        kl.GetLowPrice(),
			Close:      kl.GetClosePrice(),
			Volume:     kl.GetVolume(),
			Turnover:   kl.GetTurnover(),
			ChangeRate: kl.GetChangeRate(),
		})
	}
	return out, nil
}

// GetOrderBook subscribes to order-book data then returns bid/ask levels.
// Subscribe is required before GetOrderBook will return live data.
func (c *Client) GetOrderBook(ctx context.Context, code string) (*OrderBook, error) {
	if err := c.sdk.SubscribeWithContext(ctx, []string{code}, []int32{adapt.SubType_OrderBook}, true); err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	s2c, err := c.sdk.GetOrderBookWithContext(ctx, code)
	if err != nil {
		return nil, err
	}
	toEntries := func(list []*qotcommon.OrderBook) []OrderBookEntry {
		entries := make([]OrderBookEntry, 0, len(list))
		for _, e := range list {
			entries = append(entries, OrderBookEntry{
				Price:  e.GetPrice(),
				Volume: e.GetVolume(),
				Count:  e.GetOrederCount(),
			})
		}
		return entries
	}
	return &OrderBook{
		Code: adapt.SecurityToCode(s2c.GetSecurity()),
		Ask:  toEntries(s2c.GetOrderBookAskList()),
		Bid:  toEntries(s2c.GetOrderBookBidList()),
	}, nil
}
