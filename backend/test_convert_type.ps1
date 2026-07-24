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
function Get-Json($url) {
    try { return Invoke-RestMethod -Uri $url -Method GET -Headers $h }
    catch { $eb = Get-ErrorBody $_.Exception; return @{ __error = $_.Exception.Message; __body = $eb } }
}

Write-Host "========== Pre-check: Issue 990 current type =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/990"
if (Is-Error $r) {
    Write-Host "ERR: $($r.__error) $($r.__body)"
} else {
    $tid = if ($r.issue_type_id) { $r.issue_type_id } else { "(nil)" }
    $tname = if ($r.issue_type.name) { $r.issue_type.name } else { "(none)" }
    Write-Host "Issue 990: type_id=$tid, type_name=$tname"
}

Write-Host ""
Write-Host "========== Test 12a: ConvertType 990 -> Feature(2) =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/990/convert-type" '{"target_type_id":2}'
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else {
    $tname = if ($r.issue_type.name) { $r.issue_type.name } else { "(none)" }
    $tid = if ($r.issue_type_id) { $r.issue_type_id } else { "(nil)" }
    Write-Host "OK response: type_id=$tid, type_name=$tname"
}

Write-Host ""
Write-Host "========== Test 12a-verify: GET /issues/990 =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/990"
if (Is-Error $r) {
    Write-Host "ERR: $($r.__error)"
} else {
    $tid = if ($r.issue_type_id) { $r.issue_type_id } else { "(nil)" }
    $tname = if ($r.issue_type.name) { $r.issue_type.name } else { "(none)" }
    Write-Host "After convert: type_id=$tid, type_name=$tname"
    if ($tid -eq 2) {
        Write-Host "PASS: issue_type_id correctly updated to 2 (Feature)"
    } else {
        Write-Host "FAIL: issue_type_id is $tid, expected 2"
    }
}

Write-Host ""
Write-Host "========== Test 12b: ConvertType 990 back -> Epic(1) =========="
$r = Post-Json "http://localhost:8000/api/v1/issues/990/convert-type" '{"target_type_id":1}'
if (Is-Error $r) { Write-Host "ERR: $($r.__error) $($r.__body)" } else {
    $tname = if ($r.issue_type.name) { $r.issue_type.name } else { "(none)" }
    Write-Host "OK response: type_name=$tname"
}

Write-Host ""
Write-Host "========== Test 12b-verify: GET /issues/990 =========="
$r = Get-Json "http://localhost:8000/api/v1/issues/990"
if (Is-Error $r) {
    Write-Host "ERR: $($r.__error)"
} else {
    $tid = if ($r.issue_type_id) { $r.issue_type_id } else { "(nil)" }
    $tname = if ($r.issue_type.name) { $r.issue_type.name } else { "(none)" }
    Write-Host "After revert: type_id=$tid, type_name=$tname"
    if ($tid -eq 1) {
        Write-Host "PASS: issue_type_id correctly reverted to 1 (Epic)"
    } else