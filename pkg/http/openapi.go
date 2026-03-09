package http

import (
	"encoding/json"
	"net/http"

	"k8s.io/klog/v2"

	"github.com/containers/kubernetes-mcp-server/pkg/mcp"
	"github.com/containers/kubernetes-mcp-server/pkg/version"
)

const (
	docsEndpoint    = "/docs"
	openAPIEndpoint = "/openapi.json"
)

// openAPISpec represents an OpenAPI 3.0 specification
type openAPISpec struct {
	OpenAPI    string                `json:"openapi"`
	Info       openAPIInfo           `json:"info"`
	Paths      map[string]pathItem   `json:"paths"`
	Components *openAPIComponents    `json:"components,omitempty"`
	Servers    []openAPIServer       `json:"servers,omitempty"`
}

type openAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

type openAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type pathItem struct {
	Post *operation `json:"post,omitempty"`
}

type operation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	OperationID string              `json:"operationId,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	RequestBody *requestBody        `json:"requestBody,omitempty"`
	Responses   map[string]response `json:"responses"`
}

type requestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                  `json:"required"`
	Content     map[string]mediaType `json:"content"`
}

type mediaType struct {
	Schema any `json:"schema"`
}

type response struct {
	Description string               `json:"description"`
	Content     map[string]mediaType `json:"content,omitempty"`
}

type openAPIComponents struct {
	Schemas map[string]any `json:"schemas,omitempty"`
}

// openAPIHandler returns an HTTP handler that serves the OpenAPI 3.0 specification
// generated from the currently registered MCP tools.
func openAPIHandler(mcpServer *mcp.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		withCORSHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		spec := buildOpenAPISpec(mcpServer)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(spec); err != nil {
			klog.V(1).Infof("Failed to encode OpenAPI spec: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func buildOpenAPISpec(mcpServer *mcp.Server) openAPISpec {
	tools := mcpServer.GetServerTools()

	paths := make(map[string]pathItem, len(tools))
	for _, tool := range tools {
		path := "/mcp/tools/" + tool.Tool.Name

		var reqBody *requestBody
		if tool.Tool.InputSchema != nil {
			// Marshal and re-parse the jsonschema.Schema to get a clean map for the OpenAPI spec
			schemaJSON, err := json.Marshal(tool.Tool.InputSchema)
			if err == nil {
				var schemaMap any
				if err := json.Unmarshal(schemaJSON, &schemaMap); err == nil {
					reqBody = &requestBody{
						Required: true,
						Content: map[string]mediaType{
							"application/json": {Schema: schemaMap},
						},
					}
				}
			}
		}

		tags := []string{"tools"}

		description := tool.Tool.Description
		if tool.Tool.Annotations.Title != "" {
			description = tool.Tool.Annotations.Title + "\n\n" + description
		}

		paths[path] = pathItem{
			Post: &operation{
				Summary:     tool.Tool.Annotations.Title,
				Description: description,
				OperationID: tool.Tool.Name,
				Tags:        tags,
				RequestBody: reqBody,
				Responses: map[string]response{
					"200": {
						Description: "Tool execution result",
						Content: map[string]mediaType{
							"application/json": {
								Schema: map[string]any{
									"type": "object",
									"properties": map[string]any{
										"content": map[string]any{
											"type": "array",
											"items": map[string]any{
												"type": "object",
												"properties": map[string]any{
													"type": map[string]any{"type": "string"},
													"text": map[string]any{"type": "string"},
												},
											},
										},
										"isError": map[string]any{"type": "boolean"},
									},
								},
							},
						},
					},
				},
			},
		}
	}

	return openAPISpec{
		OpenAPI: "3.0.3",
		Info: openAPIInfo{
			Title:       version.BinaryName,
			Description: "Kubernetes MCP Server - Model Context Protocol server for Kubernetes cluster management. Tools can be invoked via the MCP protocol at /mcp or via the REST API at /mcp/tools/{name}.",
			Version:     version.Version,
		},
		Paths: paths,
	}
}
