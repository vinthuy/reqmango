$apiBase = if ($env:REQMANGO_API_URL) { $env:REQMANGO_API_URL.TrimEnd('/') } else { 'http://localhost:8000/api/v1' }
$email = if ($env:REQMANGO_EMAIL) { $env:REQMANGO_EMAIL } else { 'demo@example.com' }
$password = if ($env:REQMANGO_PASSWORD) { $env:REQMANGO_PASSWORD } else { 'demo1234' }
$login = Invoke-RestMethod -Uri "$apiBase/auth/login" -Method POST -ContentType 'application/json' -Body (@{ email = $email; password = $password } | ConvertTo-Json)
$token = $login.access_token
if (-not $token) { throw 'Login failed: no access_token (set REQMANGO_EMAIL / REQMANGO_PASSWORD or start backend with seed data)' }
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