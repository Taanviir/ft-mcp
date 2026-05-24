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
	return textResult(filterJSON[[]ftCursus](data)), nil, nil
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
	return textResult(filterJSON[[]ftProject](data)), nil, nil
}

type searchProjectsInput struct {
	Name    string `json:"name"               jsonschema:"project name to search for (e.g. ft_transcendence)"`
	Page    int    `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

func handleSearchProjects(_ context.Context, _ *mcp.CallToolRequest, input searchProjectsInput) (*mcp.CallToolResult, any, error) {
	params := paginationParams(input.Page, input.PerPage)
	params.Set("filter[name]", input.Name)
	data, err := api.Get("/projects", params)
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(filterJSON[[]ftProject](data)), nil, nil
}

type listProjectSubmissionsInput struct {
	ProjectID     int    `json:"project_id"               jsonschema:"numeric ID of the project (use search_projects to find it)"`
	CampusID      int    `json:"campus_id,omitempty"      jsonschema:"filter by campus ID (optional)"`
	OnlyValidated bool   `json:"only_validated,omitempty" jsonschema:"if true, return only validated (passed) submissions"`
	DateFrom      string `json:"date_from,omitempty"      jsonschema:"return submissions marked on or after this date (YYYY-MM-DD)"`
	DateTo        string `json:"date_to,omitempty"        jsonschema:"return submissions marked on or before this date (YYYY-MM-DD)"`
	Page          int    `json:"page,omitempty"           jsonschema:"page number, starting at 1"`
	PerPage       int    `json:"per_page,omitempty"       jsonschema:"results per page, max 100"`
}

func handleListProjectSubmissions(_ context.Context, _ *mcp.CallToolRequest, input listProjectSubmissionsInput) (*mcp.CallToolResult, any, error) {
	params := paginationParams(input.Page, input.PerPage)
	params.Set("filter[project_id]", fmt.Sprintf("%d", input.ProjectID))
	if input.CampusID > 0 {
		params.Set("filter[campus_id]", fmt.Sprintf("%d", input.CampusID))
	}
	if input.OnlyValidated {
		params.Set("filter[validated?]", "true")
	}
	if input.DateFrom != "" || input.DateTo != "" {
		params.Set("range[marked_at]", input.DateFrom+","+input.DateTo)
	}
	data, err := api.Get("/projects_users", params)
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(filterJSON[[]ftProjectUser](data)), nil, nil
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
	return textResult(filterJSON[[]ftEvent](data)), nil, nil
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
		Name:        "search_projects",
		Description: "Search projects by name to get their numeric ID — use this before list_project_submissions",
	}, handleSearchProjects)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_project_submissions",
		Description: "Get all student submissions for a project. Filter by campus, validated status, and date range. Much more efficient than checking users one by one.",
	}, handleListProjectSubmissions)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_events",
		Description: "List upcoming events, optionally filtered by campus",
	}, handleListEvents)
}
