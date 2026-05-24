package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/ft-mcp/intra"
)

type clientKey struct{}

type tools struct {
	fallback *intra.Client // non-nil for stdio, nil for HTTP (client comes from context)
}

// WithClient attaches a 42 API client to the context (used by HTTP middleware).
func WithClient(ctx context.Context, c *intra.Client) context.Context {
	return context.WithValue(ctx, clientKey{}, c)
}

func (t *tools) getClient(ctx context.Context) (*intra.Client, error) {
	if c, ok := ctx.Value(clientKey{}).(*intra.Client); ok && c != nil {
		return c, nil
	}
	if t.fallback != nil {
		return t.fallback, nil
	}
	return nil, fmt.Errorf("not authenticated — provide your 42 API credentials")
}

func RegisterAll(s *mcp.Server, c *intra.Client) {
	t := &tools{fallback: c}
	t.registerUsers(s)
	t.registerCampus(s)
	t.registerProjects(s)
	t.registerResources(s)
	t.registerPrompts(s)
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
