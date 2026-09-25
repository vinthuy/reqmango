$BASE = 'http://localhost:8000/api/v1'
if (-not $BASE) { $BASE = if ($env:REQMANGO_API_URL) { $env:REQMANGO_API_URL.TrimEnd('/') } else { 'http://localhost:8000/api/v1' } }
$email = if ($env:REQMANGO_EMAIL) { $env:REQMANGO_EMAIL } else { 'demo@example.com' }
$password = if ($env:REQMANGO_PASSWORD) { $env:REQMANGO_PASSWORD } else { 'demo1234' }
$login = Invoke-RestMethod -Uri "$BASE/auth/login" -Method POST -ContentType 'application/json' -Body (@{ email = $email; password = $password } | ConvertTo-Json)
$TOKEN = $login.access_token
if (-not $TOKEN) { throw 'Login failed: no access_token (set REQMANGO_EMAIL / REQMANGO_PASSWORD or start backend with seed data)' }
$H = @{ 'Authorization' = "Bearer $TOKEN" }
$resp = Invoke-RestMethod -Uri "$BASE/issues/990" -Headers $H
Write-Host "Issue 990:"
Write-Host "  id: $($resp.id)"
Write-Host "  name: $($resp.name)"
Write-Host "  type_id: $($resp.issue_type.id)"
Write-Host "  type_name: $($resp.issue_type.name)"
Write-Host "  watcher_count: $($resp.watcher_count)"
Write-Host "  archived_at: $($resp.archived_at)"
Write-Host "  priority: $($resp.priority)"