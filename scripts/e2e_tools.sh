#!/usr/bin/env bash
# reqmango tools e2e smoke test.
# Prerequisites: backend running (make dev-backend), database migrated.
# Usage: bash scripts/e2e_tools.sh [api_base_url]
set -euo pipefail

API="${1:-http://localhost:8000/api/v1}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/bin"
mkdir -p "$BIN"

# 必须与 sdk/mcp 依赖的 mcp-go 的 mcp.LATEST_PROTOCOL_VERSION 一致；升级 mcp-go 时同步更新
MCP_PROTOCOL_VERSION="2026-07-28"

echo "==> building tools binaries"
(cd "$ROOT/sdk" && go build -o "$BIN/reqmango" ./cmd/reqmango && go build -o "$BIN/reqmango-mcp" ./cmd/reqmango-mcp)

EMAIL="e2e-$(date +%s)@example.com"
PASSWORD="e2e-pass-123"

echo "==> registering user $EMAIL"
curl -sf -X POST "$API/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"username\":\"e2e$(date +%s)\",\"password\":\"$PASSWORD\"}" >/dev/null

echo "==> reqmango auth login"
CONFIG="$(mktemp -d)/config.json"
"$BIN/reqmango" --config "$CONFIG" auth login --api-url "$API" \
  --email "$EMAIL" --password "$PASSWORD" >/dev/null
PAT="$(sed -n 's/.*"pat": "\([^"]*\)".*/\1/p' "$CONFIG")"
[ -n "$PAT" ] || { echo "FAIL: no PAT in config"; exit 1; }

export REQMANGO_API_URL="$API"
export REQMANGO_PAT="$PAT"

echo "==> workspace discovery + switch"
WS_ID="$("$BIN/reqmango" --config "$CONFIG" workspace list --output json | grep -o '"id": [0-9]*' | head -1 | grep -o '[0-9]*')"
[ -n "$WS_ID" ] || { echo "FAIL: no workspace found (create one in the UI first)"; exit 1; }
"$BIN/reqmango" --config "$CONFIG" workspace switch "$WS_ID" >/dev/null

echo "==> project list"
PROJ_ID="$("$BIN/reqmango" --config "$CONFIG" project list --output json | grep -o '"id": [0-9]*' | head -1 | grep -o '[0-9]*')"
[ -n "$PROJ_ID" ] || { echo "FAIL: no project found (create one in the UI first)"; exit 1; }
IDENT="$("$BIN/reqmango" --config "$CONFIG" project list --output json | grep -o '"identifier": "[^"]*"' | head -1 | sed 's/.*"identifier": "\([^"]*\)"/\1/')"

echo "==> issue lifecycle (create/show by code/update)"
CREATE_OUT="$("$BIN/reqmango" --config "$CONFIG" issue create --project "$PROJ_ID" --title "e2e smoke $(date +%s)" --priority medium --output json)"
ISSUE_ID="$(echo "$CREATE_OUT" | grep -o '"id": [0-9]*' | head -1 | grep -o '[0-9]*')"
SEQ="$(echo "$CREATE_OUT" | grep -o '"sequence_id": [0-9]*' | head -1 | grep -o '[0-9]*')"
[ -n "$ISSUE_ID" ] || { echo "FAIL: issue create"; exit 1; }
"$BIN/reqmango" --config "$CONFIG" issue show "$IDENT-$SEQ" >/dev/null
"$BIN/reqmango" --config "$CONFIG" issue update "$ISSUE_ID" --priority high >/dev/null
"$BIN/reqmango" --config "$CONFIG" issue list --project "$PROJ_ID" --limit 5 >/dev/null
echo "   issue $IDENT-$SEQ (id $ISSUE_ID) OK"

echo "==> mcp stdio handshake (initialize + initialized + tools/list)"
STDIO_OUT=$(printf '%s\n%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"'"$MCP_PROTOCOL_VERSION"'","clientInfo":{"name":"e2e","version":"1.0.0"},"capabilities":{}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | "$BIN/reqmango-mcp")
echo "$STDIO_OUT" | grep -q '"tools"' || { echo "FAIL: stdio tools/list"; echo "$STDIO_OUT"; exit 1; }
# 先按含 "tools" 的行过滤出 tools/list 响应行，再统计该行中 "name":"..." 工具名字段个数
TOOL_COUNT="$(echo "$STDIO_OUT" | grep '"tools"' | grep -o '"name":"[a-z_]*"' | wc -l)"
[ "$TOOL_COUNT" -ge 24 ] || { echo "FAIL: expected >=24 tools over stdio, got $TOOL_COUNT"; exit 1; }
echo "   stdio OK ($TOOL_COUNT tools)"

echo "==> mcp streamable HTTP smoke"
"$BIN/reqmango-mcp" --http :18080 >/dev/null 2>&1 &
MCP_PID=$!
cleanup() {
  kill "$MCP_PID" 2>/dev/null || true
  # Windows/Git Bash: plain kill often does not terminate the process; taskkill as fallback
  # (doubled slashes required so Git Bash does not path-mangle them; guarded so POSIX is unaffected)
  command -v taskkill >/dev/null 2>&1 && taskkill //F //PID "$MCP_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT
sleep 1
# HTTP 模式需要 Authorization: Bearer <PAT>（spec §5.1）
AUTH_HEADER="Authorization: Bearer $PAT"
# 先验证无凭据被拒（必须恰好 401）
NOAUTH_CODE="$(curl -s -o /dev/null -w '%{http_code}' -X POST "http://localhost:18080/mcp" \
  -H 'Content-Type: application/json')"
[ "$NOAUTH_CODE" = "401" ] || { echo "FAIL: expected 401 without Bearer token, got $NOAUTH_CODE"; exit 1; }
SID=$(curl -sf -X POST "http://localhost:18080/mcp" -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' -H "$AUTH_HEADER" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"'"$MCP_PROTOCOL_VERSION"'","clientInfo":{"name":"e2e","version":"1.0.0"},"capabilities":{}}}' \
  -D - -o /dev/null | tr -d '\r' | sed -n 's/[Mm]cp-Session-Id: //p')
[ -n "$SID" ] || { echo "FAIL: no session id from HTTP initialize"; exit 1; }
curl -sf -X POST "http://localhost:18080/mcp" -H 'Content-Type: application/json' \
  -H "Mcp-Session-Id: $SID" -H "$AUTH_HEADER" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_workspaces","arguments":{}}}' \
  | grep -q '"result"' || { echo "FAIL: list_workspaces over HTTP"; exit 1; }
echo "   HTTP OK"

echo "ALL E2E SMOKE TESTS PASSED"
