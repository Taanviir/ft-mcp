package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redis/go-redis/v9"
	"github.com/tanas/ft-mcp/intra"
	tools "github.com/tanas/ft-mcp/server"
	tokenui "github.com/tanas/ft-mcp/server/token"
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

	log.SetOutput(os.Stdout)
	godotenv.Load()

	s := mcp.NewServer(&mcp.Implementation{Name: "42-api", Version: "1.0.0"}, nil)

	switch *transport {
	case "http":
		tools.RegisterAll(s, nil)

		var sessions sessionManager
		if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
			rs, err := newRedisSessionStore(redisURL)
			if err != nil {
				log.Printf("sessions: redis unavailable (%v), falling back to in-memory", err)
				sessions = newMemSessionStore()
			} else {
				log.Printf("sessions: using redis")
				sessions = rs
			}
		} else {
			sessions = newMemSessionStore()
		}
		go func() {
			ticker := time.NewTicker(time.Hour)
			for range ticker.C {
				sessions.cleanup()
				cleanupCodes()
			}
		}()

		mcpHandler := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return s },
			&mcp.StreamableHTTPOptions{Stateless: true},
		)

		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		mux.HandleFunc("/.well-known/oauth-authorization-server", oauthMetadata)
		mux.HandleFunc("POST /register", registrationHandler)
		mux.HandleFunc("/authorize", authorizeHandler)
		mux.HandleFunc("GET /token", func(w http.ResponseWriter, r *http.Request) { tokenui.Serve(w, "", "") })
		mux.HandleFunc("POST /token", tokenHandler(sessions))
		mux.Handle("/", requireAuth(sessions, http.StripPrefix("/mcp", mcpHandler)))

		addr := ":" + *port
		log.Printf("42 MCP server listening on %s/mcp", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal(err)
		}

	default:
		clientID := os.Getenv("FT_CLIENT_ID")
		clientSecret := os.Getenv("FT_CLIENT_SECRET")
		if clientID == "" || clientSecret == "" {
			log.Fatal("FT_CLIENT_ID and FT_CLIENT_SECRET must be set for stdio transport")
		}
		tools.RegisterAll(s, intra.New(clientID, clientSecret))
		log.Printf("42 MCP server running on stdio")
		if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}

// --- Session store ---

const sessionTTL = 24 * time.Hour

type sessionManager interface {
	create(client *intra.Client) string
	get(token string) (*intra.Client, bool)
	cleanup()
}

func newSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// in-memory implementation

type memSession struct {
	client    *intra.Client
	expiresAt time.Time
}

type memSessionStore struct {
	mu   sync.Mutex
	data map[string]memSession
}

func newMemSessionStore() *memSessionStore {
	return &memSessionStore{data: make(map[string]memSession)}
}

func (s *memSessionStore) create(client *intra.Client) string {
	token := newSessionToken()
	s.mu.Lock()
	s.data[token] = memSession{client: client, expiresAt: time.Now().Add(sessionTTL)}
	n := len(s.data)
	s.mu.Unlock()
	log.Printf("sessions: created (total active: %d)", n)
	return token
}

func (s *memSessionStore) get(token string) (*intra.Client, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.data[token]
	if !ok || time.Now().After(sess.expiresAt) {
		delete(s.data, token)
		return nil, false
	}
	return sess.client, true
}

func (s *memSessionStore) cleanup() {
	s.mu.Lock()
	now := time.Now()
	n := 0
	for token, sess := range s.data {
		if now.After(sess.expiresAt) {
			delete(s.data, token)
			n++
		}
	}
	s.mu.Unlock()
	if n > 0 {
		log.Printf("sessions: removed %d expired", n)
	}
}

// Redis implementation

type redisSessionStore struct {
	rdb *redis.Client
}

type redisSessionData struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

const redisKeyPrefix = "ft-mcp:session:"

func newRedisSessionStore(redisURL string) (*redisSessionStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &redisSessionStore{rdb: rdb}, nil
}

func (s *redisSessionStore) create(client *intra.Client) string {
	token := newSessionToken()
	id, secret := client.Credentials()
	data, _ := json.Marshal(redisSessionData{ClientID: id, ClientSecret: secret})
	s.rdb.Set(context.Background(), redisKeyPrefix+token, data, sessionTTL)
	log.Printf("sessions: created (redis)")
	return token
}

func (s *redisSessionStore) get(token string) (*intra.Client, bool) {
	data, err := s.rdb.Get(context.Background(), redisKeyPrefix+token).Bytes()
	if err != nil {
		return nil, false
	}
	var sd redisSessionData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, false
	}
	return intra.New(sd.ClientID, sd.ClientSecret), true
}

func (s *redisSessionStore) cleanup() {
	// Redis expires keys automatically via TTL
}

// --- OAuth metadata ---

