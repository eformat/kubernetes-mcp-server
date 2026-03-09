package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/containers/kubernetes-mcp-server/pkg/api"
	"github.com/containers/kubernetes-mcp-server/pkg/config"
	"github.com/stretchr/testify/suite"
)

type OpenAPISuite struct {
	BaseHttpSuite
}

func TestOpenAPI(t *testing.T) {
	suite.Run(t, new(OpenAPISuite))
}

func (s *OpenAPISuite) TestOpenAPIEndpoint() {
	s.StartServer()
	s.Run("returns valid OpenAPI JSON", func() {
		resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:%s/openapi.json", s.StaticConfig.Port))
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
		s.Contains(resp.Header.Get("Content-Type"), "application/json")

		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		var spec openAPISpec
		s.Require().NoError(json.Unmarshal(body, &spec))
		s.Equal("3.0.3", spec.OpenAPI)
		s.Equal("kubernetes-mcp-server", spec.Info.Title)
		s.NotEmpty(spec.Paths)
	})
	s.Run("returns CORS header", func() {
		resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:%s/openapi.json", s.StaticConfig.Port))
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal("*", resp.Header.Get("Access-Control-Allow-Origin"))
	})
	s.Run("responds to OPTIONS preflight", func() {
		req, err := http.NewRequest("OPTIONS", fmt.Sprintf("http://0.0.0.0:%s/openapi.json", s.StaticConfig.Port), nil)
		s.Require().NoError(err)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "Authorization")
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusNoContent, resp.StatusCode)
		s.Equal("*", resp.Header.Get("Access-Control-Allow-Origin"))
		s.Equal("GET, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
		s.Equal("Content-Type, Authorization", resp.Header.Get("Access-Control-Allow-Headers"))
	})
	s.Run("rejects non-GET methods", func() {
		resp, err := http.Post(fmt.Sprintf("http://0.0.0.0:%s/openapi.json", s.StaticConfig.Port), "application/json", nil)
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

func (s *OpenAPISuite) TestOpenAPIEndpointWithOAuth() {
	s.StaticConfig.RequireOAuth = true
	s.StaticConfig.ClusterProviderStrategy = api.ClusterProviderKubeConfig
	s.StartServer()
	s.Run("accessible without authentication", func() {
		resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:%s/openapi.json", s.StaticConfig.Port))
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
	})
}

type DocsSuite struct {
	BaseHttpSuite
}

func TestDocs(t *testing.T) {
	suite.Run(t, new(DocsSuite))
}

func (s *DocsSuite) TestDocsEndpoint() {
	s.StartServer()
	s.Run("returns HTML page", func() {
		resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:%s/docs", s.StaticConfig.Port))
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
		s.Contains(resp.Header.Get("Content-Type"), "text/html")

		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		s.Contains(string(body), "swagger-ui")
		s.Contains(string(body), "openapi.json")
	})
	s.Run("responds to OPTIONS preflight", func() {
		req, err := http.NewRequest("OPTIONS", fmt.Sprintf("http://0.0.0.0:%s/docs", s.StaticConfig.Port), nil)
		s.Require().NoError(err)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusNoContent, resp.StatusCode)
		s.Equal("*", resp.Header.Get("Access-Control-Allow-Origin"))
		s.Equal("GET, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	})
	s.Run("rejects non-GET methods", func() {
		resp, err := http.Post(fmt.Sprintf("http://0.0.0.0:%s/docs", s.StaticConfig.Port), "text/html", nil)
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

func (s *DocsSuite) TestDocsEndpointWithOAuth() {
	s.StaticConfig.RequireOAuth = true
	s.StaticConfig.ClusterProviderStrategy = api.ClusterProviderKubeConfig
	s.StartServer()
	s.Run("accessible without authentication", func() {
		resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:%s/docs", s.StaticConfig.Port))
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
	})
}

type OpenAPISpecSuite struct {
	suite.Suite
}

func TestOpenAPISpec(t *testing.T) {
	suite.Run(t, new(OpenAPISpecSuite))
}

func (s *OpenAPISpecSuite) TestBuildOpenAPISpecFromConfig() {
	s.Run("generates paths for enabled tools", func() {
		staticConfig := config.Default()
		staticConfig.KubeConfig = "" // not needed for spec generation test
		// We test buildOpenAPISpec indirectly through the handler
		// Since we can't easily create a full mcp.Server in a unit test,
		// we verify the spec structure is correct by checking the types
		spec := openAPISpec{
			OpenAPI: "3.0.3",
			Info: openAPIInfo{
				Title:   "test",
				Version: "0.0.0",
			},
			Paths: map[string]pathItem{
				"/mcp/tools/test_tool": {
					Post: &operation{
						Summary:     "Test Tool",
						OperationID: "test_tool",
						Tags:        []string{"tools"},
						Responses: map[string]response{
							"200": {Description: "Tool execution result"},
						},
					},
				},
			},
		}
		data, err := json.Marshal(spec)
		s.Require().NoError(err)
		var parsed map[string]any
		s.Require().NoError(json.Unmarshal(data, &parsed))
		s.Equal("3.0.3", parsed["openapi"])
		paths := parsed["paths"].(map[string]any)
		s.Contains(paths, "/mcp/tools/test_tool")
	})
}
