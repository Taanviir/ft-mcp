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
	"github.com/tanas/ft-mcp/intra"
	tools "github.com/tanas/ft-mcp/server"
)

func main() {
	transport := flag.String("transport", "stdio", "Transport: stdio or http")
	port := flag.String("port", "", "Port for HTTP mode (defaults to PORT env var, then 8080)")
	flag.Parse()

	log.SetOutput(os.Stdout)
	godotenv.Load()

	clientID := os.Getenv("FT_CLIENT_ID")
	clientSecret := os.Getenv("FT_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("FT_CLIENT_ID and FT_CLIENT_SECRET must be set")
	}

	s := mcp.NewServer(&mcp.Implementation{Name: "ft-mcp", Version: "1.0.0"}, nil)
	tools.RegisterAll(s, intra.New(clientID, clientSecret))

	switch *transport {
	case "http":
		if *port == "" {
			if p := os.Getenv("PORT"); p != "" {
				*port = p
			} else {
				*port = "8080"
			}
		}
		mcpHandler := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return s },
			&mcp.StreamableHTTPOptions{Stateless: true},
		)
		mux := http.NewServeMux()
		mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		mux.Handle("POST /mcp", withAuth(mcpHandler))
		mux.Handle("GET /mcp", withAuth(mcpHandler))

		addr := ":" + *port
		log.Printf("ft-mcp listening on %s/mcp", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal(err)
		}

	default:
		if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}

// withAuth wraps a handler with static bearer token auth when AUTH_TOKEN is set.
// If AUTH_TOKEN is not set the handler is returned unwrapped.
func withAuth(next http.Handler) http.Handler {
	token := os.Getenv("AUTH_TOKEN")
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || got != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
