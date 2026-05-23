package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/mcp42/intra"
)

var api *intra.Client

func RegisterAll(s *mcp.Server, c *intra.Client) {
	api = c
	registerUsers(s)
	registerCampus(s)
	registerProjects(s)
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
