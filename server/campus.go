package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listCampusInput struct {
	Page    int `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage int `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

func (t *tools) handleListCampus(ctx context.Context, _ *mcp.CallToolRequest, input listCampusInput) (*mcp.CallToolResult, any, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}
	data, err := client.Get("/campus", paginationParams(input.Page, input.PerPage))
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(filterJSON[[]ftCampusFull](data)), nil, nil
}

type campusUsersInput struct {
	CampusID int `json:"campus_id"          jsonschema:"numeric ID of the campus"`
	Page     int `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

func (t *tools) handleGetCampusUsers(ctx context.Context, _ *mcp.CallToolRequest, input campusUsersInput) (*mcp.CallToolResult, any, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}
	data, err := client.Get(fmt.Sprintf("/campus/%d/users", input.CampusID), paginationParams(input.Page, input.PerPage))
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(filterJSON[[]ftUserMin](data)), nil, nil
}

type locationsInput struct {
	CampusID int `json:"campus_id,omitempty" jsonschema:"filter by campus ID (optional)"`
	Page     int `json:"page,omitempty"      jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty"  jsonschema:"results per page, max 100"`
}

func (t *tools) handleGetLocations(ctx context.Context, _ *mcp.CallToolRequest, input locationsInput) (*mcp.CallToolResult, any, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}
	path := "/locations"
	if input.CampusID > 0 {
		path = fmt.Sprintf("/campus/%d/locations", input.CampusID)
	}
	data, err := client.Get(path, paginationParams(input.Page, input.PerPage))
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(filterJSON[[]ftLocation](data)), nil, nil
}

func (t *tools) registerCampus(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_campus",
		Description: "List all 42 campuses worldwide",
	}, t.handleListCampus)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_campus_users",
		Description: "Get users enrolled at a specific campus",
	}, t.handleGetCampusUsers)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_locations",
		Description: "Get active campus locations — which users are currently logged into a computer. Filter by campus_id to scope to one campus.",
	}, t.handleGetLocations)
}
