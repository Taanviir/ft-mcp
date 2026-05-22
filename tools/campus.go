package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/mcp42/client"
)

type listCampusInput struct {
	Page    int `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage int `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

func handleListCampus(c *client.Client) func(context.Context, *mcp.CallToolRequest, listCampusInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input listCampusInput) (*mcp.CallToolResult, any, error) {
		data, err := c.Get("/campus", paginationParams(input.Page, input.PerPage))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

type campusUsersInput struct {
	CampusID int `json:"campus_id"          jsonschema:"numeric ID of the campus"`
	Page     int `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

func handleGetCampusUsers(c *client.Client) func(context.Context, *mcp.CallToolRequest, campusUsersInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input campusUsersInput) (*mcp.CallToolResult, any, error) {
		data, err := c.Get(fmt.Sprintf("/campus/%d/users", input.CampusID), paginationParams(input.Page, input.PerPage))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

type locationsInput struct {
	CampusID int `json:"campus_id,omitempty" jsonschema:"filter by campus ID (optional)"`
	Page     int `json:"page,omitempty"      jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty"  jsonschema:"results per page, max 100"`
}

func handleGetLocations(c *client.Client) func(context.Context, *mcp.CallToolRequest, locationsInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input locationsInput) (*mcp.CallToolResult, any, error) {
		var path string
		if input.CampusID > 0 {
			path = fmt.Sprintf("/campus/%d/locations", input.CampusID)
		} else {
			path = "/locations"
		}
		data, err := c.Get(path, paginationParams(input.Page, input.PerPage))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

func registerCampus(s *mcp.Server, c *client.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_campus",
		Description: "List all 42 campuses worldwide",
	}, handleListCampus(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_campus_users",
		Description: "Get users enrolled at a specific campus",
	}, handleGetCampusUsers(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_locations",
		Description: "Get active campus locations — which users are currently logged into a computer. Filter by campus_id to scope to one campus.",
	}, handleGetLocations(c))
}
