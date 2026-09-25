# T20b-T24: Label/Cycle/Archive/Search/Tree Tests (fixed paths)
$BASE = 'http://localhost:8000/api/v1'
if (-not $BASE) { $BASE = if ($env:REQMANGO_API_URL) { $env:REQMANGO_API_URL.TrimEnd('/') } else { 'http://localhost:8000/api/v1' } }
$email = if ($env:REQMANGO_EMAIL) { $env:REQMANGO_EMAIL } else { 'demo@example.com' }
$password = if ($env:REQMANGO_PASSWORD) { $env:REQMANGO_PASSWORD } else { 'demo1234' }
$login = Invoke-RestMethod -Uri "$BASE/auth/login" -Method POST -ContentType 'application/json' -Body (@{ email = $email; password = $password } | ConvertTo-Json)
$TOKEN = $login.access_token
if (-not $TOKEN) { throw 'Login failed: no access_token (set REQMANGO_EMAIL / REQMANGO_PASSWORD or start backend with seed data)' }
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

Write-Host "===== T20b: Add Label to Issue $ISSUE ====="
try {
    $labels = Invoke-RestMethod -Uri "$BASE/projects/$PID_/settings/labels" -Headers $H -Method Get
    $labelId = $labels[0].id
    Write-Host "  Using label_id=$labelId ($($labels[0].name))"
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$ISSUE/labels?label_id=$labelId" -Headers $H -Method Post
    TestResult "T20b AddLabel" ($resp -ne $null) "label added"
} catch {
    TestResult "T20b AddLabel" $false $_.Exception.Message
}

Write-Host "===== T20c: Remove Label from Issue $ISSUE ====="
try {
    $labels = Invoke-RestMethod -Uri "$BASE/projects/$PID_/settings/labels" -Headers $H -Method Get
    $labelId = $labels[0].id
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$ISSUE/labels/$labelId" -Headers $H -Method Delete
    TestResult "T20c RemoveLabel" $true "removed"
} catch {
    TestResult "T20c RemoveLabel" $false $_.Exception.Message
}

Write-Host "===== T21a: List Cycles in Project $PID_ ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/projects/$PID_/cycles" -Headers $H -Method Get
    if ($resp -is [array]) { $cnt = $resp.Count } elseif ($resp.cycles) { $cnt = $resp.cycles.Count } else { $cnt = 0 }
    TestResult "T21a ListCycles" ($cnt -ge 0) "count=$cnt"
} catch {
    TestResult "T21a ListCycles" $false $_.Exception.Message
}

Write-Host "===== T22a: Archive Issue $SUB ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$SUB/archive" -Headers $H -Method Post
    TestResult "T22a Archive" $true "archived"
} catch {
    TestResult "T22a Archive" $false $_.Exception.Message
}

Write-Host "===== T22b: Verify archived issue ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$SUB" -Headers $H -Method Get
    $isArchived = $resp.archived_at -ne $null -or $resp.is_archived -eq $true
    TestResult "T22b VerifyArchived" $isArchived "archived_at=$($resp.archived_at)"
} catch {
    TestResult "T22b VerifyArchived" $false $_.Exception.Message
}

Write-Host "===== T22c: Restore Issue $SUB (POST archive with restore param) ====="
try {
    $body = @{ restore = $true } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$SUB/archive" -Headers $H -Method Post -Body $body
    TestResult "T22c Restore" $true "restored"
} catch {
    TestResult "T22c Restore" $false $_.Exception.Message
}

Write-Host "===== T23a: Search Issues ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/search?workspace_id=$WID&query=Epic" -Headers $H -Method Get
    $cnt = if ($resp -is [array]) { $resp.Count } elseif ($resp.results) { $resp.results.Count } else { 0 }
    TestResult "T23a Search" ($cnt -ge 0) "results=$cnt"
} catch {
    TestResult "T23a Search" $false $_.Exception.Message
}

Write-Host "===== T23b: Suggest Issues ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/suggest?project_id=$PID_&query=" -Headers $H -Method Get
    $cnt = if ($resp -is [array]) { $resp.Count } elseif ($resp.results) { $resp.results.Count } else { 0 }
    TestResult "T23b Suggest" ($cnt -ge 0) "candidates=$cnt"
} catch {
    TestResult "T23b Suggest" $false $_.Exception.Message
}

Write-Host "===== T24a: List Issues Tree ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/tree?project_id=$PID_" -Headers $H -Method Get
    $cnt = if ($resp -is [array]) { $resp.Count } elseif ($resp.tree) { $resp.tree.Count } else { 0 }
    TestResult "T24a TreeList" ($cnt -ge 0) "root_nodes=$cnt"
} catch {
    TestResult "T24a TreeList" $false $_.Exception.Message
}

Write-Host "===== T24b: Get Issue $ISSUE children ====="
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$ISSUE/children" -Headers $H -Method Get
    $cnt = if ($resp -is [array]) { $resp.Count } elseif ($resp.children) { $resp.children.Count } else { 0 }
    TestResult "T24b Children" ($cnt -ge 1) "children=$cnt"
} catch {
    TestResult "T24b Children" $false $_.Exception.Message
}

Write-Host ""
Write-Host "===== SUMMARY T20b-T24 ====="
Write-Host "PASS: $pass  FAIL: $fail"
$results | ForEach-Object { Write-Host "  $_" }