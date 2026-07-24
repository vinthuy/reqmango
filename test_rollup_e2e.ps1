# E2E: Rollup test - uses raw JSON to avoid PowerShell serialization bugs
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000/api/v1"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Status Rollup E2E (raw JSON)" -ForegroundColor Cyan  
Write-Host "========================================" -ForegroundColor Cyan

# Login
Write-Host "`n[1] Login..." -ForegroundColor Yellow
$b = @{email="admin@reqmango.com";password="demo1234"} | ConvertTo-Json
$token = (Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method Post -Body $b -ContentType "application/json").access_token
$h = @{Authorization="Bearer $token"; "Content-Type"="application/json"}
Write-Host "  OK" -ForegroundColor Green

# Get workspace
$pj = Invoke-RestMethod -Uri "$baseUrl/projects/1" -Headers $h
$ws = $pj.workspace_id

# Clean
Write-Host "`n[2] Clean old E2E data..." -ForegroundColor Yellow
$r = Invoke-WebRequest -Uri "$baseUrl/issues?project_id=1&limit=50" -Headers $h
foreach ($i in ($r.Content | ConvertFrom-Json)) { if ($i.name -match "E2E") { Invoke-WebRequest -Uri "$baseUrl/issues/$($i.id)" -Method Delete -Headers $h | Out-Null } }
$r = Invoke-WebRequest -Uri "$baseUrl/projects/1/automations" -Headers $h
foreach ($a in ($r.Content | ConvertFrom-Json)) { if ($a.name -match "E2E") { Invoke-WebRequest -Uri "$baseUrl/projects/1/automations/$($a.id)" -Method Delete -Headers $h | Out-Null } }
Write-Host "  Clean" -ForegroundColor Green

# Create parent
Write-Host "`n[3] Create parent..." -ForegroundColor Yellow
$parent = Invoke-RestMethod -Uri "$baseUrl/issues?project_id=1&workspace_id=$ws" -Method Post -Headers $h -Body '{"name":"(E2E) Parent","state_id":2}'
$parentID = $parent.id
Write-Host "  Parent #$parentID (Todo)" -ForegroundColor Green

# Create child A
Write-Host "`n[4] Create child A..." -ForegroundColor Yellow
$c1 = Invoke-RestMethod -Uri "$baseUrl/issues?project_id=1&workspace_id=$ws" -Method Post -Headers $h -Body "{`"name`":`"(E2E) Child A`",`"state_id`":2,`"parent_id`":$parentID}"
$c1ID = $c1.id
Write-Host "  ChildA #$c1ID (Todo)" -ForegroundColor Green

# Create child B
Write-Host "`n[5] Create child B..." -ForegroundColor Yellow
$c2 = Invoke-RestMethod -Uri "$baseUrl/issues?project_id=1&workspace_id=$ws" -Method Post -Headers $h -Body "{`"name`":`"(E2E) Child B`",`"state_id`":2,`"parent_id`":$parentID}"
$c2ID = $c2.id
Write-Host "  ChildB #$c2ID (Todo)" -ForegroundColor Green

# Create rule - use raw JSON string, NO ConvertTo-Json for the body
Write-Host "`n[6] Create rollup rule (raw JSON)..." -ForegroundColor Yellow
# MUST use here-string to preserve the JSON escaping properly
$rawRule = @'
{"name":"(E2E) Rollup Rule","description":"E2E test","trigger_type":"issue.state_changed","is_enabled":true,"actions":"[{\"type\":\"rollup_to_parent\",\"value\":{\"rules\":[{\"condition\":\"any\",\"child_state\":\"In Progress\",\"parent_state\":\"In Progress\"}]}}]"}
'@
$rule = Invoke-RestMethod -Uri "$baseUrl/projects/1/automations" -Method Post -Headers $h -Body $rawRule
$ruleID = $rule.id
Write-Host "  Rule #$ruleID" -ForegroundColor Green

# Verify actions start with [
$r = Invoke-WebRequest -Uri "$baseUrl/projects/1/automations" -Headers $h
$rules = ($r.Content | ConvertFrom-Json)
$rr = $rules | Where-Object { $_.id -eq $ruleID }
$firstChar = $rr.actions[0]
Write-Host "  Actions[0] = '$firstChar' (expect '[')" -ForegroundColor White
if ($firstChar -ne '[') {
    Write-Host "  FATAL: actions not stored as array, aborting" -ForegroundColor Red
    exit 1
}

# Verify parent initial state
Write-Host "`n[7] Parent initial: $($parent.state_id) (expect 2)" -ForegroundColor White

# Trigger: child A -> In Progress
Write-Host "`n[8] Change ChildA -> In Progress..." -ForegroundColor Yellow
Invoke-RestMethod -Uri "$baseUrl/issues/$c1ID" -Method Put -Headers $h -Body '{"state_id":3}' | Out-Null
Write-Host "  Done" -ForegroundColor Green
Start-Sleep -Seconds 3

# Check result
Write-Host "`n[9] Check parent state..." -ForegroundColor Yellow
$pFinal = Invoke-RestMethod -Uri "$baseUrl/issues/$parentID" -Headers $h
Write-Host "  Parent #$parentID state_id=$($pFinal.state_id) (expect 3=InProgress)" -ForegroundColor White

if ($pFinal.state_id -eq 3) {
    Write-Host "  *** PASS: Rollup SUCCESS! Parent auto-changed to InProgress! ***" -ForegroundColor Green
} else {
    Write-Host "  *** FAIL: Expected 3, got $($pFinal.state_id) ***" -ForegroundColor Red
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  Done. IDs: P=$parentID C1=$c1ID C2=$c2ID R=$ruleID" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
