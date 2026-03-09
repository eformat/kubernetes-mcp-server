package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RESTToolsSuite struct {
	BaseHttpSuite
}

func TestRESTTools(t *testing.T) {
	suite.Run(t, new(RESTToolsSuite))
}

func (s *RESTToolsSuite) TestCallTool() {
	s.StartServer()
	s.Run("calls a tool and returns result", func() {
		body := `{"namespace": "default"}`
		resp, err := http.Post(
			fmt.Sprintf("http://0.0.0.0:%s/mcp/tools/pods_list", s.StaticConfig.Port),
			"application/json",
			strings.NewReader(body),
		)
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
		s.Contains(resp.Header.Get("Content-Type"), "application/json")

		respBody, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		var result map[string]any
		s.Require().NoError(json.Unmarshal(respBody, &result))
		_, hasContent := result["content"]
		s.True(hasContent, "response should have 'content' field")
		_, hasIsError := result["isError"]
		s.True(hasIsError, "response should have 'isError' field")
	})
	s.Run("returns error for unknown tool", func() {
		resp, err := http.Post(
			fmt.Sprintf("http://0.0.0.0:%s/mcp/tools/nonexistent_tool", s.StaticConfig.Port),
			"application/json",
			strings.NewReader("{}"),
		)
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusInternalServerError, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		var result map[string]any
		s.Require().NoError(json.Unmarshal(respBody, &result))
		s.True(result["isError"].(bool))
	})
	s.Run("returns error for missing tool name", func() {
		resp, err := http.Post(
			fmt.Sprintf("http://0.0.0.0:%s/mcp/tools/", s.StaticConfig.Port),
			"application/json",
			strings.NewReader("{}"),
		)
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
	s.Run("returns error for invalid JSON body", func() {
		resp, err := http.Post(
			fmt.Sprintf("http://0.0.0.0:%s/mcp/tools/pods_list", s.StaticConfig.Port),
			"application/json",
			strings.NewReader("not json"),
		)
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
	s.Run("rejects non-POST methods", func() {
		resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:%s/mcp/tools/pods_list", s.StaticConfig.Port))
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
	})
	s.Run("responds to OPTIONS preflight", func() {
		req, err := http.NewRequest("OPTIONS", fmt.Sprintf("http://0.0.0.0:%s/mcp/tools/pods_list", s.StaticConfig.Port), nil)
		s.Require().NoError(err)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusNoContent, resp.StatusCode)
		s.Equal("*", resp.Header.Get("Access-Control-Allow-Origin"))
		s.Equal("POST, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	})
}
