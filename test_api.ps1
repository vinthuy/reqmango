$ErrorActionPreference = "Continue"
$loginResp = Invoke-RestMethod -Uri "http://localhost:8000/api/v1/auth/login" -Method POST -ContentType "application/json" -Body '{"email":"admin@reqmango.com","password":"demo1234"}'
$h = @{ Authorization = "Bearer $($loginResp.access_token)" }
$baseUrl = "http://localhost:8000/api/v1"
$PID_VAR = 1
$WS_ID = 1

$passCount = 0
$failCount = 0
$results = @()

function Test-Api {
    param([string]$Name, [string]$Method, [string]$Url, [string]$Body = $null)
    Write-Host "`n--- $Name ---"
    try {
        $params = @{ Uri = $Url; Method = $Method; Headers = $h; UseBasicParsing = $true; TimeoutSec = 30 }
        if ($Body) { $params.Body = $Body; $params.ContentType = "application/json" }
        $r = Invoke-RestMethod @params
        $json = $r | ConvertTo-Json -Compress -Depth 5
        Write-Host "  PASS: $json"
        $script:passCount++
        $script:results += [PSCustomObject]@{Name=$Name; Status="PASS"; Detail=$json}
        return $r
    } catch {
        $status = $_.Exception.Response.StatusCode.value__
        Write-Host "  FAIL ($status): $($_.Exception.Message)"
        $script:failCount++
        $script:results += [PSCustomObject]@{Name=$Name; Status="FAIL ($status)"; Detail=$_.Exception.Message}
        return $null
    }
}

Write-Host "============================================"
Write-Host "  Agent-Project Integration API Tests"
Write-Host "  Project ID: $PID_VAR  Workspace ID: $WS_ID"
Write-Host "============================================"

# ============================================================
# 1. Agent Members CRUD
# ============================================================
Test-Api "List Agent Members (initial)" GET "$baseUrl/projects/$PID_VAR/agent-members"
$m1 = Test-Api "Add Agent Member (agent=1, role=admin)" POST "$baseUrl/projects/$PID_VAR/agent-members" '{"agent_id":1,"role":"admin"}'
$m3 = Test-Api "Add Agent Member (agent=3, role=member)" POST "$baseUrl/projects/$PID_VAR/agent-members" '{"agent_id":3,"role":"member"}'
Test-Api "List Agent Members (2 expected)" GET "$baseUrl/projects/$PID_VAR/agent-members"

# Use member IDs returned from POST (NOT agent_ids)
$memId1 = $null; $memId3 = $null
if ($m1) { $memId1 = $m1.id }
if ($m3) { $memId3 = $m3.id }
Write-Host "  >> member IDs: m1=$memId1 m3=$memId3"

if ($memId1) {
    Test-Api "Update Role (member=$memId1 -> member)" PUT "$baseUrl/projects/$PID_VAR/agent-members/$memId1/role" '{"role":"member"}'
}
if ($memId3) {
    Test-Api "Remove Agent Member (member=$memId3)" DELETE "$baseUrl/projects/$PID_VAR/agent-members/$memId3"
}
Test-Api "List Agent Members (1 expected)" GET "$baseUrl/projects/$PID_VAR/agent-members"

# ============================================================
# 2. Workflows CRUD
# ============================================================
Test-Api "List Workflows (initial)" GET "$baseUrl/projects/$PID_VAR/workflows"
$wf1 = Test-Api "Create Workflow 1" POST "$baseUrl/projects/$PID_VAR/workflows" '{"name":"Test Workflow","description":"Integration test"}'
$wf2 = Test-Api "Create Workflow 2" POST "$baseUrl/projects/$PID_VAR/workflows" '{"name":"Multi-Agent Pipeline","description":"Agent pipeline test"}'
Test-Api "List Workflows (2 expected)" GET "$baseUrl/projects/$PID_VAR/workflows"

$wfId = $null
if ($wf1) { $wfId = $wf1.id }
if (-not $wfId) { $wfId = 1 }
Write-Host "  >> using workflowId=$wfId"

Test-Api "Get Workflow Detail" GET "$baseUrl/projects/$PID_VAR/workflows/$wfId"
Test-Api "Update Workflow" PUT "$baseUrl/projects/$PID_VAR/workflows/$wfId" '{"name":"Updated Workflow","description":"Updated desc"}'

# ============================================================
# 3. Workflow Nodes
# ============================================================
$n1 = Test-Api "Add Node 1 (agent)" POST "$baseUrl/projects/$PID_VAR/workflows/$wfId/nodes" '{"agent_id":1,"node_type":"agent","name":"Analyzer","config":{"task_description":"Analyze requirements"}}'
$n2 = Test-Api "Add Node 2 (agent)" POST "$baseUrl/projects/$PID_VAR/workflows/$wfId/nodes" '{"agent_id":3,"node_type":"agent","name":"Developer","config":{"task_description":"Implement feature"}}'
$n3 = Test-Api "Add Node 3 (condition)" POST "$baseUrl/projects/$PID_VAR/workflows/$wfId/nodes" '{"agent_id":1,"node_type":"condition","name":"Quality Check","config":{"condition_expression":"score > 0.8"}}'
$n4 = Test-Api "Add Node 4 (notification)" POST "$baseUrl/projects/$PID_VAR/workflows/$wfId/nodes" '{"agent_id":1,"node_type":"notification","name":"Notify Team","config":{"channel":"email","recipients":"team@reqmango.com"}}'

