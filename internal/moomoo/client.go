package moomoo

import (
	"context"
	"fmt"

	"github.com/hyperjiang/futu"
	"github.com/hyperjiang/futu/client"
)

// MoomooClient is the interface for all moomoo operations. Defining it here
// lets tools in internal/tools use mock implementations in unit tests.
type MoomooClient interface {
	Health(ctx context.Context) error
	Close() error
	GetSnapshot(ctx context.Context, codes []string) ([]MarketSnapshot, error)
	GetQuote(ctx context.Context, codes []string) ([]Quote, error)
	GetKlines(ctx context.Context, code string, klType int32, beginTime, endTime string) ([]Kline, error)
	GetOrderBook(ctx context.Context, code string) (*OrderBook, error)
	GetAccounts(ctx context.Context) ([]Account, error)
	GetAssets(ctx context.Context, accountID uint64, trdEnv, trdMarket string) (*Assets, error)
	GetPositions(ctx context.Context, accountID uint64, trdEnv, trdMarket string) ([]Position, error)
	GetMaxTradable(ctx context.Context, accountID uint64, trdEnv, trdMarket, orderType, code string, price float64) (*MaxTradable, error)
	GetMarginRatio(ctx context.Context, accountID uint64, trdEnv, trdMarket string, codes []string) ([]MarginRatio, error)
	GetCashFlow(ctx context.Context, accountID uint64, trdEnv, trdMarket, clearingDate string) ([]CashFlow, error)
	GetUserSecurityGroup(ctx context.Context, groupType string) ([]SecurityGroup, error)
	GetUserSecurity(ctx context.Context, groupName string) ([]WatchlistSecurity, error)
}

// Client wraps the hyperjiang/futu SDK.
type Client struct {
	sdk *futu.SDK
	// simulateOnly blocks building a REAL trade header when no trade
	// password was configured, so read-only account tools can't be pointed
	// at a real account by mistake.
	simulateOnly bool
}

// New connects to OpenD and returns a Client ready for use. simulateOnly
// should be true when no trade password is configured; it prevents any
// account tool from querying a REAL trading account.
func New(host string, port int, simulateOnly bool) (*Client, error) {
	sdk, err := futu.NewSDK(
		client.WithAddr(fmt.Sprintf("%s:%d", host, port)),
	)
	if err != nil {
		return nil, err
	}
	return &Client{sdk: sdk, simulateOnly: simulateOnly}, nil
}

// Close shuts down the OpenD connection.
func (c *Client) Close() error {
	return c.sdk.Close()
}

// Health pings OpenD via GetGlobalState. Returns nil when the connection is live.
func (c *Client) Health(ctx context.Context) error {
	_, err := c.sdk.GetGlobalStateWithContext(ctx)
	return err
}
