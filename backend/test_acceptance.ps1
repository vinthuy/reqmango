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

$p1Gao = "P1-" + [char]0x9AD8
$p2Zhong = "P2-" + [char]0x4E2D
$p3Di = "P3-" + [char]0x4F4E

Write-Host "========== Test 10c-retry: SetIssueValue field_id=1 (dropdown -> P1) with UTF-8 body =========="
$body = '{"field_id":1,"value":"' + $p1Gao + '"}'
$r = Post-Json "http://localhost:8000/api/v1/custom-fields/issues/990/values" $body
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else { Write-Host "OK: $($r.field_name)='$($r.value)'" }

Write-Host "========== Test 10g-retry: ListIssueValues =========="
$r = Get-Json "http://localhost:8000/api/v1/custom-fields/issues/990/values"
if (Is-Error $r) {
    Write-Host "ERR: $($r.__error) $($r.__body)"
} else {
    $count = if ($r -is [Array]) { $r.Count } else { 1 }
    Write-Host "values_count: $count"
    $r | ForEach-Object { Write-Host ("  - field_id={0}, name={1}, value={2}" -f $_.field_id, $_.field_name, $_.value) }
}

Write-Host ""
Write-Host "========== Test 11a: Create child issue WITH parent_id=990 (Epic -> Feature) =========="
$body = '{"name":"acceptance-test-child-Feature","type_id":2,"priority":"medium","parent_id":990,"custom_field_values":{"1":"' + $p2Zhong + '"}}'
$r = Post-Json "http://localhost:8000/api/v1/issues?project_id=1&workspace_id=1" $body
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else { Write-Host "OK: id=$($r.id), parent_id=$($r.parent_id), depth=$($r.depth), type=$($r.issue_type.name)" }

Write-Host "========== Test 11b: Verify parent 990 has sub_issues =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/990"
if (Is-Error $r) { Write-Host "ERR: $($r.__error)" } else { Write-Host "Parent sub_issues_count=$($r.sub_issues_count)" }

Write-Host "========== Test 11c: Create child with INVALID hierarchy (Bug as child of Epic - level mismatch) =========="
$body = '{"name":"acceptance-test-invalid-hierarchy","type_id":3,"priority":"low","parent_id":990,"custom_field_values":{"1":"' + $p3Di + '"}}'
$r = Post-Json "http://localhost:8000/api/v1/issues?project_id=1&workspace_id=1" $body
if (Is-Error $r) { Write-Host "Expected ERR: $($r.__error) $($r.__body)" } else { Write-Host "Created (no hierarchy check): id=$($r.id), type=$($r.issue_type.name)" }

Write-Host ""
Write-Host "========== Test 12a: ConvertType 990 Epic(1) -> Feature(2) =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/990/convert-type" '{"target_type_id":2}'
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else { Write-Host "OK: type now $($r.issue_type.name)" }

Write-Host "========== Test 12b: ConvertType 990 Feature(2) -> Epic(1) =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/990/convert-type" '{"target_type_id":1}'
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else { Write-Host "OK: type now $($r.issue_type.name)" }

Write-Host ""
Write-Host "========== Test 13a: List issues in project 1 =========="
$r = Get-Json "http://localhost:8000/api/v1/issues?project_id=1&workspace_id=1&page=1&page_size=5"
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else {
    $count = if ($r -is [Array]) { $r.Count } else { 1 }
    Write-Host "OK: returned $count issues (page_size=5)"
    $r | Select-Object -First 3 | ForEach-Object { Write-Host ("  - id={0}, seq={1}, name={2}, type={3}" -f $_.id, $_.sequence_id, $_.name, $_.issue_type.name) }
}

Write-Host "========== Test 13b: Search issues =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/search?workspace_id=1&q=acceptance-test"
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else {
    $count = if ($r -is [Array]) { $r.Count } else { 1 }
    Write-Host "OK: search returned $count results"
    $r | Select-Object -First 5 | ForEach-Object { Write-Host ("  - id={0}, seq={1}, name={2}" -f $_.id, $_.sequence_id, $_.name) }
}

Write-Host "========== Test 13c: Filter issues by type_id=1 (Epic) =========="
$r = Get-Json "http://localhost:8000/api/v1/issues?project_id=1&workspace_id=1&type_id=1&page=1&page_size=10"
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else {
    $count = if ($r -is [Array]) { $r.Count } else { 1 }
    Write-Host "OK: filter type_id=1 returned $count issues"
}

Write-Host ""
Write-Host "========== Test 14a: Get activities for issue 990 =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/990/activities"
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else {
    $count = if ($r -is [Array]) { $r.Count } else { 1 }
    Write-Host "OK: $count activities recorded"
    $r | Select-Object -First 5 | ForEach-Object { Write-Host ("  - verb={0}, field={1}, old={2}, new={3}" -f $_.verb, $_.field, $_.old_value, $_.new_value) }
}

Write-Host ""
Write-Host "========== Test 14b: Delete issue 989 (broken issue with empty type) =========="
$r = Del-Json "http://localhost:8000/api/v1/issues/989"
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else { Write-Host "OK: deleted issue 989" }

Write-Host "========== Test 14c: Verify issue 989 is gone =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/989"
if (Is-Error $r) { Write-Host "OK: 404 as expected - $($r.__error)" } else { Write-Host "UNEXPECTED: issue still exists" }