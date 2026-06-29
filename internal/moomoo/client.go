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
}

// Client wraps the hyperjiang/futu SDK.
type Client struct {
	sdk *futu.SDK
}

// New connects to OpenD and returns a Client ready for use.
func New(host string, port int) (*Client, error) {
	sdk, err := futu.NewSDK(
		client.WithAddr(fmt.Sprintf("%s:%d", host, port)),
	)
	if err != nil {
		return nil, err
	}
	return &Client{sdk: sdk}, nil
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
