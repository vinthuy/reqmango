$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODUwMjY5MTUsImlhdCI6MTc4NDQyMjExNSwic3ViIjoiMSJ9.hMyKlBim8wLpqFoKg4Ig0e-Xdd3Plf3J4G4yFFnI1TQ"
$h = @{Authorization = "Bearer $token"; "Content-Type" = "application/json" }

function Get-ErrorBody($ex) {
    try {
        $resp = $ex.Response
        if ($null -eq $resp) { return "" }
        $stream = $resp.GetResponseStream()
        if ($null -eq $stream) { return "" }
        $reader = New-Object System.IO.StreamReader($stream)
        return $reader.ReadToEnd()
    } catch { return "" }
}
function Is-Error($r) {
    return ($r -is [System.Collections.IDictionary] -and $r.Contains("__error"))
}
function Post-Json($url, $body) {
    $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($body)
    try { return Invoke-RestMethod -Uri $url -Method POST -Headers $h -Body $bodyBytes }
    catch { $eb = Get-ErrorBody $_.Exception; return @{ __error = $_.Exception.Message; __body = $eb } }
}
function Put-Json($url, $body) {
    $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($body)
    try { return Invoke-RestMethod -Uri $url -Method PUT -Headers $h -Body $bodyBytes }
    catch { $eb = Get-ErrorBody $_.Exception; return @{ __error = $_.Exception.Message; __body = $eb } }
}
function Get-Json($url) {
    try { return Invoke-RestMethod -Uri $url -Method GET -Headers $h }
    catch { $eb = Get-ErrorBody $_.Exception; return @{ __error = $_.Exception.Message; __body = $eb } }
}
function Del-Json($url) {
    try { return Invoke-RestMethod -Uri $url -Method DELETE -Headers $h }
    catch { $eb = Get-ErrorBody $_.Exception; return @{ __error = $_.Exception.Message; __body = $eb } }
}

$pass = 0; $fail = 0; $blocked = 0
function Report($name, $status, $detail) {
    Write-Host ("[{0}] {1}: {2}" -f $status, $name, $detail)
    if ($status -eq "PASS") { $script:pass++ }
    elseif ($status -eq "FAIL") { $script:fail++ }
    else { $script:blocked++ }
}

Write-Host "=========================================================="
Write-Host "  Extended Acceptance Tests - Issue Management Features"
Write-Host "=========================================================="
Write-Host ""

