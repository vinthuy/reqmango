package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// -------- JSON-RPC 2.0 types --------

// RPCRequest is a JSON-RPC 2.0 request.
type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"` // null for notifications
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// RPCResponse is a JSON-RPC 2.0 response.
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// -------- MCP-specific initialize types --------

type initializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    serverCapabilities `json:"capabilities"`
	ServerInfo      MCPServerInfo      `json:"serverInfo"`
}

type serverCapabilities struct {
	Tools *toolsCapability `json:"tools,omitempty"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// Server handles JSON-RPC 2.0 messages over stdio.
type Server struct {
	config *Config
	client *Client
}

// NewServer creates a new MCP server.
func NewServer(cfg *Config) *Server {
	return &Server{
		config: cfg,
		client: NewClient(cfg.APIBaseURL, cfg.APIToken),
	}
}

// Run starts the server loop reading from stdin and writing to stdout.
func (s *Server) Run(stdin io.Reader, stdout io.Writer) error {
	// Redirect standard library log output to stderr to avoid corrupting stdout JSON-RPC.
	log.SetOutput(os.Stderr)

	scanner := bufio.NewScanner(stdin)
	// Increase buffer size for large payloads (max 10MB).
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req RPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeError(stdout, nil, -32700, "Parse error: "+err.Error())
			continue
		}

		resp := s.handleRequest(&req)
		if resp == nil {
			// This was a notification (no ID) — don't respond.
			continue
		}

		s.writeResponse(stdout, resp)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stdin scanner error: %w", err)
	}
	return nil
}

func (s *Server) handleRequest(req *RPCRequest) *RPCResponse {
	// Notifications have no ID — we don't respond to them.
	isNotification := len(req.ID) == 0 || string(req.ID) == "null"

	switch req.Method {
	case "initialize":
		result := s.handleInitialize()
		if isNotification {
			return nil
		}
		return &RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}

	case "tools/list":
		if isNotification {
			return nil
		}
		return s.handleToolsList(req.ID)

	case "tools/call":
		if isNotification {
			return nil
		}
		return s.handleToolsCall(req.ID, req.Params)

	default:
		if isNotification {
			return nil
		}
		return &RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

func (s *Server) handleInitialize() initializeResult {
	return initializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: serverCapabilities{
			Tools: &toolsCapability{ListChanged: false},
		},
		ServerInfo: MCPServerInfo{
			Name:    "reqmanpy-mcp",
			Version: "1.0.0",
		},
	}
}

func (s *Server) handleToolsList(id json.RawMessage) *RPCResponse {
	tools := ListAllTools()
	return &RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *Server) handleToolsCall(id json.RawMessage, paramsRaw json.RawMessage) *RPCResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		return &RPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &RPCError{Code: -32602, Message: "Invalid params: " + err.Error()},
		}
	}

	result, err := ExecuteTool(params.Name, params.Arguments, s.client)
	if err != nil {
		return &RPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error:   &RPCError{Code: -32603, Message: err.Error()},
		}
	}

	return &RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func (s *Server) writeResponse(w io.Writer, resp *RPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("ERROR: failed to marshal response: %v", err)
		return
	}
	fmt.Fprintf(w, "%s\n", string(data))
}

func (s *Server) writeError(w io.Writer, id json.RawMessage, code int, message string) {
	resp := &RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "%s\n", string(data))
}
