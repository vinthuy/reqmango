$BASE = 'http://localhost:8000/api/v1'
if (-not $BASE) { $BASE = if ($env:REQMANGO_API_URL) { $env:REQMANGO_API_URL.TrimEnd('/') } else { 'http://localhost:8000/api/v1' } }
$email = if ($env:REQMANGO_EMAIL) { $env:REQMANGO_EMAIL } else { 'demo@example.com' }
$password = if ($env:REQMANGO_PASSWORD) { $env:REQMANGO_PASSWORD } else { 'demo1234' }
$login = Invoke-RestMethod -Uri "$BASE/auth/login" -Method POST -ContentType 'application/json' -Body (@{ email = $email; password = $password } | ConvertTo-Json)
$TOKEN = $login.access_token
if (-not $TOKEN) { throw 'Login failed: no access_token (set REQMANGO_EMAIL / REQMANGO_PASSWORD or start backend with seed data)' }
$H = @{ 'Authorization' = "Bearer $TOKEN"; 'Content-Type' = 'application/json' }

Write-Host "===== Issue Types in Project 1 ====="
$types = Invoke-RestMethod -Uri "$BASE/issue-types?workspace_id=1&project_id=1" -Headers $H
$types | ForEach-Object { Write-Host "  id=$($_.id) name=$($_.name) level=$($_.level) parent=$($_.parent_type_id)" }

Write-Host ""
Write-Host "===== Sub-issues of 990 ====="
$subs = Invoke-RestMethod -Uri "$BASE/issues/990/children" -Headers $H
$subs | ForEach-Object { Write-Host "  id=$($_.id) name=$($_.name) type=$($_.issue_type.name)" }

Write-Host ""
Write-Host "===== Try ConvertType with verbose error ====="
try {
    $body = @{ type_id = 2 } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE/issues/990/convert-type" -Headers $H -Method Post -Body $body
    Write-Host "Success: $($resp | ConvertTo-Json -Compress)"
} catch {
    Write-Host "Error: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response body: $responseBody"
    }
}