func oauthMetadata(w http.ResponseWriter, r *http.Request) {
	base := serverBase(r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"issuer":                                base,
		"authorization_endpoint":                base + "/authorize",
		"token_endpoint":                        base + "/token",
		"registration_endpoint":                 base + "/register",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "client_credentials"},
		"code_challenge_methods_supported":       []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
	})
}

// --- Dynamic client registration (RFC 7591) ---

type clientReg struct {
	secret       string
	redirectURIs []string
}

var (
	clientsMu sync.Mutex
	clients   = map[string]clientReg{}
)

func registrationHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RedirectURIs []string `json:"redirect_uris"`
		ClientName   string   `json:"client_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.RedirectURIs) == 0 {
		http.Error(w, "redirect_uris required", http.StatusBadRequest)
		return
	}

	idBytes := make([]byte, 16)
	rand.Read(idBytes)
	secretBytes := make([]byte, 32)
	rand.Read(secretBytes)
	clientID := base64.RawURLEncoding.EncodeToString(idBytes)
	clientSecret := base64.RawURLEncoding.EncodeToString(secretBytes)

	clientsMu.Lock()
	clients[clientID] = clientReg{secret: clientSecret, redirectURIs: req.RedirectURIs}
	clientsMu.Unlock()

	log.Printf("oauth: registered client %q", req.ClientName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"client_id":                  clientID,
		"client_secret":              clientSecret,
		"redirect_uris":              req.RedirectURIs,
		"grant_types":                []string{"authorization_code"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "client_secret_post",
	})
}

// --- Authorization Code + PKCE ---

type codeEntry struct {
	challenge      string
	ftClientID     string // set by the authorize form; empty for legacy clients
	ftClientSecret string
	expiresAt      time.Time
}

var (
	codeMu sync.Mutex
	codes  = map[string]codeEntry{}
)

func cleanupCodes() {
	codeMu.Lock()
	now := time.Now()
	n := 0
	for code, entry := range codes {
		if now.After(entry.expiresAt) {
			delete(codes, code)
			n++
		}
	}
	codeMu.Unlock()
	if n > 0 {
		log.Printf("oauth: removed %d expired codes", n)
	}
}

func validRedirectURI(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "localhost" || host == "127.0.0.1" || u.Scheme == "https"
}

const authorizeFormHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Connect ft-mcp</title>
  <style>
    *{box-sizing:border-box}
    body{font-family:system-ui,sans-serif;max-width:380px;margin:80px auto;padding:0 20px;color:#111}
    h2{margin:0 0 6px}
    p{margin:0 0 24px;color:#555;font-size:14px}
    a{color:#0066ff}
    label{display:block;margin-bottom:4px;font-weight:500;font-size:14px}
    input[type=text],input[type=password]{display:block;width:100%;padding:8px 10px;border:1px solid #ccc;border-radius:6px;margin-bottom:14px;font-size:14px}
    button{width:100%;padding:10px;background:#0066ff;color:#fff;border:none;border-radius:6px;font-size:15px;cursor:pointer}
    button:hover{background:#0052cc}
    .err{color:#c00;font-size:14px;margin-bottom:14px}
  </style>
</head>
<body>
  <h2>Connect ft-mcp</h2>
  <p>Enter your <a href="https://profile.intra.42.fr/oauth/applications" target="_blank">42 API credentials</a> to grant access.</p>
  <form method="POST">
    {{ERROR}}
    <input type="hidden" name="redirect_uri" value="{{REDIRECT_URI}}">
    <input type="hidden" name="state" value="{{STATE}}">
    <input type="hidden" name="code_challenge" value="{{CODE_CHALLENGE}}">
    <input type="hidden" name="client_id" value="{{CLIENT_ID}}">
    <label>Client UID</label>
    <input type="text" name="ft_client_id" placeholder="u-s4t2ud-..." autocomplete="username" required>
    <label>Client Secret</label>
    <input type="password" name="ft_client_secret" autocomplete="current-password" required>
    <button type="submit">Connect</button>
  </form>
</body>
</html>`

