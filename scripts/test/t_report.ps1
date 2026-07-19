# Report and Kanban Acceptance Tests
$BASE = 'http://localhost:8000/api/v1'
$TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODUwMjY5MTUsImlhdCI6MTc4NDQyMjExNSwic3ViIjoiMSJ9.hMyKlBim8wLpqFoKg4Ig0e-Xdd3Plf3J4G4yFFnI1TQ'
$H = @{ 'Authorization' = "Bearer $TOKEN"; 'Content-Type' = 'application/json' }
$PID_ = 1
$WID = 1
$pass = 0; $fail = 0
$results = @()

function TestResult($name, $condition, $detail) {
    if ($condition) {
        $script:pass++
        $script:results += "[PASS] $name"
        Write-Host "[PASS] $name"
    } else {
        $script:fail++
        $script:results += "[FAIL] $name :: $detail"
        Write-Host "[FAIL] $name :: $detail"
    }
}

Write-Host "===== R1: Project Statistics ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/projects/$PID_/statistics" -Headers $H -Method Get
    TestResult "R1 ProjectStats" ($resp -ne $null) "stats loaded"
} catch {
    TestResult "R1 ProjectStats" $false $_.Exception.Message
}

Write-Host "===== R2: Issue Statistics ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/statistics?project_id=$PID_" -Headers $H -Method Get
    TestResult "R2 IssueStats" ($resp -ne $null) "stats loaded"
} catch {
    TestResult "R2 IssueStats" $false $_.Exception.Message
}

Write-Host "===== R3: Report Generate (V1 - distribution by state) ====="
try {
    $body = @{ report_type = "distribution"; group_by = "state"; chart = "bar" } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE/projects/$PID_/reports" -Headers $H -Method Post -Body $body
    TestResult "R3 ReportV1" ($resp -ne $null -and $resp.Labels.Count -gt 0) "labels=$($resp.Labels.Count) values=$($resp.Values.Count)"
} catch {
    TestResult "R3 ReportV1" $false $_.Exception.Message
}

Write-Host "===== R4: Report Generate (V1 - distribution by module) ====="
try {
    $body = @{ report_type = "distribution"; group_by = "module"; chart = "bar" } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE/projects/$PID_/reports" -Headers $H -Method Post -Body $body
    TestResult "R4 ReportV1_Module" ($resp -ne $null) "labels=$($resp.Labels.Count)"
} catch {
    TestResult "R4 ReportV1_Module" $false $_.Exception.Message
}

Write-Host "===== R5: Report Generate (V2 - x_axis=state, y_axis=count) ====="
try {
    $body = @{ x_axis = "state"; y_axis = "count" } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE/projects/$PID_/reports/v2" -Headers $H -Method Post -Body $body
    TestResult "R5 ReportV2" ($resp -ne $null -and $resp.Labels.Count -gt 0) "labels=$($resp.Labels.Count)"
} catch {
    TestResult "R5 ReportV2" $false $_.Exception.Message
}

Write-Host "===== R6: Report Generate (V2 - x_axis=module, y_axis=count) ====="
try {
    $body = @{ x_axis = "module"; y_axis = "count" } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE/projects/$PID_/reports/v2" -Headers $H -Method Post -Body $body
    TestResult "R6 ReportV2_Module" ($resp -ne $null) "labels=$($resp.Labels.Count)"
} catch {
    TestResult "R6 ReportV2_Module" $false $_.Exception.Message
}

Write-Host "===== R7: Report Generate (V2 - created_vs_resolved) ====="
try {
    $body = @{ x_axis = "created_week"; y_axis = "created_vs_resolved" } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE/projects/$PID_/reports/v2" -Headers $H -Method Post -Body $body
    TestResult "R7 ReportV2_CreatedVsResolved" ($resp -ne $null) "labels=$($resp.Labels.Count)"
} catch {
    TestResult "R7 ReportV2_CreatedVsResolved" $false $_.Exception.Message
}

Write-Host "===== R8: Issue List for Kanban (state grouping) ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues?project_id=$PID_&limit=10" -Headers $H -Method Get
    $cnt = if ($resp.items) { $resp.items.Count } else { 0 }
    TestResult "R8 KanbanList" ($cnt -ge 0) "items=$cnt"
} catch {
    TestResult "R8 KanbanList" $false $_.Exception.Message
}

Write-Host "===== R9: Saved Reports List ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/projects/$PID_/saved-reports" -Headers $H -Method Get
    $cnt = if ($resp -is [array]) { $resp.Count } else { 0 }
    TestResult "R9 SavedReports" ($cnt -ge 0) "reports=$cnt"
} catch {
    TestResult "R9 SavedReports" $false $_.Exception.Message
}

Write-Host "===== R10: Flow Metrics ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/flow-metrics?project_id=$PID_" -Headers $H -Method Get
    TestResult "R10 FlowMetrics" ($resp -ne $null) "flow metrics loaded"
} catch {
    TestResult "R10 FlowMetrics" $false $_.Exception.Message
}

Write-Host ""
Write-Host "===== SUMMARY: Report & Kanban ====="
Write-Host "PASS: $pass  FAIL: $fail"
$results | ForEach-Object { Write-Host "  $_" }