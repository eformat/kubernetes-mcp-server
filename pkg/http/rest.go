package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"k8s.io/klog/v2"

	"github.com/containers/kubernetes-mcp-server/pkg/mcp"
)

const toolsRESTEndpoint = "/mcp/tools/"

// toolsRESTHandler returns an HTTP handler that provides a REST API proxy
// for MCP tools. It translates REST calls (POST /mcp/tools/{name} with JSON body)
// into direct tool invocations, making tools callable from Swagger UI and other
// REST clients.
func toolsRESTHandler(mcpServer *mcp.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		toolName := strings.TrimPrefix(r.URL.Path, toolsRESTEndpoint)
		if toolName == "" {
			http.Error(w, "Tool name is required", http.StatusBadRequest)
			return
		}

		var arguments map[string]any
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&arguments); err != nil {
				http.Error(w, "Invalid JSON request body: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		result, err := mcpServer.CallTool(r.Context(), toolName, arguments)
		if err != nil {
			klog.V(1).Infof("REST tool call %q failed: %v", toolName, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"isError": true,
				"content": []map[string]any{
					{"type": "text", "text": err.Error()},
				},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")

		resp := map[string]any{
			"isError": result.Error != nil,
		}
		if result.Error != nil {
			resp["content"] = []map[string]any{
				{"type": "text", "text": result.Error.Error()},
			}
		} else {
			resp["content"] = []map[string]any{
				{"type": "text", "text": result.Content},
			}
		}
		if result.StructuredContent != nil {
			resp["structuredContent"] = result.StructuredContent
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			klog.V(1).Infof("Failed to encode REST tool response: %v", err)
		}
	}
}
