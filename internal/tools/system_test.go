package tools

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockClient struct {
	healthErr error
}

func (m *mockClient) Health(_ context.Context) error { return m.healthErr }
func (m *mockClient) Close() error                   { return nil }

// connect creates an in-memory MCP client session connected to the given server.
func connectTest(ctx context.Context, s *mcp.Server) (*mcp.ClientSession, func()) {
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := s.Connect(ctx, t1, nil); err != nil {
		log.Fatalf("server connect: %v", err)
	}
	c := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	cs, err := c.Connect(ctx, t2, nil)
	if err != nil {
		log.Fatalf("client connect: %v", err)
	}
	return cs, func() { cs.Close() }
}

func TestCheckHealth_connected(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterSystem(s, &mockClient{})

	ctx := context.Background()
	cs, cleanup := connectTest(ctx, s)
	defer cleanup()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "check_health"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}

	text := res.Content[0].(*mcp.TextContent).Text
	var h healthResult
	if err := json.Unmarshal([]byte(text), &h); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if !h.Connected {
		t.Errorf("want connected=true, got %+v", h)
	}
	if h.Error != "" {
		t.Errorf("want empty error, got %q", h.Error)
	}
}

func TestCheckHealth_disconnected(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterSystem(s, &mockClient{healthErr: errors.New("dial failed")})

	ctx := context.Background()
	cs, cleanup := connectTest(ctx, s)
	defer cleanup()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "check_health"})
	if err != nil {
		t.Fatal(err)
	}

	text := res.Content[0].(*mcp.TextContent).Text
	var h healthResult
	if err := json.Unmarshal([]byte(text), &h); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, text)
	}
	if h.Connected {
		t.Error("want connected=false")
	}
	if h.Error == "" {
		t.Error("want non-empty error message")
	}
}
