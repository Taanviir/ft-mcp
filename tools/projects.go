package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listCursusInput struct {
	Page    int `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage int `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

func handleListCursus(_ context.Context, _ *mcp.CallToolRequest, input listCursusInput) (*mcp.CallToolResult, any, error) {
	data, err := api.Get("/cursus", paginationParams(input.Page, input.PerPage))
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(data), nil, nil
}

type listProjectsInput struct {
	CursusID int `json:"cursus_id,omitempty" jsonschema:"filter projects by cursus ID (optional)"`
	Page     int `json:"page,omitempty"      jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty"  jsonschema:"results per page, max 100"`
}

func handleListProjects(_ context.Context, _ *mcp.CallToolRequest, input listProjectsInput) (*mcp.CallToolResult, any, error) {
	path := "/projects"
	if input.CursusID > 0 {
		path = fmt.Sprintf("/cursus/%d/projects", input.CursusID)
	}
	data, err := api.Get(path, paginationParams(input.Page, input.PerPage))
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(data), nil, nil
}

type listEventsInput struct {
	CampusID int `json:"campus_id,omitempty" jsonschema:"filter events by campus ID (optional)"`
	Page     int `json:"page,omitempty"      jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty"  jsonschema:"results per page, max 100"`
}

func handleListEvents(_ context.Context, _ *mcp.CallToolRequest, input listEventsInput) (*mcp.CallToolResult, any, error) {
	path := "/events"
	if input.CampusID > 0 {
		path = fmt.Sprintf("/campus/%d/events", input.CampusID)
	}
	data, err := api.Get(path, paginationParams(input.Page, input.PerPage))
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(data), nil, nil
}

func registerProjects(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_cursus",
		Description: "List all available cursus (curricula) on the 42 network",
	}, handleListCursus)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_projects",
		Description: "List projects, optionally filtered to a specific cursus",
	}, handleListProjects)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_events",
		Description: "List upcoming events, optionally filtered by campus",
	}, handleListEvents)
}
