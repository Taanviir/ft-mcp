package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/mcp42/client"
)

func RegisterAll(s *mcp.Server, c *client.Client) {
	registerUsers(s, c)
	registerCampus(s, c)
	registerProjects(s, c)
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
