package moomoo

import (
	"context"
	"fmt"

	"github.com/hyperjiang/futu"
	"github.com/hyperjiang/futu/adapt"
	"github.com/hyperjiang/futu/pb/qotcommon"
)

// SessionData holds price/volume data for a single extended trading session
// (pre-market, after-market, or overnight).
type SessionData struct {
	Price      float64 `json:"price"`
	HighPrice  float64 `json:"high_price"`
	LowPrice   float64 `json:"low_price"`
	Volume     int64   `json:"volume"`
	Turnover   float64 `json:"turnover"`
	ChangeVal  float64 `json:"change_val"`
	ChangeRate float64 `json:"change_rate"`
}

// sessionFromProto converts a protobuf PreAfterMarketData to SessionData.
// Returns nil when the source is nil (session not available). market is the
// security's market prefix (e.g. "US"), used to decide rounding precision.
func (c *Client) sessionFromProto(p *qotcommon.PreAfterMarketData, market string) *SessionData {
	if p == nil {
		return nil
	}
	return &SessionData{
		Price:      c.roundPrice(p.GetPrice(), market),
		HighPrice:  c.roundPrice(p.GetHighPrice(), market),
		LowPrice:   c.roundPrice(p.GetLowPrice(), market),
		Volume:     p.GetVolume(),
		Turnover:   c.roundTurnover(p.GetTurnover()),
		ChangeVal:  c.roundPrice(p.GetChangeVal(), market),
		ChangeRate: c.roundRate(p.GetChangeRate()),
	}
}

// MarketSnapshot is a compact snapshot of a security's current market state.
type MarketSnapshot struct {
	Code        string       `json:"code"`
	Name        string       `json:"name,omitempty"`
	CurPrice    float64      `json:"cur_price"`
	OpenPrice   float64      `json:"open_price"`
	HighPrice   float64      `json:"high_price"`
	LowPrice    float64      `json:"low_price"`
	LastClose   float64      `json:"last_close"`
	Volume      int64        `json:"volume"`
	Turnover    float64      `json:"turnover"`
	UpdateTime  string       `json:"update_time,omitempty"`
	PreMarket   *SessionData `json:"pre_market,omitempty"`
	AfterMarket *SessionData `json:"after_market,omitempty"`
	Overnight   *SessionData `json:"overnight,omitempty"`
}

// Quote holds real-time basic quote data for a security.
type Quote struct {
	Code        string       `json:"code"`
	Name        string       `json:"name,omitempty"`
	CurPrice    float64      `json:"cur_price"`
	OpenPrice   float64      `json:"open_price"`
	HighPrice   float64      `json:"high_price"`
	LowPrice    float64      `json:"low_price"`
	LastClose   float64      `json:"last_close"`
	Volume      int64        `json:"volume"`
	Turnover    float64      `json:"turnover"`
	UpdateTime  string       `json:"update_time,omitempty"`
	PreMarket   *SessionData `json:"pre_market,omitempty"`
	AfterMarket *SessionData `json:"after_market,omitempty"`
	Overnight   *SessionData `json:"overnight,omitempty"`
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
	return call(c, ctx, func(ctx context.Context, sdk *futu.SDK) ([]MarketSnapshot, error) {
		snaps, err := sdk.GetSecuritySnapshotWithContext(ctx, codes)
		if err != nil {
			return nil, err
		}
		out := make([]MarketSnapshot, 0, len(snaps))
		for _, s := range snaps {
			b := s.GetBasic()
			if b == nil {
				continue
			}
			code := adapt.SecurityToCode(b.GetSecurity())
			market := marketFromCode(code)
			out = append(out, MarketSnapshot{
				Code:        code,
				Name:        b.GetName(),
				CurPrice:    c.roundPrice(b.GetCurPrice(), market),
				OpenPrice:   c.roundPrice(b.GetOpenPrice(), market),
				HighPrice:   c.roundPrice(b.GetHighPrice(), market),
				LowPrice:    c.roundPrice(b.GetLowPrice(), market),
				LastClose:   c.roundPrice(b.GetLastClosePrice(), market),
				Volume:      b.GetVolume(),
				Turnover:    c.roundTurnover(b.GetTurnover()),
				UpdateTime:  b.GetUpdateTime(),
				PreMarket:   c.sessionFromProto(b.GetPreMarket(), market),
				AfterMarket: c.sessionFromProto(b.GetAfterMarket(), market),
				Overnight:   c.sessionFromProto(b.GetOvernight(), market),
			})
		}
		return out, nil
	})
}