func serveAuthorizeForm(w http.ResponseWriter, redirectURI, state, challenge, clientID, errMsg string) {
	errHTML := ""
	if errMsg != "" {
		errHTML = `<div class="err">` + html.EscapeString(errMsg) + `</div>`
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	strings.NewReplacer(
		"{{ERROR}}", errHTML,
		"{{REDIRECT_URI}}", html.EscapeString(redirectURI),
		"{{STATE}}", html.EscapeString(state),
		"{{CODE_CHALLENGE}}", html.EscapeString(challenge),
		"{{CLIENT_ID}}", html.EscapeString(clientID),
	).WriteString(w, authorizeFormHTML)
}

func issueCodeAndRedirect(w http.ResponseWriter, r *http.Request, redirectURI, state, challenge, ftClientID, ftClientSecret string) {
	b := make([]byte, 32)
	rand.Read(b)
	code := base64.RawURLEncoding.EncodeToString(b)

	codeMu.Lock()
	codes[code] = codeEntry{
		challenge:      challenge,
		ftClientID:     ftClientID,
		ftClientSecret: ftClientSecret,
		expiresAt:      time.Now().Add(5 * time.Minute),
	}
	codeMu.Unlock()

	log.Printf("oauth: code issued for redirect %s", redirectURI)
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

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		redirectURI := r.FormValue("redirect_uri")
		state := r.FormValue("state")
		challenge := r.FormValue("code_challenge")
		clientID := r.FormValue("client_id")
		ftClientID := r.FormValue("ft_client_id")
		ftClientSecret := r.FormValue("ft_client_secret")

		if !validRedirectURI(redirectURI) {
			http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
			return
		}
		if ftClientID == "" || ftClientSecret == "" {
			serveAuthorizeForm(w, redirectURI, state, challenge, clientID, "Client UID and Secret are required.")
			return
		}
		c := intra.New(ftClientID, ftClientSecret)
		if err := c.Validate(); err != nil {
			log.Printf("oauth: invalid 42 credentials for %s", ftClientID)
			serveAuthorizeForm(w, redirectURI, state, challenge, clientID, "Invalid credentials — check your Client UID and Secret.")
			return
		}
		issueCodeAndRedirect(w, r, redirectURI, state, challenge, ftClientID, ftClientSecret)
		return
	}

	q := r.URL.Query()
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	challenge := q.Get("code_challenge")
	clientID := q.Get("client_id")

	if redirectURI == "" || challenge == "" {
		http.Error(w, "missing redirect_uri or code_challenge", http.StatusBadRequest)
		return
	}
	if !validRedirectURI(redirectURI) {
		http.Error(w, "redirect_uri must be localhost or https", http.StatusBadRequest)
		return
	}

	serveAuthorizeForm(w, redirectURI, state, challenge, clientID, "")
}

// --- Token endpoint ---

func tokenHandler(sessions sessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		// Browser form submission from the token UI page
		if r.FormValue("_source") == "ui" {
			clientID := r.FormValue("client_id")
			clientSecret := r.FormValue("client_secret")
			if clientID == "" || clientSecret == "" {
				tokenui.Serve(w, "Client UID and Secret are required.", "")
				return
			}
			c := intra.New(clientID, clientSecret)
			if err := c.Validate(); err != nil {
				log.Printf("token ui: invalid credentials for %s: %v", clientID, err)
				tokenui.Serve(w, "Invalid credentials — check your Client UID and Secret.", "")
				return
			}
			log.Printf("token ui: session created for %s", clientID)
			tokenui.Serve(w, "", sessions.create(c))
			return
		}

		clientID, clientSecret := extractCreds(r)

		switch r.FormValue("grant_type") {
		case "authorization_code":
			code := r.FormValue("code")
			verifier := r.FormValue("code_verifier")

			codeMu.Lock()
			entry, ok := codes[code]
			if ok {
				delete(codes, code)
			}
			codeMu.Unlock()

			if !ok || time.Now().After(entry.expiresAt) || !verifyPKCE(verifier, entry.challenge) {
				log.Printf("token: invalid grant — bad code or PKCE mismatch")
				oauthError(w, "invalid_grant", http.StatusBadRequest)
				return
			}

			if entry.ftClientID == "" {
				oauthError(w, "invalid_grant", http.StatusBadRequest)
				return
			}
			c := intra.New(entry.ftClientID, entry.ftClientSecret)
			log.Printf("token: session created for %s (authorization_code)", entry.ftClientID)
			writeToken(w, sessions.create(c))

		case "client_credentials":
			if clientID == "" || clientSecret == "" {
				oauthError(w, "invalid_client", http.StatusUnauthorized)
				return
			}
			c := intra.New(clientID, clientSecret)
			if err := c.Validate(); err != nil {
				log.Printf("credential validation failed for client %q: %v", clientID, err)
				oauthError(w, "invalid_client", http.StatusUnauthorized)
				return
			}
			log.Printf("token: session created for %s (client_credentials)", clientID)
			writeToken(w, sessions.create(c))

		default:
			oauthError(w, "unsupported_grant_type", http.StatusBadRequest)
		}
	}
}

// --- MCP auth middleware ---

func requireAuth(sessions sessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		client, ok := sessions.get(token)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(tools.WithClient(r.Context(), client)))
	})
}

// --- Helpers ---

func extractCreds(r *http.Request) (clientID, clientSecret string) {
	clientID = r.FormValue("client_id")
	clientSecret = r.FormValue("client_secret")
	if clientID == "" {
		clientID, clientSecret, _ = r.BasicAuth()
	}
	return
}

func verifyPKCE(verifier, challenge string) bool {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:]) == challenge
}

func writeToken(w http.ResponseWriter, token string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": token,
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