$nodeId1 = $null; $nodeId2 = $null; $nodeId3 = $null; $nodeId4 = $null
if ($n1) { $nodeId1 = $n1.id }
if ($n2) { $nodeId2 = $n2.id }
if ($n3) { $nodeId3 = $n3.id }
if ($n4) { $nodeId4 = $n4.id }
Write-Host "  >> node IDs: n1=$nodeId1 n2=$nodeId2 n3=$nodeId3 n4=$nodeId4"

# ============================================================
# 4. Workflow Edges (use returned node IDs)
# ============================================================
if ($nodeId1 -and $nodeId2) {
    Test-Api "Add Edge 1->2" POST "$baseUrl/projects/$PID_VAR/workflows/$wfId/edges" "{`"source_node_id`":$nodeId1,`"target_node_id`":$nodeId2}"
}
if ($nodeId2 -and $nodeId3) {
    Test-Api "Add Edge 2->3" POST "$baseUrl/projects/$PID_VAR/workflows/$wfId/edges" "{`"source_node_id`":$nodeId2,`"target_node_id`":$nodeId3,`"condition`":`"always`"}"
}
if ($nodeId3 -and $nodeId4) {
    Test-Api "Add Edge 3->4" POST "$baseUrl/projects/$PID_VAR/workflows/$wfId/edges" "{`"source_node_id`":$nodeId3,`"target_node_id`":$nodeId4}"
}

Test-Api "Get Workflow (with nodes/edges)" GET "$baseUrl/projects/$PID_VAR/workflows/$wfId"

# ============================================================
# 5. Budget & SLA
# ============================================================
Test-Api "Get Budget" GET "$baseUrl/projects/$PID_VAR/budget"
Test-Api "Update Budget" PUT "$baseUrl/projects/$PID_VAR/budget" '{"monthly_budget":500,"alert_threshold":0.8,"auto_block":true}'
Test-Api "Get Budget (after update)" GET "$baseUrl/projects/$PID_VAR/budget"

Test-Api "Get SLA" GET "$baseUrl/projects/$PID_VAR/sla"
Test-Api "Update SLA" PUT "$baseUrl/projects/$PID_VAR/sla" '{"normal_task_max":3600,"complex_task_max":14400,"auto_escalation":true}'
Test-Api "Get SLA (after update)" GET "$baseUrl/projects/$PID_VAR/sla"

# ============================================================
# 6. Issue-Agent Assignment (create issue with correct format)
# ============================================================
Write-Host "`n--- Creating test issue ---"
# Create issue: name (not title) in body; project_id + workspace_id as query params
$issueBody = @{name="Test Issue for Agent Assignment";description_html="<p>Automated test</p>";priority="medium";state_id=2} | ConvertTo-Json -Compress
$issue = $null
try {
    $issue = Invoke-RestMethod -Uri "$baseUrl/issues?project_id=$PID_VAR&workspace_id=$WS_ID" -Method POST -ContentType "application/json" -Body $issueBody -Headers $h -UseBasicParsing
    Write-Host "  PASS: Created issue ID: $($issue.id)"
    $script:passCount++
    $script:results += [PSCustomObject]@{Name="Create Issue"; Status="PASS"; Detail="issue id=$($issue.id)"}
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    Write-Host "  FAIL ($status): $($_.Exception.Message)"
    $script:failCount++
    $script:results += [PSCustomObject]@{Name="Create Issue"; Status="FAIL ($status)"; Detail=$_.Exception.Message}
}

if ($issue) {
    $ISS_ID = $issue.id
    Test-Api "Assign Agent to Issue" POST "$baseUrl/issues/$ISS_ID/assign-agent" '{"agent_id":1,"priority":"high","notes":"Urgent task"}'
    Test-Api "Get Agent Status" GET "$baseUrl/issues/$ISS_ID/agent-status"
    Test-Api "Preview Execution" POST "$baseUrl/issues/$ISS_ID/preview-execution?agent_id=1" '{}'
    Test-Api "Get Decision Records" GET "$baseUrl/issues/$ISS_ID/decisions"
    Test-Api "Unassign Agent" DELETE "$baseUrl/issues/$ISS_ID/assign-agent"
}

# ============================================================
# 7. Workflow Execute & Runs (expected to need LLM)
# ============================================================
Test-Api "Execute Workflow (may fail without LLM)" POST "$baseUrl/projects/$PID_VAR/workflows/$wfId/execute" '{}'
Test-Api "List Workflow Runs" GET "$baseUrl/projects/$PID_VAR/workflows/$wfId/runs"

# ============================================================
# Summary
# ============================================================
Write-Host "`n============================================"
Write-Host "  Tests Complete!"
Write-Host "  PASS: $passCount"
Write-Host "  FAIL: $failCount"
Write-Host "============================================"