// GetQuote subscribes to basic quotes then returns the current quote for each security.
// Subscribe is required before GetBasicQot will return live data.
func (c *Client) GetQuote(ctx context.Context, codes []string) ([]Quote, error) {
	return call(c, ctx, func(ctx context.Context, sdk *futu.SDK) ([]Quote, error) {
		if err := sdk.SubscribeWithContext(ctx, codes, []int32{adapt.SubType_Basic}, true); err != nil {
			return nil, fmt.Errorf("subscribe: %w", err)
		}
		qots, err := sdk.GetBasicQotWithContext(ctx, codes)
		if err != nil {
			return nil, err
		}
		out := make([]Quote, 0, len(qots))
		for _, q := range qots {
			code := adapt.SecurityToCode(q.GetSecurity())
			market := marketFromCode(code)
			out = append(out, Quote{
				Code:        code,
				Name:        q.GetName(),
				CurPrice:    c.roundPrice(q.GetCurPrice(), market),
				OpenPrice:   c.roundPrice(q.GetOpenPrice(), market),
				HighPrice:   c.roundPrice(q.GetHighPrice(), market),
				LowPrice:    c.roundPrice(q.GetLowPrice(), market),
				LastClose:   c.roundPrice(q.GetLastClosePrice(), market),
				Volume:      q.GetVolume(),
				Turnover:    c.roundTurnover(q.GetTurnover()),
				UpdateTime:  q.GetUpdateTime(),
				PreMarket:   c.sessionFromProto(q.GetPreMarket(), market),
				AfterMarket: c.sessionFromProto(q.GetAfterMarket(), market),
				Overnight:   c.sessionFromProto(q.GetOvernight(), market),
			})
		}
		return out, nil
	})
}

// GetKlines requests historical K-line data for the given security.
// klType must be one of the adapt.KLType_* constants.
func (c *Client) GetKlines(ctx context.Context, code string, klType int32, beginTime, endTime string) ([]Kline, error) {
	return call(c, ctx, func(ctx context.Context, sdk *futu.SDK) ([]Kline, error) {
		s2c, err := sdk.RequestHistoryKLWithContext(ctx, code, klType, beginTime, endTime)
		if err != nil {
			return nil, err
		}
		market := marketFromCode(code)
		kls := s2c.GetKlList()
		out := make([]Kline, 0, len(kls))
		for _, kl := range kls {
			if kl.GetIsBlank() {
				continue
			}
			out = append(out, Kline{
				Time:       kl.GetTime(),
				Open:       c.roundPrice(kl.GetOpenPrice(), market),
				High:       c.roundPrice(kl.GetHighPrice(), market),
				Low:        c.roundPrice(kl.GetLowPrice(), market),
				Close:      c.roundPrice(kl.GetClosePrice(), market),
				Volume:     kl.GetVolume(),
				Turnover:   c.roundTurnover(kl.GetTurnover()),
				ChangeRate: c.roundRate(kl.GetChangeRate()),
			})
		}
		return out, nil
	})
}

// GetOrderBook subscribes to order-book data then returns bid/ask levels.
// Subscribe is required before GetOrderBook will return live data.
func (c *Client) GetOrderBook(ctx context.Context, code string) (*OrderBook, error) {
	return call(c, ctx, func(ctx context.Context, sdk *futu.SDK) (*OrderBook, error) {
		if err := sdk.SubscribeWithContext(ctx, []string{code}, []int32{adapt.SubType_OrderBook}, true); err != nil {
			return nil, fmt.Errorf("subscribe: %w", err)
		}
		s2c, err := sdk.GetOrderBookWithContext(ctx, code)
		if err != nil {
			return nil, err
		}
		market := marketFromCode(code)
		toEntries := func(list []*qotcommon.OrderBook) []OrderBookEntry {
			entries := make([]OrderBookEntry, 0, len(list))
			for _, e := range list {
				entries = append(entries, OrderBookEntry{
					Price:  c.roundPrice(e.GetPrice(), market),
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
	})
}
