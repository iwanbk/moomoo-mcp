package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/config"
	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
	"github.com/iwanbk/moomoo-mcp/internal/tools"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	// New does not fail on connectivity: OpenD may not be up yet. The client
	// connects lazily and reconnects automatically after a drop, so the MCP
	// server stays alive across OpenD restarts. check_health reports status.
	client, err := moomoo.New(cfg.OpendHost, cfg.OpendPort, cfg.SimulateOnly, cfg.DisableRounding)
	if err != nil {
		log.Fatalf("init moomoo client: %v", err)
	}
	defer client.Close()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "moomoo-mcp",
		Version: "0.1.0",
	}, nil)

	tools.RegisterSystem(server, client)
	tools.RegisterMarketData(server, client)
	tools.RegisterAccount(server, client)
	tools.RegisterWatchlist(server, client)
	tools.RegisterOrders(server, client)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server: %v", err)
	}
}
