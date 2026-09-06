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
	"net"
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
			// --http is an explicit opt-in for remote serving; the operator MUST terminate TLS in front
			// (see sdk/README.md) — non-loopback binds print a plaintext warning above.
			server.WithDisableLocalhostProtection(true),
		)
		if isNonLoopback(*httpAddr) {
			fmt.Fprintf(os.Stderr, "WARNING: serving MCP over plaintext HTTP on %s — every request carries your full-privilege PAT in the clear. Put a TLS-terminating reverse proxy in front of this endpoint, or bind a loopback address (e.g. 127.0.0.1:8080). See sdk/README.md \"HTTP 模式（远程 / CI）\".\n", *httpAddr)
		}
		log.Printf("reqmango-mcp listening on %s (endpoint /mcp, Bearer PAT required)", *httpAddr)
		log.Fatal(http.ListenAndServe(*httpAddr, reqmcp.BearerAuth(pat, h)))
	}

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// isNonLoopback reports whether addr binds to a non-loopback interface.
// A non-loopback bind exposes the plaintext-PAT HTTP endpoint beyond the
// local machine and must print the plaintext warning.
func isNonLoopback(addr string) bool {
	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	switch host {
	case "", "0.0.0.0", "::", "[::]":
		return true
	case "localhost", "127.0.0.1", "::1":
		return false
	}
	return true
}
