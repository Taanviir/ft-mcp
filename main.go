package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/mcp42/intra"
	"github.com/tanas/mcp42/server"
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
		mcpSecret := os.Getenv("MCP_SECRET")
		if mcpSecret == "" {
			log.Println("warning: MCP_SECRET not set, server is unauthenticated")
		}

		mcpHandler := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return s },
			&mcp.StreamableHTTPOptions{Stateless: true},
		)

		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/oauth-authorization-server", oauthMetadata)
		mux.HandleFunc("/token", tokenHandler(mcpSecret))
		mux.Handle("/", requireSecret(mcpSecret, http.StripPrefix("/mcp", mcpHandler)))

		addr := ":" + *port
		log.Printf("42 MCP server listening on %s/mcp", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal(err)
		}
	default:
		if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}

func oauthMetadata(w http.ResponseWriter, r *http.Request) {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	base := scheme + "://" + r.Host
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"issuer":                                base,
		"token_endpoint":                        base + "/token",
		"grant_types_supported":                 []string{"client_credentials"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
	})
}

func tokenHandler(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()

		clientSecret := r.FormValue("client_secret")

		// also support Basic auth
		if clientSecret == "" {
			_, cs, ok := r.BasicAuth()
			if ok {
				clientSecret = cs
			}
		}

		if r.FormValue("grant_type") != "client_credentials" || clientSecret != secret {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"error": "invalid_client"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": secret,
			"token_type":   "Bearer",
			"expires_in":   86400,
		})
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
