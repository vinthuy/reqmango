// Package mcp assembles the reqmango MCP server on top of the shared client.
package mcp

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/reqmango/tools/client"
)

// ServerName and ServerVersion identify this MCP server to clients.
const (
	ServerName    = "reqmango-mcp"
	ServerVersion = "1.0.0"
)

// New creates the reqmango MCP server backed by cli. Tool registration is
// split across registerCoreTools (19 core) and registerAITools (5 AI).
func New(cli *client.Client) *server.MCPServer {
	s := server.NewMCPServer(ServerName, ServerVersion,
		server.WithInputSchemaValidation(),
		server.WithToolCapabilities(false), // required by mcp-go v1.0.0 for tools/list+calls to work at all
	)
	registerCoreTools(s, cli)
	return s
}
