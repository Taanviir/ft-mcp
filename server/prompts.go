package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (t *tools) registerPrompts(s *mcp.Server) {
	s.AddPrompt(&mcp.Prompt{
		Name:        "analyze_student",
		Description: "Fetch and analyze a 42 student's full profile, projects, and skills",
		Arguments: []*mcp.PromptArgument{
			{Name: "login", Description: "The student's 42 login", Required: true},
		},
	}, t.handleAnalyzeStudent)

	s.AddPrompt(&mcp.Prompt{
		Name:        "campus_overview",
		Description: "Summarize activity and stats for a 42 campus",
		Arguments: []*mcp.PromptArgument{
			{Name: "campus_id", Description: "Numeric campus ID (use list_campus to find it)", Required: true},
		},
	}, t.handleCampusOverview)
}

func (t *tools) handleAnalyzeStudent(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	login := req.Params.Arguments["login"]
	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Analyze 42 student: %s", login),
		Messages: []*mcp.PromptMessage{{
			Role: "user",
			Content: &mcp.TextContent{Text: fmt.Sprintf(
				`Please analyze the 42 student with login "%s".

Call tools sequentially (not in parallel) to avoid hitting the 42 API rate limit:
1. get_user — already includes level, grade, campus, pool cohort, and cursus info. Sufficient for most profiles; only call get_user_cursus additionally if you need the per-skill breakdown.
2. get_user_projects — project submissions and scores.
3. get_user_achievements — badges and achievements.

Then provide:
1. Basic info (campus, pool cohort, current status)
2. Academic progress (level, grade, active cursus)
3. Top skills by level
4. Completed projects — highlight any perfect or bonus scores
5. In-progress or searching-group projects
6. Notable achievements
7. A brief overall assessment

Be concise and use the data directly — no need to re-fetch anything already shown.`, login),
			},
		}},
	}, nil
}

func (t *tools) handleCampusOverview(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	campusID := req.Params.Arguments["campus_id"]
	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Overview of campus %s", campusID),
		Messages: []*mcp.PromptMessage{{
			Role: "user",
			Content: &mcp.TextContent{Text: fmt.Sprintf(
				`Please provide an overview of 42 campus ID %s.

Use the available tools to fetch and summarize:
1. Campus details (name, city, country, total students)
2. Currently active students (logged into a computer right now)
3. Upcoming events at this campus
4. A brief summary of campus activity`, campusID),
			},
		}},
	}, nil
}
