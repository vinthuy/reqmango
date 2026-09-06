// Command reqmango-mcp is the reqmango MCP server.
//
// Stdio transport (default, for Claude Code / Claude Desktop / Cursor):
//
//	REQMANGO_PAT=reqmango_pat_xxx reqmango-mcp
//
// Streamable HTTP transport (for remote / CI):
//
//	REQMANGO_PAT=reqmango_pat_xxx reqmango-mcp --http :8080
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/reqmango/tools/client"
	reqmcp "github.com/reqmango/tools/mcp"
)

func main() {
	httpAddr := flag.String("http", "", "serve streamable HTTP on this address (e.g. :8080) instead of stdio")
	flag.Parse()

	apiURL := os.Getenv("REQMANGO_API_URL")
	if apiURL == "" {
		apiURL = client.DefaultBaseURL
	}
	pat := os.Getenv("REQMANGO_PAT")
	if pat == "" {
		fmt.Fprintln(os.Stderr, "ERROR: REQMANGO_PAT environment variable is required (create one with `reqmango auth login`)")
		os.Exit(1)
	}

	s := reqmcp.New(client.New(apiURL, pat))

	if *httpAddr != "" {
		h := server.NewStreamableHTTPServer(s,
			server.WithEndpointPath("/mcp"),
			server.WithStateful(true),
			server.WithDisableLocalhostProtection(true), // --http is an explicit opt-in for remote serving
		)
		log.Printf("reqmango-mcp listening on %s (endpoint /mcp, Bearer PAT required)", *httpAddr)
		log.Fatal(http.ListenAndServe(*httpAddr, reqmcp.BearerAuth(pat, h)))
	}

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
