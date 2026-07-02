package moomoo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperjiang/futu/adapt"
	"github.com/hyperjiang/futu/pb/trdcommon"
	"github.com/hyperjiang/futu/pb/trdflowsummary"
)

// Account is a trading account returned by GetAccList.
// AccountID is a decimal string (not a JSON number) so large account IDs
// never lose precision when decoded by clients that use float64 for numbers.
type Account struct {
	AccountID    string   `json:"account_id"`
	TrdEnv       string   `json:"trd_env"`
	TrdMarkets   []string `json:"trd_markets"`
	AccType      string   `json:"acc_type,omitempty"`
	SecurityFirm string   `json:"security_firm,omitempty"`
}

// Assets holds account funds/asset information.
type Assets struct {
	Power                 float64 `json:"power"`
	TotalAssets           float64 `json:"total_assets"`
	Cash                  float64 `json:"cash"`
	MarketValue           float64 `json:"market_value"`
	FrozenCash            float64 `json:"frozen_cash"`
	DebtCash              float64 `json:"debt_cash"`
	AvailableWithdrawCash float64 `json:"available_withdraw_cash"`
	RiskLevel             int32   `json:"risk_level,omitempty"`
}

// Position is a single open position.
type Position struct {
	PositionID   string  `json:"position_id"`
	PositionSide string  `json:"position_side"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Qty          float64 `json:"qty"`
	CanSellQty   float64 `json:"can_sell_qty"`
	Price        float64 `json:"price"`
	CostPrice    float64 `json:"cost_price,omitempty"`
	Value        float64 `json:"value"`
	PlValue      float64 `json:"pl_value"`
	PlRatio      float64 `json:"pl_ratio,omitempty"`
}

// AccountSummary is a composite of Assets and Position for one account.
type AccountSummary struct {
	Assets    *Assets    `json:"assets"`
	Positions []Position `json:"positions"`
}

// MaxTradable holds the maximum tradable quantities for a security.
type MaxTradable struct {
	MaxCashBuy          float64 `json:"max_cash_buy"`
	MaxCashAndMarginBuy float64 `json:"max_cash_and_margin_buy,omitempty"`
	MaxPositionSell     float64 `json:"max_position_sell"`
	MaxSellShort        float64 `json:"max_sell_short,omitempty"`
	MaxBuyBack          float64 `json:"max_buy_back,omitempty"`
}

// MarginRatio holds margin ratio info for a security.
type MarginRatio struct {
	Code            string  `json:"code"`
	IsLongPermit    bool    `json:"is_long_permit"`
	IsShortPermit   bool    `json:"is_short_permit"`
	ShortPoolRemain float64 `json:"short_pool_remain,omitempty"`
	ShortFeeRate    float64 `json:"short_fee_rate,omitempty"`
}

// CashFlow is a single entry from the trading flow summary.
type CashFlow struct {
	ClearingDate string  `json:"clearing_date,omitempty"`
	CashFlowType string  `json:"cash_flow_type,omitempty"`
	Direction    string  `json:"direction,omitempty"`
	Amount       float64 `json:"amount"`
	Remark       string  `json:"remark,omitempty"`
}

// trdMarketNames maps trdcommon.TrdMarket values to human-readable names.
// Values returned here in Account.TrdMarkets are meant to be echoed back as
// the trd_market argument to the other account tools.
var trdMarketNames = map[int32]string{
	adapt.TrdMarket_Unknown:             "UNKNOWN",
	adapt.TrdMarket_HK:                  "HK",
	adapt.TrdMarket_US:                  "US",
	adapt.TrdMarket_CN:                  "CN",
	adapt.TrdMarket_HKCC:                "HKCC",
	adapt.TrdMarket_Futures:             "FUTURES",
	adapt.TrdMarket_SG:                  "SG",
	adapt.TrdMarket_Futures_Simulate_HK: "FUTURES_SIMULATE_HK",
	adapt.TrdMarket_Futures_Simulate_US: "FUTURES_SIMULATE_US",
	adapt.TrdMarket_Futures_Simulate_SG: "FUTURES_SIMULATE_SG",
	adapt.TrdMarket_Futures_Simulate_JP: "FUTURES_SIMULATE_JP",
	adapt.TrdMarket_HK_Fund:             "HK_FUND",
	adapt.TrdMarket_US_Fund:             "US_FUND",
}

var trdMarketIDs = func() map[string]int32 {
	m := make(map[string]int32, len(trdMarketNames))
	for id, name := range trdMarketNames {
		m[name] = id
	}
	return m
}()

var trdAccTypeNames = map[int32]string{
	int32(trdcommon.TrdAccType_TrdAccType_Unknown):     "UNKNOWN",
	int32(trdcommon.TrdAccType_TrdAccType_Cash):        "CASH",
	int32(trdcommon.TrdAccType_TrdAccType_Margin):      "MARGIN",
	int32(trdcommon.TrdAccType_TrdAccType_TFSA):        "TFSA",
	int32(trdcommon.TrdAccType_TrdAccType_RRSP):        "RRSP",
	int32(trdcommon.TrdAccType_TrdAccType_SRRSP):       "SRRSP",
	int32(trdcommon.TrdAccType_TrdAccType_Derivatives): "DERIVATIVES",
}

var securityFirmNames = map[int32]string{
	adapt.SecurityFirm_Unknown:        "UNKNOWN",
	adapt.SecurityFirm_FutuSecurities: "FUTU_SECURITIES",
	adapt.SecurityFirm_FutuInc:        "FUTU_INC",
	adapt.SecurityFirm_FutuSG:         "FUTU_SG",
	adapt.SecurityFirm_FutuAU:         "FUTU_AU",
}

var orderTypeIDs = map[string]int32{
	"NORMAL":              adapt.OrderType_Normal,
	"MARKET":              adapt.OrderType_Market,
	"ABSOLUTE_LIMIT":      adapt.OrderType_AbsoluteLimit,
	"AUCTION":             adapt.OrderType_Auction,
	"AUCTION_LIMIT":       adapt.OrderType_AuctionLimit,
	"SPECIAL_LIMIT":       adapt.OrderType_SpecialLimit,
	"SPECIAL_LIMIT_ALL":   adapt.OrderType_SpecialLimit_All,
	"STOP":                adapt.OrderType_Stop,
	"STOP_LIMIT":          adapt.OrderType_StopLimit,
	"MARKET_IF_TOUCHED":   adapt.OrderType_MarketifTouched,
	"LIMIT_IF_TOUCHED":    adapt.OrderType_LimitifTouched,
	"TRAILING_STOP":       adapt.OrderType_TrailingStop,
	"TRAILING_STOP_LIMIT": adapt.OrderType_TrailingStopLimit,
}

var cashFlowDirectionNames = map[int32]string{
	int32(trdflowsummary.TrdCashFlowDirection_TrdCashFlowDirection_Unknown): "UNKNOWN",
	int32(trdflowsummary.TrdCashFlowDirection_TrdCashFlowDirection_In):      "IN",
	int32(trdflowsummary.TrdCashFlowDirection_TrdCashFlowDirection_Out):     "OUT",
}

func trdEnvName(env int32) string {
	if env == int32(trdcommon.TrdEnv_TrdEnv_Real) {
		return "REAL"
	}
	return "SIMULATE"
}

func positionSideName(side int32) string {
	switch trdcommon.PositionSide(side) {
	case trdcommon.PositionSide_PositionSide_Long:
		return "LONG"
	case trdcommon.PositionSide_PositionSide_Short:
		return "SHORT"
	default:
		return "UNKNOWN"
	}
}

// tradeHeader builds a TrdHeader for the given account, environment and
// market. It refuses to build a REAL header when the client is
// simulate-only, so account tools can never reach a real trading account
// without an explicit trade password configured.
func (c *Client) tradeHeader(accountID uint64, trdEnv, trdMarket string) (*trdcommon.TrdHeader, error) {
	marketID, ok := trdMarketIDs[strings.ToUpper(trdMarket)]
	if !ok {
		return nil, fmt.Errorf("unknown trd_market %q", trdMarket)
	}

	switch strings.ToUpper(trdEnv) {
	case "", "SIMULATE":
		return adapt.NewSimulationTradeHeader(accountID, marketID), nil
	case "REAL":
		if c.simulateOnly {
			return nil, fmt.Errorf("real account access is disabled: no trade password configured")
		}
		return adapt.NewTradeHeader(accountID, marketID), nil
	default:
		return nil, fmt.Errorf("unknown trd_env %q; must be REAL or SIMULATE", trdEnv)
	}
}

// GetAccounts returns the trading accounts visible to this OpenD session.
func (c *Client) GetAccounts(ctx context.Context) ([]Account, error) {
	accs, err := c.sdk.GetAccListWithContext(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Account, 0, len(accs))
	for _, a := range accs {
		markets := make([]string, 0, len(a.GetTrdMarketAuthList()))
		for _, m := range a.GetTrdMarketAuthList() {
			if name, ok := trdMarketNames[m]; ok {
				markets = append(markets, name)
			}
		}
		out = append(out, Account{
			AccountID:    strconv.FormatUint(a.GetAccID(), 10),
			TrdEnv:       trdEnvName(a.GetTrdEnv()),
			TrdMarkets:   markets,
			AccType:      trdAccTypeNames[a.GetAccType()],
			SecurityFirm: securityFirmNames[a.GetSecurityFirm()],
		})
	}
	return out, nil
}

// GetAssets returns funds/asset information for one account.
func (c *Client) GetAssets(ctx context.Context, accountID uint64, trdEnv, trdMarket string) (*Assets, error) {
	header, err := c.tradeHeader(accountID, trdEnv, trdMarket)
	if err != nil {
		return nil, err
	}
	f, err := c.sdk.GetFundsWithContext(ctx, header)
	if err != nil {
		return nil, err
	}
	return &Assets{
		Power:                 f.GetPower(),
		TotalAssets:           f.GetTotalAssets(),
		Cash:                  f.GetCash(),
		MarketValue:           f.GetMarketVal(),
		FrozenCash:            f.GetFrozenCash(),
		DebtCash:              f.GetDebtCash(),
		AvailableWithdrawCash: f.GetAvlWithdrawalCash(),
		RiskLevel:             f.GetRiskLevel(),
	}, nil
}

// GetPositions returns open positions for one account.
func (c *Client) GetPositions(ctx context.Context, accountID uint64, trdEnv, trdMarket string) ([]Position, error) {
	header, err := c.tradeHeader(accountID, trdEnv, trdMarket)
	if err != nil {
		return nil, err
	}
	positions, err := c.sdk.GetPositionListWithContext(ctx, header)
	if err != nil {
		return nil, err
	}
	out := make([]Position, 0, len(positions))
	for _, p := range positions {
		out = append(out, Position{
			PositionID:   strconv.FormatUint(p.GetPositionID(), 10),
			PositionSide: positionSideName(p.GetPositionSide()),
			Code:         p.GetCode(),
			Name:         p.GetName(),
			Qty:          p.GetQty(),
			CanSellQty:   p.GetCanSellQty(),
			Price:        p.GetPrice(),
			CostPrice:    p.GetCostPrice(),
			Value:        p.GetVal(),
			PlValue:      p.GetPlVal(),
			PlRatio:      p.GetPlRatio(),
		})
	}
	return out, nil
}

// GetMaxTradable returns the maximum tradable quantities for a security
// under one account.
func (c *Client) GetMaxTradable(ctx context.Context, accountID uint64, trdEnv, trdMarket, orderType, code string, price float64) (*MaxTradable, error) {
	header, err := c.tradeHeader(accountID, trdEnv, trdMarket)
	if err != nil {
		return nil, err
	}
	ot, ok := orderTypeIDs[strings.ToUpper(orderType)]
	if !ok {
		return nil, fmt.Errorf("unknown order_type %q", orderType)
	}
	m, err := c.sdk.GetMaxTrdQtysWithContext(ctx, header, ot, code, price)
	if err != nil {
		return nil, err
	}
	return &MaxTradable{
		MaxCashBuy:          m.GetMaxCashBuy(),
		MaxCashAndMarginBuy: m.GetMaxCashAndMarginBuy(),
		MaxPositionSell:     m.GetMaxPositionSell(),
		MaxSellShort:        m.GetMaxSellShort(),
		MaxBuyBack:          m.GetMaxBuyBack(),
	}, nil
}

// GetMarginRatio returns margin ratio info for the given securities under
// one account.
func (c *Client) GetMarginRatio(ctx context.Context, accountID uint64, trdEnv, trdMarket string, codes []string) ([]MarginRatio, error) {
	header, err := c.tradeHeader(accountID, trdEnv, trdMarket)
	if err != nil {
		return nil, err
	}
	infos, err := c.sdk.GetMarginRatioWithContext(ctx, header, codes)
	if err != nil {
		return nil, err
	}
	out := make([]MarginRatio, 0, len(infos))
	for _, info := range infos {
		out = append(out, MarginRatio{
			Code:            adapt.SecurityToCode(info.GetSecurity()),
			IsLongPermit:    info.GetIsLongPermit(),
			IsShortPermit:   info.GetIsShortPermit(),
			ShortPoolRemain: info.GetShortPoolRemain(),
			ShortFeeRate:    info.GetShortFeeRate(),
		})
	}
	return out, nil
}

// GetCashFlow returns the trading cash flow summary for one account on the
// given clearing date.
func (c *Client) GetCashFlow(ctx context.Context, accountID uint64, trdEnv, trdMarket, clearingDate string) ([]CashFlow, error) {
	header, err := c.tradeHeader(accountID, trdEnv, trdMarket)
	if err != nil {
		return nil, err
	}
	infos, err := c.sdk.TrdFlowSummaryWithContext(ctx, header, clearingDate)
	if err != nil {
		return nil, err
	}
	out := make([]CashFlow, 0, len(infos))
	for _, info := range infos {
		out = append(out, CashFlow{
			ClearingDate: info.GetClearingDate(),
			CashFlowType: info.GetCashFlowType(),
			Direction:    cashFlowDirectionNames[info.GetCashFlowDirection()],
			Amount:       info.GetCashFlowAmount(),
			Remark:       info.GetCashFlowRemark(),
		})
	}
	return out, nil
}
