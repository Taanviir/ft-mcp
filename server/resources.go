package tools

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerResources(s *mcp.Server) {
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "42://users/{login}",
		Name:        "42 User Profile",
		Description: "Full profile for a 42 student: level, grade, skills, campus, and cursus info.",
		MIMEType:    "application/json",
	}, handleUserResource)

	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "42://campus/{campus_id}",
		Name:        "42 Campus",
		Description: "Campus details including name, country, city, and student count.",
		MIMEType:    "application/json",
	}, handleCampusResource)
}

func handleUserResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// URI format: 42://users/{login}
	login := strings.TrimPrefix(req.Params.URI, "42://users/")
	if login == "" {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	data, err := api.Get("/users/"+login, nil)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(filterJSON[ftUser](data)),
		}},
	}, nil
}

func handleCampusResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// URI format: 42://campus/{campus_id}
	id := strings.TrimPrefix(req.Params.URI, "42://campus/")
	if id == "" {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	data, err := api.Get("/campus/"+id, nil)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(filterJSON[ftCampusFull](data)),
		}},
	}, nil
}
