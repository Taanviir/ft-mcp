package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/mcp42/client"
)

type listCursusInput struct {
	Page    int `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage int `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

func handleListCursus(c *client.Client) func(context.Context, *mcp.CallToolRequest, listCursusInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input listCursusInput) (*mcp.CallToolResult, any, error) {
		data, err := c.Get("/cursus", paginationParams(input.Page, input.PerPage))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

type listProjectsInput struct {
	CursusID int `json:"cursus_id,omitempty" jsonschema:"filter projects by cursus ID (optional)"`
	Page     int `json:"page,omitempty"      jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty"  jsonschema:"results per page, max 100"`
}

func handleListProjects(c *client.Client) func(context.Context, *mcp.CallToolRequest, listProjectsInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input listProjectsInput) (*mcp.CallToolResult, any, error) {
		var path string
		if input.CursusID > 0 {
			path = fmt.Sprintf("/cursus/%d/projects", input.CursusID)
		} else {
			path = "/projects"
		}
		data, err := c.Get(path, paginationParams(input.Page, input.PerPage))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

type listEventsInput struct {
	CampusID int `json:"campus_id,omitempty" jsonschema:"filter events by campus ID (optional)"`
	Page     int `json:"page,omitempty"      jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty"  jsonschema:"results per page, max 100"`
}

func handleListEvents(c *client.Client) func(context.Context, *mcp.CallToolRequest, listEventsInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input listEventsInput) (*mcp.CallToolResult, any, error) {
		var path string
		if input.CampusID > 0 {
			path = fmt.Sprintf("/campus/%d/events", input.CampusID)
		} else {
			path = "/events"
		}
		data, err := c.Get(path, paginationParams(input.Page, input.PerPage))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

func registerProjects(s *mcp.Server, c *client.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_cursus",
		Description: "List all available cursus (curricula) on the 42 network",
	}, handleListCursus(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_projects",
		Description: "List projects, optionally filtered to a specific cursus",
	}, handleListProjects(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_events",
		Description: "List upcoming events, optionally filtered by campus",
	}, handleListEvents(c))
}