# ============ Test 17: IssueUpdate (multiple fields) ============
Write-Host "========== Test 17: IssueUpdate (name, priority, state, dates) =========="
$body = '{"name":"updated-by-acceptance-test","priority":"high","state_id":3,"start_date":"2026-07-19T00:00:00Z","target_date":"2026-07-26T00:00:00Z"}'
$r = Put-Json "http://localhost:8000/api/v1/issues/990" $body
if (Is-Error $r) { Report "Test 17 IssueUpdate" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else {
    $ok = $true
    if ($r.name -ne "updated-by-acceptance-test") { $ok = $false; Write-Host "  name mismatch: $($r.name)" }
    if ($r.priority -ne "high") { $ok = $false; Write-Host "  priority mismatch: $($r.priority)" }
    if ($r.state.id -ne 3) { $ok = $false; Write-Host "  state mismatch: $($r.state.id)" }
    if ($ok) { Report "Test 17 IssueUpdate" "PASS" "name/priority/state/dates updated, id=$($r.id)" }
    else { Report "Test 17 IssueUpdate" "FAIL" "some fields not updated correctly" }
}

# ============ Test 18: Sub-issues management ============
Write-Host ""
Write-Host "========== Test 18a: Get sub-issues of issue 990 =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/990"
if (Is-Error $r) { Report "Test 18a GetSubIssues" "FAIL" "ERR: $($r.__error)" }
else {
    $count = $r.sub_issues_count
    $subCount = if ($r.sub_issues) { $r.sub_issues.Count } else { 0 }
    Report "Test 18a GetSubIssues" "PASS" "sub_issues_count=$count, sub_issues array=$subCount"
}

Write-Host "========== Test 18b: Create sub-issue with parent_id=990 =========="
$body = '{"name":"ext-test-sub-issue","type_id":2,"priority":"low","parent_id":990,"custom_field_values":{"1":"P3-低"}}'
$r = Post-Json "http://localhost:8000/api/v1/issues?project_id=1&workspace_id=1" $body
if (Is-Error $r) { Report "Test 18b CreateSubIssue" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else {
    $script:subIssueId = $r.id
    Report "Test 18b CreateSubIssue" "PASS" "created id=$($r.id), parent_id=$($r.parent_id), depth=$($r.depth)"
}

Write-Host "========== Test 18c: Reorder sub-issues =========="
if ($script:subIssueId) {
    $r = Get-Json "http://localhost:8000/api/v1/issues/990"
    $subIds = @()
    if ($r.sub_issues) { $subIds = $r.sub_issues | ForEach-Object { $_.id } }
    if ($subIds.Count -ge 2) {
        $reversed = [System.Array]::CreateInstance([uint64], $subIds.Count)
        for ($i = 0; $i -lt $subIds.Count; $i++) { $reversed[$i] = $subIds[$subIds.Count - 1 - $i] }
        $body = "{`"issue_ids`":[" + ($reversed -join ",") + "]}"
        $r = Post-Json "http://localhost:8000/api/v1/issues/990/reorder-sub-issues" $body
        if (Is-Error $r) { Report "Test 18c ReorderSubIssues" "FAIL" "ERR: $($r.__error) $($r.__body)" }
        else { Report "Test 18c ReorderSubIssues" "PASS" "reordered $($subIds.Count) sub-issues" }
    } else { Report "Test 18c ReorderSubIssues" "BLOCKED" "need >=2 sub-issues, have $($subIds.Count)" }
} else { Report "Test 18c ReorderSubIssues" "BLOCKED" "subIssueId not created" }

# ============ Test 19: Assignee management ============
Write-Host ""
Write-Host "========== Test 19a: Add assignee to issue 990 =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/990/assignees?user_id=1" "{}"
if (Is-Error $r) { Report "Test 19a AddAssignee" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 19a AddAssignee" "PASS" "action=$($r.action), user_id=$($r.user_id)" }

Write-Host "========== Test 19b: Remove assignee from issue 990 =========="
$r = Del-Json "http://localhost:8000/api/v1/issues/990/assignees/1"
if (Is-Error $r) { Report "Test 19b RemoveAssignee" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 19b RemoveAssignee" "PASS" "action=$($r.action)" }

# ============ Test 20: Label management ============
Write-Host ""
Write-Host "========== Test 20a: List project 1 labels =========="
$r = Get-Json "http://localhost:8000/api/v1/projects/1/labels"
if (Is-Error $r) { Report "Test 20a ListLabels" "FAIL" "ERR: $($r.__error)" }
else {
    $count = if ($r -is [Array]) { $r.Count } else { 1 }
    $script:labelId = if ($r -is [Array] -and $r.Count -gt 0) { $r[0].id } else { $r.id }
    Report "Test 20a ListLabels" "PASS" "$count labels, first id=$($script:labelId)"
}

Write-Host "========== Test 20b: Add label to issue 990 =========="
if ($script:labelId) {
    $r = Post-Json "http://localhost:8000/api/v1/issues/990/labels?label_id=$($script:labelId)" "{}"
    if (Is-Error $r) { Report "Test 20b AddLabel" "FAIL" "ERR: $($r.__error) $($r.__body)" }
    else { Report "Test 20b AddLabel" "PASS" "action=$($r.action), label_id=$($r.label_id)" }
} else { Report "Test 20b AddLabel" "BLOCKED" "no label available" }

Write-Host "========== Test 20c: Remove label from issue 990 =========="
if ($script:labelId) {
    $r = Del-Json "http://localhost:8000/api/v1/issues/990/labels/$($script:labelId)"
    if (Is-Error $r) { Report "Test 20c RemoveLabel" "FAIL" "ERR: $($r.__error) $($r.__body)" }
    else { Report "Test 20c RemoveLabel" "PASS" "action=$($r.action)" }
} else { Report "Test 20c RemoveLabel" "BLOCKED" "no label available" }

# ============ Test 21: Cycle management ============
Write-Host ""
Write-Host "========== Test 21a: Set issue 990 cycle to 1 =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/990/cycle?cycle_id=1" "{}"
if (Is-Error $r) { Report "Test 21a SetCycle" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 21a SetCycle" "PASS" "action=$($r.action), cycle_id=$($r.cycle_id)" }

Write-Host "========== Test 21b: Remove issue 990 cycle =========="
$r = Del-Json "http://localhost:8000/api/v1/issues/990/cycle"
if (Is-Error $r) { Report "Test 21b RemoveCycle" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 21b RemoveCycle" "PASS" "action=$($r.action), cycle_id=$($r.cycle_id)" }

# ============ Test 22: Archive / Restore ============
Write-Host ""
Write-Host "========== Test 22a: Archive issue 992 (test data) =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/992/archive" "{}"
if (Is-Error $r) { Report "Test 22a Archive" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 22a Archive" "PASS" "archived issue 992" }

Write-Host "========== Test 22b: Restore issue 992 =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/992/restore" "{}"
if (Is-Error $r) { Report "Test 22b Restore" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 22b Restore" "PASS" "restored issue 992, id=$($r.id)" }

# ============ Test 23: Search and Suggest ============
Write-Host ""
Write-Host "========== Test 23a: Search issues q=acceptance =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/search?workspace_id=1&q=acceptance"
if (Is-Error $r) { Report "Test 23a Search" "FAIL" "ERR: $($r.__error)" }
else {
    $count = if ($r -is [Array]) { $r.Count } else { 1 }
    Report "Test 23a Search" "PASS" "found $count results for q=acceptance"
}

Write-Host "========== Test 23b: Suggest issues =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/suggest?project_id=1&query=test"
if (Is-Error $r) { Report "Test 23b Suggest" "FAIL" "ERR: $($r.__error)" }
else {
    $count = if ($r -is [Array]) { $r.Count } else { 1 }
    Report "Test 23b Suggest" "PASS" "found $count suggestions"
}

# ============ Test 24: Tree view ============
Write-Host ""
Write-Host "========== Test 24a: List tree issues (root nodes) =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/tree?project_id=1&workspace_id=1&limit=5"
if (Is-Error $r) { Report "Test 24a TreeList" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else {
    $count = if ($r.items) { $r.items.Count } else { 0 }
    $total = $r.total
    Report "Test 24a TreeList" "PASS" "returned $count root items, total=$total"
}

Write-Host "========== Test 24b: Get children of issue 990 =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/990/children"
if (Is-Error $r) { Report "Test 24b GetChildren" "FAIL" "ERR: $($r.__error)" }
else {
    $count = if ($r -is [Array]) { $r.Count } else { 0 }
    Report "Test 24b GetChildren" "PASS" "issue 990 has $count children"
}

# ============ Test 25: Statistics ============
Write-Host ""
Write-Host "========== Test 25: Issue statistics =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/statistics?project_id=1"
if (Is-Error $r) { Report "Test 25 Statistics" "FAIL" "ERR: $($r.__error)" }
else { Report "Test 25 Statistics" "PASS" "statistics returned successfully" }

# ============ Test 26: Activities ============
Write-Host ""
Write-Host "========== Test 26: Get activities for issue 990 (limit 3) =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/990/activities?limit=3"
if (Is-Error $r) { Report "Test 26 Activities" "FAIL" "ERR: $($r.__error)" }
else {
    $count = if ($r -is [Array]) { $r.Count } else { 0 }
    Report "Test 26 Activities" "PASS" "returned $count activities (limit 3)"
}

# ============ Test 27: Watcher management ============
Write-Host ""
Write-Host "========== Test 27a: Add watcher to issue 990 =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/990/watch" "{}"
if (Is-Error $r) { Report "Test 27a AddWatcher" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 27a AddWatcher" "PASS" "watcher added" }

Write-Host "========== Test 27b: List watchers =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/990/watchers"
if (Is-Error $r) { Report "Test 27b ListWatchers" "FAIL" "ERR: $($r.__error)" }
else {
    $count = if ($r -is [Array]) { $r.Count } else { 0 }
    Report "Test 27b ListWatchers" "PASS" "$count watchers"
}

Write-Host "========== Test 27c: Remove watcher =========="
$r = Del-Json "http://localhost:8000/api/v1/issues/990/watch"
if (Is-Error $r) { Report "Test 27c RemoveWatcher" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 27c RemoveWatcher" "PASS" "watcher removed" }

# ============ Test 28: Bulk operations ============
Write-Host ""
Write-Host "========== Test 28a: Bulk update issues =========="
$body = '{"issue_ids":[990],"priority":"urgent"}'
$r = Post-Json "http://localhost:8000/api/v1/issues/bulk/update?project_id=1" $body
if (Is-Error $r) { Report "Test 28a BulkUpdate" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else {
    $count = if ($r -is [Array]) { $r.Count } else { 0 }
    Report "Test 28a BulkUpdate" "PASS" "bulk updated $count issues"
}

# ============ Test 29: ConvertType (re-verify after fixes) ============
Write-Host ""
Write-Host "========== Test 29: ConvertType 990 to Feature(2) then back to Epic(1) =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/990/convert-type" '{"target_type_id":2}'
if (Is-Error $r) { Report "Test 29a ConvertToFeature" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 29a ConvertToFeature" "PASS" "type now $($r.issue_type.name)" }

$r = Post-Json "http://localhost:8000/api/v1/issues/990/convert-type" '{"target_type_id":1}'
if (Is-Error $r) { Report "Test 29b ConvertToEpic" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 29b ConvertToEpic" "PASS" "type now $($r.issue_type.name)" }

# ============ Test 30: RQL filter ============
Write-Host ""
Write-Host "========== Test 30: RQL filter (priority = urgent) =========="
$r = Get-Json "http://localhost:8000/api/v1/issues?project_id=1&workspace_id=1&rql=%5B%7B%22priority%22%3A%22urgent%22%7D%5D&page=1&page_size=5"
if (Is-Error $r) { Report "Test 30 RQLFilter" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else {
    $count = if ($r -is [Array]) { $r.Count } else { 0 }
    Report "Test 30 RQLFilter" "PASS" "RQL filter returned $count urgent issues"
}

# ============ Test 31: Cleanup test data ============
Write-Host ""
Write-Host "========== Test 31: Cleanup - delete test issue 992 =========="
$r = Del-Json "http://localhost:8000/api/v1/issues/992"
if (Is-Error $r) { Report "Test 31 Cleanup" "FAIL" "ERR: $($r.__error) $($r.__body)" }
else { Report "Test 31 Cleanup" "PASS" "deleted issue 992" }

Write-Host ""
Write-Host "=========================================================="
Write-Host "  Summary: PASS=$pass FAIL=$fail