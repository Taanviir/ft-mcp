package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/ft-mcp/intra"
	"github.com/tanas/ft-mcp/server"
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
		mux.HandleFunc("/authorize", authorizeHandler)
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

// --- OAuth metadata ---

func oauthMetadata(w http.ResponseWriter, r *http.Request) {
	base := serverBase(r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"issuer":                                base,
		"authorization_endpoint":                base + "/authorize",
		"token_endpoint":                        base + "/token",
		"grant_types_supported":                 []string{"authorization_code", "client_credentials"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
	})
}

// --- Authorization Code + PKCE ---

type codeEntry struct {
	challenge string
	expiresAt time.Time
}

var (
	codeMu sync.Mutex
	codes  = map[string]codeEntry{}
)

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	challenge := q.Get("code_challenge")

	if redirectURI == "" || challenge == "" {
		http.Error(w, "missing redirect_uri or code_challenge", http.StatusBadRequest)
		return
	}

	// Generate a one-time auth code
	b := make([]byte, 32)
	rand.Read(b)
	code := base64.RawURLEncoding.EncodeToString(b)

	codeMu.Lock()
	codes[code] = codeEntry{challenge: challenge, expiresAt: time.Now().Add(5 * time.Minute)}
	codeMu.Unlock()

	// Redirect back to client with code
	callback, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}
	cq := callback.Query()
	cq.Set("code", code)
	if state != "" {
		cq.Set("state", state)
	}
	callback.RawQuery = cq.Encode()
	http.Redirect(w, r, callback.String(), http.StatusFound)
}

// --- Token endpoint ---

func tokenHandler(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()

		switch r.FormValue("grant_type") {
		case "authorization_code":
			code := r.FormValue("code")
			verifier := r.FormValue("code_verifier")
			clientSecret := formOrBasicSecret(r)

			if secret != "" && clientSecret != secret {
				oauthError(w, "invalid_client", http.StatusUnauthorized)
				return
			}

			codeMu.Lock()
			entry, ok := codes[code]
			if ok {
				delete(codes, code)
			}
			codeMu.Unlock()

			if !ok || time.Now().After(entry.expiresAt) || !verifyPKCE(verifier, entry.challenge) {
				oauthError(w, "invalid_grant", http.StatusBadRequest)
				return
			}
			writeToken(w, secret)

		case "client_credentials":
			clientSecret := formOrBasicSecret(r)
			if secret != "" && clientSecret != secret {
				oauthError(w, "invalid_client", http.StatusUnauthorized)
				return
			}
			writeToken(w, secret)

		default:
			oauthError(w, "unsupported_grant_type", http.StatusBadRequest)
		}
	}
}

// --- MCP auth middleware ---

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

// --- Helpers ---

func verifyPKCE(verifier, challenge string) bool {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:]) == challenge
}

func formOrBasicSecret(r *http.Request) string {
	if s := r.FormValue("client_secret"); s != "" {
		return s
	}
	_, s, ok := r.BasicAuth()
	if ok {
		return s
	}
	return ""
}

func writeToken(w http.ResponseWriter, secret string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": secret,
		"token_type":   "Bearer",
		"expires_in":   86400,
	})
}

func oauthError(w http.ResponseWriter, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": code})
}

func serverBase(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	return scheme + "://" + r.Host
}
