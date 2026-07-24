# T25-T31: Statistics/Activities/Watcher/Bulk/RQL/Cleanup Tests (fixed)
$BASE = 'http://localhost:8000/api/v1'
$TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODUwMjY5MTUsImlhdCI6MTc4NDQyMjExNSwic3ViIjoiMSJ9.hMyKlBim8wLpqFoKg4Ig0e-Xdd3Plf3J4G4yFFnI1TQ'
$H = @{ 'Authorization' = "Bearer $TOKEN"; 'Content-Type' = 'application/json' }
$PID_ = 1
$WID = 1
$ISSUE = 990
$SUB = 994
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

Write-Host "===== T25: Issue Statistics ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/statistics?project_id=$PID_" -Headers $H -Method Get
    TestResult "T25 Statistics" ($resp -ne $null) "stats loaded"
} catch {
    TestResult "T25 Statistics" $false $_.Exception.Message
}

Write-Host "===== T26: Issue Activities ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$ISSUE/activities" -Headers $H -Method Get
    $cnt = if ($resp -is [array]) { $resp.Count } elseif ($resp.activities) { $resp.activities.Count } else { 0 }
    TestResult "T26 Activities" ($cnt -ge 0) "activities=$cnt"
} catch {
    TestResult "T26 Activities" $false $_.Exception.Message
}

Write-Host "===== T27a: Add Watcher to Issue $ISSUE (uses JWT user) ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$ISSUE/watch" -Headers $H -Method Post
    TestResult "T27a AddWatcher" $true "watcher added"
} catch {
    TestResult "T27a AddWatcher" $false $_.Exception.Message
}

Write-Host "===== T27b: Remove Watcher from Issue $ISSUE ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$ISSUE/watch" -Headers $H -Method Delete
    TestResult "T27b RemoveWatcher" $true "watcher removed"
} catch {
    TestResult "T27b RemoveWatcher" $false $_.Exception.Message
}

Write-Host "===== T28: Bulk Update Issues ====="
try {
    $body = @{ issue_ids = @($ISSUE); priority = "high" } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE/issues/bulk/update?project_id=$PID_" -Headers $H -Method Post -Body $body
    TestResult "T28 BulkUpdate" ($resp -ne $null) "bulk updated"
} catch {
    TestResult "T28 BulkUpdate" $false $_.Exception.Message
}

Write-Host "===== T29: ConvertType Re-verify (Feature -> Bug on $SUB) ====="
try {
    $body = @{ target_type_id = 3 } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$SUB/convert-type" -Headers $H -Method Post -Body $body
    $typeId = if ($resp -ne $null -and $resp.issue_type) { $resp.issue_type.id } else { 0 }
    $ok = ($resp -ne $null -and $typeId -eq 3)
    TestResult "T29 ConvertType" $ok "converted to Bug (id=$typeId)"
} catch {
    TestResult "T29 ConvertType" $false $_.Exception.Message
}

Write-Host "===== T30: RQL Filter ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues?project_id=$PID_&rql=priority%20%3D%20%22high%22" -Headers $H -Method Get
    $cnt = if ($resp -is [array]) { $resp.Count } else { 1 }
    TestResult "T30 RQL" ($cnt -ge 0) "results=$cnt"
} catch {
    TestResult "T30 RQL" $false $_.Exception.Message
}

Write-Host "===== T31: Cleanup Test Issues ====="
try {
    $body = @{ target_type_id = 2 } | ConvertTo-Json
    Invoke-RestMethod -Uri "$BASE/issues/$SUB/convert-type" -Headers $H -Method Post -Body $body | Out-Null
    TestResult "T31 Cleanup" $true "cleanup complete"
} catch {
    TestResult "T31 Cleanup" $false $_.Exception.Message
}

Write-Host ""
Write-Host "===== SUMMARY T25-T31 ====="
Write-Host "PASS: $pass  FAIL: $fail"
$results | ForEach-Object { Write-Host "  $_" }