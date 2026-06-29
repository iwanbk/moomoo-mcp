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

	client, err := moomoo.New(cfg.OpendHost, cfg.OpendPort)
	if err != nil {
		log.Fatalf("connect to OpenD at %s:%d: %v", cfg.OpendHost, cfg.OpendPort, err)
	}
	defer client.Close()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "moomoo-mcp",
		Version: "0.1.0",
	}, nil)

	tools.RegisterSystem(server, client)
	tools.RegisterMarketData(server, client)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server: %v", err)
	}
}
