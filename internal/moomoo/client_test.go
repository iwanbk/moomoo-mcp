package moomoo

import (
	"context"
	"errors"
	"io"
	"net"
	"syscall"
	"testing"

	"github.com/hyperjiang/futu/client"
)

// TestNew_doesNotFailWhenOpenDDown is the core resilience guarantee: the MCP
// server must come up even when OpenD is unreachable, so Claude Code keeps the
// stdio connection alive instead of prompting for `/mcp reconnect`.
func TestNew_doesNotFailWhenOpenDDown(t *testing.T) {
	// Port 1 is reserved/unusable, so the dial is guaranteed to fail fast.
	c, err := New("127.0.0.1", 1, true, false)
	if err != nil {
		t.Fatalf("New returned error for unreachable OpenD, want nil: %v", err)
	}
	if c == nil {
		t.Fatal("New returned nil client")
	}
	if c.sdk != nil {
		t.Error("expected no live SDK connection when OpenD is down")
	}
}

// TestHealth_reportsErrorWhenOpenDDown verifies a call surfaces a connection
// error (rather than hanging or crashing) when OpenD cannot be reached, and
// leaves no live connection behind.
func TestHealth_reportsErrorWhenOpenDDown(t *testing.T) {
	c, _ := New("127.0.0.1", 1, true, false)
	if err := c.Health(context.Background()); err == nil {
		t.Fatal("Health returned nil error with OpenD down, want a connection error")
	}
	if c.sdk != nil {
		t.Error("expected no live SDK connection after a failed dial")
	}
}

func TestIsConnErr(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"eof", io.EOF, true},
		{"net closed", net.ErrClosed, true},
		{"broken pipe", syscall.EPIPE, true},
		{"conn reset", syscall.ECONNRESET, true},
		{"conn refused", syscall.ECONNREFUSED, true},
		{"sdk channel closed", client.ErrChannelClosed, true},
		{"sdk interrupted", client.ErrInterrupted, true},
		{"wrapped net op error", &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}, true},
		{"wrapped eof", errors.New("read: " + io.EOF.Error()), false},
		{"plain request error", errors.New("invalid security code"), false},
		{"context canceled is not a conn error", context.Canceled, false},
	}
	for _, tc := range cases {
		if got := isConnErr(tc.err); got != tc.want {
			t.Errorf("%s: isConnErr(%v) = %v, want %v", tc.name, tc.err, got, tc.want)
		}
	}
}
