#!/bin/bash
BASE="http://localhost:8000/api/v1"
PASS=0
FAIL=0

check() { if [ "$2" = "$3" ]; then PASS=$((PASS+1)); echo "  PASS: $1"; else FAIL=$((FAIL+1)); echo "  FAIL: $1 (got '$2', expected '$3')"; fi; }
check_contains() { if echo "$2" | grep -q "$3"; then PASS=$((PASS+1)); echo "  PASS: $1"; else FAIL=$((FAIL+1)); echo "  FAIL: $1 (response missing '$3')"; echo "    Response: ${2:0:200}"; fi; }

echo "=== API Integration Tests ==="

# 1. Register (may already exist, that's ok)
echo "1. Register"
R=$(curl -s -X POST $BASE/auth/register -H "Content-Type: application/json" -d '{"email":"inttest2@test.com","username":"inttest2","password":"test1234","display_name":"Integration Test 2"}')
if echo "$R" | grep -q '"email"'; then
  check_contains "Register" "$R" '"email"'
elif echo "$R" | grep -q "already registered"; then
  PASS=$((PASS+1)); echo "  PASS: Register (already exists, ok)"
else
  FAIL=$((FAIL+1)); echo "  FAIL: Register unexpected response: ${R:0:100}"
fi

# 2. Login
echo "2. Login"
R=$(curl -s -X POST $BASE/auth/login -H "Content-Type: application/json" -d '{"email":"inttest2@test.com","password":"test1234"}')
TOKEN=$(echo "$R" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
check_contains "Login" "$R" "access_token"

if [ -z "$TOKEN" ]; then echo "FATAL: No token"; exit 1; fi
AUTH="Authorization: Bearer $TOKEN"

# 3. Create Workspace
echo "3. Create Workspace"
R=$(curl -s -X POST $BASE/workspaces -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"IntTest WS 2","slug":"inttest-ws2","organization_size":"5","timezone":"UTC"}')
WS_ID=$(echo "$R" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
check_contains "Create Workspace" "$R" '"slug":"inttest-ws2"'

# 4. Create Project
echo "4. Create Project"
R=$(curl -s -X POST "$BASE/projects?workspace_id=$WS_ID" -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"IntTest Project","identifier":"INT","description":"Test","is_public":true}')
PROJ_ID=$(echo "$R" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
check_contains "Create Project" "$R" '"identifier":"INT"'

# 5. List States (project-level)
echo "5. List Project States"
R=$(curl -s "$BASE/projects/$PROJ_ID/settings/states" -H "$AUTH")
check_contains "List States" "$R" '"group"'

# 6. Create State
echo "6. Create State"
R=$(curl -s -X POST "$BASE/projects/$PROJ_ID/settings/states?workspace_id=$WS_ID" -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"Testing","color":"#F59E0B","group":"started"}')
STATE_ID=$(echo "$R" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
check_contains "Create State" "$R" '"name":"Testing"'

# 7. Update State
echo "7. Update State"
R=$(curl -s -X PUT "$BASE/projects/$PROJ_ID/settings/states/$STATE_ID" -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"QA Testing","color":"#10B981"}')
check_contains "Update State" "$R" '"name":"QA Testing"'

# 8. Create Label (project-level)
echo "8. Create Label"
R=$(curl -s -X POST "$BASE/projects/$PROJ_ID/settings/labels?workspace_id=$WS_ID" -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"frontend","color":"#3B82F6"}')
LBL_ID=$(echo "$R" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
check_contains "Create Label" "$R" '"name":"frontend"'

# 9. List Labels
echo "9. List Labels"
R=$(curl -s "$BASE/projects/$PROJ_ID/settings/labels" -H "$AUTH")
check_contains "List Labels" "$R" '"frontend"'

# 10. Create Workflow (project-level)
echo "10. Create Workflow"
R=$(curl -s -X POST "$BASE/projects/$PROJ_ID/workflows" -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"Standard","description":"Default workflow"}')
WF_ID=$(echo "$R" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
check_contains "Create Workflow" "$R" '"name":"Standard"'

# 11. List Workflows
echo "11. List Workflows"
R=$(curl -s "$BASE/projects/$PROJ_ID/workflows" -H "$AUTH")
check_contains "List Workflows" "$R" '"Standard"'

# 12. Create Automation (project-level)
echo "12. Create Automation"
R=$(curl -s -X POST "$BASE/projects/$PROJ_ID/automations" -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"Auto Assign","trigger_type":"issue_created","conditions":"[]","actions":"[]"}')
AUTO_ID=$(echo "$R" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
check_contains "Create Automation" "$R" '"name":"Auto Assign"'

# 13. List Automations
echo "13. List Automations"
R=$(curl -s "$BASE/projects/$PROJ_ID/automations" -H "$AUTH")
check_contains "List Automations" "$R" '"Auto Assign"'

# 14. Issue Types (workspace-level)
echo "14. Create Issue Type"
R=$(curl -s -X POST "$BASE/issue-types?workspace_id=$WS_ID" -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"Bug","color":"#EF4444","icon":"bug","description":"Software bug"}')
IT_ID=$(echo "$R" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
check_contains "Create Issue Type" "$R" '"name":"Bug"'

# 15. Custom Fields (workspace-level)
echo "15. Create Custom Field"
R=$(curl -s -X POST "$BASE/custom-fields?workspace_id=$WS_ID" -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"Severity","description":"Bug severity","field_type":"dropdown","is_required":false}')
CF_ID=$(echo "$R" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
check_contains "Create Custom Field" "$R" '"name":"Severity"'

# 16. Bind Field to Issue Type
echo "16. Bind Field to Type"
R=$(curl -s -X POST "$BASE/issue-types/$IT_ID/fields" -H "Content-Type: application/json" -H "$AUTH" -d "{\"field_id\":$CF_ID,\"is_required\":true,\"sequence\":1}")
check_contains "Bind Field" "$R" '"is_required"'

# 17. List Type Fields
echo "17. List Type Fields"
R=$(curl -s "$BASE/issue-types/$IT_ID/fields" -H "$AUTH")
check_contains "List Fields" "$R" '"Severity"'

# 18. Relations (workspace-level)
echo "18. Create Relation Type"
R=$(curl -s -X POST "$BASE/relations/types?workspace_id=$WS_ID" -H "Content-Type: application/json" -H "$AUTH" -d '{"name":"Blocks","inward_name":"blocked by","outward_name":"blocks"}')
check_contains "Create Relation" "$R" '"name":"Blocks"'

# Cleanup
echo "19. Delete State"
R=$(curl -s -X DELETE "$BASE/projects/$PROJ_ID/settings/states/$STATE_ID" -H "$AUTH")
check "Delete State (204)" "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE/projects/$PROJ_ID/settings/states/$STATE_ID" -H "$AUTH")" "204"

echo "20. Delete Label"
check "Delete Label (204)" "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE/projects/$PROJ_ID/settings/labels/$LBL_ID" -H "$AUTH")" "204"

echo "21. Unbind Field"
R=$(curl -s -X DELETE "$BASE/issue-types/$IT_ID/fields/$CF_ID" -H "$AUTH")
check_contains "Unbind Field" "$R" "removed"

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
