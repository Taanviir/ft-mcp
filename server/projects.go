package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listCursusInput struct {
	Page    int `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage int `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

func (t *tools) handleListCursus(ctx context.Context, _ *mcp.CallToolRequest, input listCursusInput) (*mcp.CallToolResult, any, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}
	data, err := client.Get("/cursus", paginationParams(input.Page, input.PerPage))
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

func (t *tools) handleListProjects(ctx context.Context, _ *mcp.CallToolRequest, input listProjectsInput) (*mcp.CallToolResult, any, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}
	path := "/projects"
	if input.CursusID > 0 {
		path = fmt.Sprintf("/cursus/%d/projects", input.CursusID)
	}
	data, err := client.Get(path, paginationParams(input.Page, input.PerPage))
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(filterJSON[[]ftProject](data)), nil, nil
}

type searchProjectsInput struct {
	Name    string `json:"name"               jsonschema:"partial or full project name to search for (e.g. transcendence, ft_printf)"`
	Page    int    `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

func (t *tools) handleSearchProjects(ctx context.Context, _ *mcp.CallToolRequest, input searchProjectsInput) (*mcp.CallToolResult, any, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}
	params := paginationParams(input.Page, input.PerPage)
	params.Set("search[name]", input.Name)
	data, err := client.Get("/projects", params)
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(filterJSON[[]ftProject](data)), nil, nil
}

type submissionFilters struct {
	ProjectID int    `json:"project_id"          jsonschema:"numeric ID of the project (use search_projects to find it)"`
	CampusID  int    `json:"campus_id,omitempty" jsonschema:"filter by campus ID (optional)"`
	DateFrom  string `json:"date_from,omitempty" jsonschema:"submissions marked on or after this date (YYYY-MM-DD)"`
	DateTo    string `json:"date_to,omitempty"   jsonschema:"submissions marked on or before this date (YYYY-MM-DD)"`
}

func submissionParams(f submissionFilters, page, perPage int) url.Values {
	params := paginationParams(page, perPage)
	params.Set("filter[project_id]", fmt.Sprintf("%d", f.ProjectID))
	if f.CampusID > 0 {
		params.Set("filter[campus_id]", fmt.Sprintf("%d", f.CampusID))
	}
	if f.DateFrom != "" || f.DateTo != "" {
		params.Set("range[marked_at]", f.DateFrom+","+f.DateTo)
	}
	return params
}

type listProjectSubmissionsInput struct {
	submissionFilters
	Page    int `json:"page,omitempty"     jsonschema:"page number, starting at 1"`
	PerPage int `json:"per_page,omitempty" jsonschema:"results per page, max 100"`
}

type submissionsPage struct {
	Total   *int            `json:"total,omitempty"`
	Results []ftProjectUser `json:"results"`
}

func (t *tools) handleListProjectSubmissions(ctx context.Context, _ *mcp.CallToolRequest, input listProjectSubmissionsInput) (*mcp.CallToolResult, any, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}
	params := submissionParams(input.submissionFilters, input.Page, input.PerPage)
	data, total, err := client.GetWithTotal("/projects_users", params)
	if err != nil {
		return errorResult(err), nil, nil
	}
	var results []ftProjectUser
	if err := json.Unmarshal(data, &results); err != nil {
		return textResult(data), nil, nil
	}
	resp := submissionsPage{Results: results}
	if total >= 0 {
		resp.Total = &total
	}
	out, err := json.Marshal(resp)
	if err != nil {
		return textResult(data), nil, nil
	}
	return textResult(out), nil, nil
}

type countProjectSubmissionsInput struct {
	submissionFilters
}

func (t *tools) handleCountProjectSubmissions(ctx context.Context, _ *mcp.CallToolRequest, input countProjectSubmissionsInput) (*mcp.CallToolResult, any, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}
	params := submissionParams(input.submissionFilters, 0, 0)
	total, err := client.Count("/projects_users", params)
	if err != nil {
		return errorResult(err), nil, nil
	}
	out, _ := json.Marshal(map[string]int{"total": total})
	return textResult(out), nil, nil
}

type listEventsInput struct {
	CampusID int `json:"campus_id,omitempty" jsonschema:"filter events by campus ID (optional)"`
	Page     int `json:"page,omitempty"      jsonschema:"page number, starting at 1"`
	PerPage  int `json:"per_page,omitempty"  jsonschema:"results per page, max 100"`
}

func (t *tools) handleListEvents(ctx context.Context, _ *mcp.CallToolRequest, input listEventsInput) (*mcp.CallToolResult, any, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}
	path := "/events"
	if input.CampusID > 0 {
		path = fmt.Sprintf("/campus/%d/events", input.CampusID)
	}
	data, err := client.Get(path, paginationParams(input.Page, input.PerPage))
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(filterJSON[[]ftEvent](data)), nil, nil
}

func (t *tools) registerProjects(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_cursus",
		Description: "List all available cursus (curricula) on the 42 network",
	}, t.handleListCursus)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all projects, optionally filtered to a specific cursus",
	}, t.handleListProjects)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_projects",
		Description: "Search projects by partial name to find their numeric ID. Supports partial matches (e.g. \"transcendence\" finds ft_transcendence). Use this before list_project_submissions or count_project_submissions.",
	}, t.handleSearchProjects)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_project_submissions",
		Description: "Get student submissions for a project, paginated (up to 100 per page). Response includes a total count so you know how many pages to expect. Each result has a \"validated?\" boolean and \"final_mark\" — filter on those client-side after fetching. Use count_project_submissions first if you only need the total without the records.",
	}, t.handleListProjectSubmissions)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "count_project_submissions",
		Description: "Get the total number of submissions for a project matching the given filters — without fetching the records. Use this before paginating list_project_submissions to know the full scope. Note: validated status cannot be pre-filtered; check the \"validated?\" field in results from list_project_submissions.",
	}, t.handleCountProjectSubmissions)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_events",
		Description: "List events, optionally filtered by campus",
	}, t.handleListEvents)
}
