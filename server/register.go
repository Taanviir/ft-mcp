package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/Taanviir/ft-mcp/intra"
)

type tools struct {
	client *intra.Client
}

func (t *tools) getClient(_ context.Context) (*intra.Client, error) {
	if t.client != nil {
		return t.client, nil
	}
	return nil, fmt.Errorf("not authenticated — FT_CLIENT_ID and FT_CLIENT_SECRET must be set")
}

func RegisterAll(s *mcp.Server, c *intra.Client) {
	t := &tools{client: c}
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
