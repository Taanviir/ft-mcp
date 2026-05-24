package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/ft-mcp/intra"
)

var api *intra.Client

type clientKey struct{}

// WithClient attaches a 42 API client to the context (used by HTTP middleware).
func WithClient(ctx context.Context, c *intra.Client) context.Context {
	return context.WithValue(ctx, clientKey{}, c)
}

// getClient returns the per-request client from context, falling back to the
// global set by RegisterAll (stdio transport only).
func getClient(ctx context.Context) (*intra.Client, error) {
	if c, ok := ctx.Value(clientKey{}).(*intra.Client); ok && c != nil {
		return c, nil
	}
	if api != nil {
		return api, nil
	}
	return nil, fmt.Errorf("not authenticated — provide your 42 API credentials")
}

func RegisterAll(s *mcp.Server, c *intra.Client) {
	api = c
	registerUsers(s)
	registerCampus(s)
	registerProjects(s)
	registerResources(s)
	registerPrompts(s)
}

func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}
}

func textResult(data []byte) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}
