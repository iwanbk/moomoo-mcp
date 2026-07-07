package moomoo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"syscall"
	"time"

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
	GetOrders(ctx context.Context, accountID uint64, trdEnv, trdMarket string) ([]Order, error)
	GetDeals(ctx context.Context, accountID uint64, trdEnv, trdMarket string) ([]Deal, error)
	GetHistoryOrders(ctx context.Context, accountID uint64, trdEnv, trdMarket, beginTime, endTime string, codes []string) ([]Order, error)
	GetHistoryDeals(ctx context.Context, accountID uint64, trdEnv, trdMarket, beginTime, endTime string, codes []string) ([]Deal, error)
}

// callTimeout bounds every OpenD request. Without it, a half-dead connection
// (peer gone but our socket not yet notified) can block a tool call forever,
// which makes the MCP server look hung to Claude Code and forces a manual
// `/mcp reconnect`.
const callTimeout = 30 * time.Second

// Client wraps the hyperjiang/futu SDK with lazy connect + auto-reconnect to
// OpenD. The underlying SDK does not reconnect on its own, so the Client owns
// the connection lifecycle: it dials on demand and re-dials after a drop. This
// keeps the MCP server alive and usable across OpenD restarts.
type Client struct {
	addr string
	// simulateOnly blocks building a REAL trade header when no trade
	// password was configured, so read-only account tools can't be pointed
	// at a real account by mistake.
	simulateOnly bool
	// disableRounding turns off number rounding on outgoing float fields
	// (prices, rates, turnover), returning raw SDK values instead.
	disableRounding bool

	// mu guards sdk. sdk is nil whenever there is no live connection.
	mu  sync.Mutex
	sdk *futu.SDK
}

// New returns a Client for the given OpenD address. simulateOnly should be
// true when no trade password is configured; it prevents any account tool from
// querying a REAL trading account. disableRounding turns off the number
// rounding normally applied to outgoing float fields.
//
// New never fails on connectivity: OpenD may not be up yet when the MCP server
// starts. The connection is established lazily on first use and re-established
// automatically after a drop, so the MCP server stays alive regardless of
// OpenD's state. check_health reports the live status.
func New(host string, port int, simulateOnly, disableRounding bool) (*Client, error) {
	c := &Client{
		addr:            fmt.Sprintf("%s:%d", host, port),
		simulateOnly:    simulateOnly,
		disableRounding: disableRounding,
	}
	// Try to connect eagerly so a healthy OpenD is ready immediately, but do
	// not treat failure as fatal.
	if sdk, err := c.dial(); err != nil {
		log.Printf("moomoo: initial connect to OpenD at %s failed, will retry on demand: %v", c.addr, err)
	} else {
		c.sdk = sdk
	}
	return c, nil
}

// dial opens a fresh SDK connection to OpenD.
func (c *Client) dial() (*futu.SDK, error) {
	return futu.NewSDK(client.WithAddr(c.addr))
}

// getSDK returns a live SDK, dialing lazily if there is no current connection.
func (c *Client) getSDK() (*futu.SDK, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sdk != nil {
		return c.sdk, nil
	}
	sdk, err := c.dial()
	if err != nil {
		return nil, fmt.Errorf("connect to OpenD at %s: %w", c.addr, err)
	}
	c.sdk = sdk
	return sdk, nil
}

// dropSDK invalidates the given connection (closing it) so the next call
// re-dials. It is a no-op if the current connection has already been replaced.
func (c *Client) dropSDK(bad *futu.SDK) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sdk == bad && c.sdk != nil {
		_ = c.sdk.Close()
		c.sdk = nil
	}
}

// Close shuts down the OpenD connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sdk == nil {
		return nil
	}
	err := c.sdk.Close()
	c.sdk = nil
	return err
}

// Health pings OpenD via GetGlobalState. Returns nil when the connection is live.
func (c *Client) Health(ctx context.Context) error {
	_, err := call(c, ctx, func(ctx context.Context, sdk *futu.SDK) (struct{}, error) {
		_, err := sdk.GetGlobalStateWithContext(ctx)
		return struct{}{}, err
	})
	return err
}

// call runs fn against a live OpenD connection, dialing lazily on first use.
// Every call is bounded by callTimeout. If the call fails because the
// connection is dead, call drops it and retries once on a fresh dial, so a
// transient OpenD restart doesn't require restarting the MCP server.
func call[T any](c *Client, ctx context.Context, fn func(ctx context.Context, sdk *futu.SDK) (T, error)) (T, error) {
	var zero T

	sdk, err := c.getSDK()
	if err != nil {
		return zero, err
	}

	cctx, cancel := context.WithTimeout(ctx, callTimeout)
	res, err := fn(cctx, sdk)
	cancel()
	if err == nil || !isConnErr(err) {
		return res, err
	}

	// Connection looks dead: drop it and retry once on a fresh dial.
	c.dropSDK(sdk)
	sdk, err = c.getSDK()
	if err != nil {
		return zero, err
	}
	cctx, cancel = context.WithTimeout(ctx, callTimeout)
	defer cancel()
	return fn(cctx, sdk)
}

// isConnErr reports whether err indicates a dead OpenD connection that a
// reconnect could recover, as opposed to a request-level or validation error.
func isConnErr(err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, io.EOF),
		errors.Is(err, net.ErrClosed),
		errors.Is(err, syscall.EPIPE),
		errors.Is(err, syscall.ECONNRESET),
		errors.Is(err, syscall.ECONNREFUSED),
		errors.Is(err, client.ErrChannelClosed),
		errors.Is(err, client.ErrInterrupted):
		return true
	}
	var netErr *net.OpError
	return errors.As(err, &netErr)
}
