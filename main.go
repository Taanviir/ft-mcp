package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/mcp42/client"
	"github.com/tanas/mcp42/tools"
)

func main() {
	transport := flag.String("transport", "stdio", "Transport to use: stdio or http")
	port := flag.String("port", "", "Port to listen on when transport=http (defaults to PORT env var, then 8080)")
	flag.Parse()

	// Hosting platforms inject PORT automatically.
	if *port == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8080"
		}
	}

	godotenv.Load() // load .env if present; real env vars take precedence

	clientID := os.Getenv("FT_CLIENT_ID")
	clientSecret := os.Getenv("FT_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("FT_CLIENT_ID and FT_CLIENT_SECRET must be set")
	}

	apiClient := client.New(clientID, clientSecret)

	s := mcp.NewServer(&mcp.Implementation{Name: "42-api", Version: "1.0.0"}, nil)
	tools.RegisterAll(s, apiClient)

	switch *transport {
	case "http":
		handler := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return s },
			&mcp.StreamableHTTPOptions{Stateless: true},
		)
		addr := ":" + *port
		log.Printf("42 MCP server listening on %s/mcp", addr)
		if err := http.ListenAndServe(addr, http.StripPrefix("/mcp", handler)); err != nil {
			log.Fatal(err)
		}
	default: // run with stdio transport by default
		if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}
