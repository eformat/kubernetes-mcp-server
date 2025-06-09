package mcp

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initNamespaces() []server.ServerTool {
	ret := make([]server.ServerTool, 0)
	ret = append(ret, server.ServerTool{
		Tool: mcp.NewTool("namespaces_list",
			mcp.WithDescription("List all of the Kubernetes namespaces in the current cluster"),
		), Handler: s.namespacesList,
	})
	ret = append(ret, server.ServerTool{
		Tool: mcp.NewTool("namespace_get",
			mcp.WithDescription("Get a single Kubernetes namespace in the current cluster"),
			mcp.WithString("namespace",
				mcp.Required(),
				mcp.Description("Name of the namespace to get"),
			),
		), Handler: s.namespaceGet,
	})
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
	if s.k.IsOpenShift(context.Background()) {
		ret = append(ret, server.ServerTool{
			Tool: mcp.NewTool("projects_list",
				mcp.WithDescription("List all the OpenShift projects in the current cluster"),
			), Handler: s.projectsList,
		})
		ret = append(ret, server.ServerTool{
			Tool: mcp.NewTool("project_get",
				mcp.WithDescription("Get a single OpenShift project in the current cluster"),
				mcp.WithString("namespace",
					mcp.Required(),
					mcp.Description("Name of the namespace to get"),
				),
			), Handler: s.namespaceGet,
		})
	}
	return ret
}

func (s *Server) namespacesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ret, err := s.k.NamespacesList(ctx)
	if err != nil {
		err = fmt.Errorf("failed to list namespaces: %v", err)
	}
	return NewTextResult(ret, err), nil
}

func (s *Server) namespaceGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"]
	if ns == nil {
		return NewTextResult("", fmt.Errorf("failed to get namespace missing a namespace name")), nil
	}
	ret, err := s.k.NamespaceGet(ctx, ns.(string))
	if err != nil {
		err = fmt.Errorf("failed to get namespace: %v", err)
	}
	return NewTextResult(ret, err), nil
}

func (s *Server) projectsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ret, err := s.k.ProjectsList(ctx)
	if err != nil {
		err = fmt.Errorf("failed to list projects: %v", err)
	}
	return NewTextResult(ret, err), nil
}

func (s *Server) projectGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"]
	if ns == nil {
		return NewTextResult("", fmt.Errorf("failed to get project missing a project name")), nil
	}
	ret, err := s.k.ProjectGet(ctx, ns.(string))
	if err != nil {
		err = fmt.Errorf("failed to get project: %v", err)
	}
	return NewTextResult(ret, err), nil
}

func (s *Server) namespaceCreate(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"]
	if ns == nil {
		return NewTextResult("", fmt.Errorf("failed to create namespace missing a namespace name")), nil
	}
	ret, err := s.k.NamespaceCreate(ctx, ns.(string), metav1.CreateOptions{})
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create namespace %s ", ns, err)), nil
	}
	return NewTextResult(ret.Name, err), nil
}

func (s *Server) namespaceDelete(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"]
	if ns == nil {
		return NewTextResult("", fmt.Errorf("failed to delete namespace missing a namespace name")), nil
	}
	err := s.k.NamespaceDelete(ctx, ns.(string), metav1.DeleteOptions{})
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to delete namespace %s ", ns, err)), nil
	}
	return NewTextResult(ns.(string), err), nil
}
