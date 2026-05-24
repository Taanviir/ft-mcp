package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/mcp42/intra"
	"github.com/tanas/mcp42/tools"
)

func main() {
	transport := flag.String("transport", "stdio", "Transport to use: stdio or http")
	port := flag.String("port", "", "Port to listen on when transport=http (defaults to PORT env var, then 8080)")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8080"
		}
	}

	godotenv.Load()

	clientID := os.Getenv("FT_CLIENT_ID")
	clientSecret := os.Getenv("FT_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("FT_CLIENT_ID and FT_CLIENT_SECRET must be set")
	}

	apiClient := intra.New(clientID, clientSecret)

	s := mcp.NewServer(&mcp.Implementation{Name: "42-api", Version: "1.0.0"}, nil)
	tools.RegisterAll(s, apiClient)

	switch *transport {
	case "http":
		secret := os.Getenv("MCP_SECRET")
		if secret == "" {
			log.Println("warning: MCP_SECRET not set, server is unauthenticated")
		}
		handler := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return s },
			&mcp.StreamableHTTPOptions{Stateless: true},
		)
		addr := ":" + *port
		log.Printf("42 MCP server listening on %s/mcp", addr)
		if err := http.ListenAndServe(addr, requireSecret(secret, http.StripPrefix("/mcp", handler))); err != nil {
			log.Fatal(err)
		}
	default:
		if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}

func requireSecret(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			next.ServeHTTP(w, r)
			return
		}
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token != secret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
