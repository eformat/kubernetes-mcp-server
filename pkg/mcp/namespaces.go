package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/containers/kubernetes-mcp-server/pkg/kubernetes"
)

func (s *Server) initNamespaces() []server.ServerTool {
	ret := make([]server.ServerTool, 0)
	ret = append(ret, server.ServerTool{
		Tool: mcp.NewTool("namespaces_list",
			mcp.WithDescription("List all the Kubernetes namespaces in the current cluster"),
			// Tool annotations
			mcp.WithTitleAnnotation("Namespaces: List"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(true),
		), Handler: s.namespacesList,
	})
	if s.k.IsOpenShift(context.Background()) {
		ret = append(ret, server.ServerTool{
			Tool: mcp.NewTool("projects_list",
				mcp.WithDescription("List all the OpenShift projects in the current cluster"),
				// Tool annotations
				mcp.WithTitleAnnotation("Projects: List"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
				mcp.WithOpenWorldHintAnnotation(true),
			), Handler: s.projectsList,
		})
	}
	ret = append(ret, server.ServerTool{
		Tool: mcp.NewTool("namespace_create",
			mcp.WithDescription("Create the Kubernetes namespace in the current cluster"),
			mcp.WithString("namespace",
				mcp.Required(),
				mcp.Description("Name of the namespace to create"),
			),
		), Handler: s.namespaceCreate,
	})
	ret = append(ret, server.ServerTool{
		Tool: mcp.NewTool("namespace_delete",
			mcp.WithDescription("Delete the Kubernetes namespace in the current cluster"),
			mcp.WithString("namespace",
				mcp.Required(),
				mcp.Description("Name of the namespace to delete"),
			),
		), Handler: s.namespaceDelete,
	})
	return ret
}

func (s *Server) namespacesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	derived, err := s.k.Derived(ctx)
	if err != nil {
		return nil, err
	}
	ret, err := derived.NamespacesList(ctx, kubernetes.ResourceListOptions{AsTable: s.configuration.ListOutput.AsTable()})
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to list namespaces: %v", err)), nil
	}
	return NewTextResult(s.configuration.ListOutput.PrintObj(ret)), nil
}

func (s *Server) projectsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	derived, err := s.k.Derived(ctx)
	if err != nil {
		return nil, err
	}
	ret, err := derived.ProjectsList(ctx, kubernetes.ResourceListOptions{AsTable: s.configuration.ListOutput.AsTable()})
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to list projects: %v", err)), nil
	}
	return NewTextResult(s.configuration.ListOutput.PrintObj(ret)), nil
}

func (s *Server) namespaceCreate(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	derived, err := s.k.Derived(ctx)
	if err != nil {
		return nil, err
	}
	ns := ctr.GetArguments()["namespace"]
	if ns == nil {
		return NewTextResult("", fmt.Errorf("failed to create namespace missing a namespace name")), nil
	}
	ret, err := derived.NamespaceCreate(ctx, ns.(string))
	if ret == nil {
		return NewTextResult("", fmt.Errorf("failed to create namespace %s ", ns)), err
	}
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create namespace %s ", ns)), err
	}
	return NewTextResult(ns.(string), err), nil
}

func (s *Server) namespaceDelete(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	derived, err := s.k.Derived(ctx)
	if err != nil {
		return nil, err
	}
	ns := ctr.GetArguments()["namespace"]
	if ns == nil {
		return NewTextResult("", fmt.Errorf("failed to delete namespace missing a namespace name")), nil
	}
	ret, err := derived.NamespaceDelete(ctx, ns.(string))
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to delete namespace %s ", ns)), err
	}
	return NewTextResult(ret, err), nil
}

