package tools

import (
	"context"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tanas/mcp42/client"
)

type getUserInput struct {
	LoginOrID string `json:"login_or_id" jsonschema:"login name or numeric ID of the 42 user"`
}

func handleGetUser(c *client.Client) func(context.Context, *mcp.CallToolRequest, getUserInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input getUserInput) (*mcp.CallToolResult, any, error) {
		data, err := c.Get("/users/"+input.LoginOrID, nil)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

type listUsersInput struct {
	CampusID int `json:"campus_id,omitempty" jsonschema:"filter by campus ID (optional)"`
	Page     int `json:"page,omitempty"      jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty"  jsonschema:"results per page, max 100"`
}

func handleListUsers(c *client.Client) func(context.Context, *mcp.CallToolRequest, listUsersInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input listUsersInput) (*mcp.CallToolResult, any, error) {
		params := paginationParams(input.Page, input.PerPage)
		if input.CampusID > 0 {
			params.Set("filter[campus_id]", fmt.Sprintf("%d", input.CampusID))
		}
		data, err := c.Get("/users", params)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

type userSubInput struct {
	LoginOrID string `json:"login_or_id" jsonschema:"login name or numeric ID of the 42 user"`
	Page      int    `json:"page,omitempty" jsonschema:"page number, starting at 1"`
}

func handleGetUserCursus(c *client.Client) func(context.Context, *mcp.CallToolRequest, userSubInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input userSubInput) (*mcp.CallToolResult, any, error) {
		data, err := c.Get("/users/"+input.LoginOrID+"/cursus_users", paginationParams(input.Page, 0))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

func handleGetUserProjects(c *client.Client) func(context.Context, *mcp.CallToolRequest, userSubInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input userSubInput) (*mcp.CallToolResult, any, error) {
		data, err := c.Get("/users/"+input.LoginOrID+"/projects_users", paginationParams(input.Page, 0))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

func handleGetUserAchievements(c *client.Client) func(context.Context, *mcp.CallToolRequest, userSubInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input userSubInput) (*mcp.CallToolResult, any, error) {
		data, err := c.Get("/users/"+input.LoginOrID+"/achievements", paginationParams(input.Page, 0))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

func registerUsers(s *mcp.Server, c *client.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_user",
		Description: "Get a 42 user profile by login name or numeric ID",
	}, handleGetUser(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_users",
		Description: "List 42 users, optionally filtered by campus ID",
	}, handleListUsers(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_user_cursus",
		Description: "Get a user's cursus information including level, grade, and skills",
	}, handleGetUserCursus(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_user_projects",
		Description: "Get a user's project submissions and their status (finished, in_progress, etc.)",
	}, handleGetUserProjects(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_user_achievements",
		Description: "Get a user's achievements (badges) on the 42 platform",
	}, handleGetUserAchievements(c))
}

func paginationParams(page, perPage int) url.Values {
	p := url.Values{}
	if page > 1 {
		p.Set("page[number]", fmt.Sprintf("%d", page))
	}
	if perPage > 0 {
		p.Set("page[size]", fmt.Sprintf("%d", perPage))
	}
	return p
}
