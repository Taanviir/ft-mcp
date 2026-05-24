package intra

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	baseURL  = "https://api.intra.42.fr"
	tokenURL = baseURL + "/oauth/token"
	apiBase  = baseURL + "/v2"
)

type Client struct {
	clientID     string
	clientSecret string
	mu           sync.Mutex
	accessToken  string
	expiresAt    time.Time
	http         *http.Client
}

func New(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		http:         &http.Client{Timeout: 15 * time.Second},
	}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *Client) ensureToken() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return nil
	}

	log.Printf("intra: fetching 42 API token for %s", c.clientID)
	body := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
	}
	resp, err := c.http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request returned %d: %s", resp.StatusCode, raw)
	}

	var tr tokenResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return fmt.Errorf("parsing token response: %w", err)
	}

	c.accessToken = tr.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(tr.ExpiresIn-60) * time.Second)
	return nil
}

// Credentials returns the client ID and secret used to construct this client.
func (c *Client) Credentials() (string, string) {
	return c.clientID, c.clientSecret
}

// Validate verifies that the credentials can obtain a 42 API token.
func (c *Client) Validate() error {
	return c.ensureToken()
}

func (c *Client) doGet(path string, params url.Values) ([]byte, http.Header, error) {
	reqURL := apiBase + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, nil, err
	}
	c.mu.Lock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.Unlock()
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.Header, fmt.Errorf("GET %s returned %d: %s", path, resp.StatusCode, raw)
	}
	return raw, resp.Header, nil
}

// Get fetches path (e.g. "/users/login") with optional query params.
func (c *Client) Get(path string, params url.Values) ([]byte, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}
	body, _, err := c.doGet(path, params)
	return body, err
}

// GetWithTotal fetches path and also returns the X-Total count header (-1 if absent).
func (c *Client) GetWithTotal(path string, params url.Values) ([]byte, int, error) {
	if err := c.ensureToken(); err != nil {
		return nil, -1, err
	}
	body, headers, err := c.doGet(path, params)
	if err != nil {
		return nil, -1, err
	}
	total := -1
	if h := headers.Get("X-Total"); h != "" {
		if n, err := strconv.Atoi(h); err == nil {
			total = n
		}
	}
	return body, total, nil
}

// Count returns the total number of records for a query without fetching all data.
func (c *Client) Count(path string, params url.Values) (int, error) {
	p := make(url.Values)
	for k, v := range params {
		p[k] = v
	}
	p.Set("page[size]", "1")
	p.Set("page[number]", "1")
	_, total, err := c.GetWithTotal(path, p)
	if err != nil {
		return 0, err
	}
	if total < 0 {
		return 0, fmt.Errorf("API did not return a total count for this endpoint")
	}
	return total, nil
}
