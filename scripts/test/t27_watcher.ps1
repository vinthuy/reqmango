$BASE = 'http://localhost:8000/api/v1'
if (-not $BASE) { $BASE = if ($env:REQMANGO_API_URL) { $env:REQMANGO_API_URL.TrimEnd('/') } else { 'http://localhost:8000/api/v1' } }
$email = if ($env:REQMANGO_EMAIL) { $env:REQMANGO_EMAIL } else { 'demo@example.com' }
$password = if ($env:REQMANGO_PASSWORD) { $env:REQMANGO_PASSWORD } else { 'demo1234' }
$login = Invoke-RestMethod -Uri "$BASE/auth/login" -Method POST -ContentType 'application/json' -Body (@{ email = $email; password = $password } | ConvertTo-Json)
$TOKEN = $login.access_token
if (-not $TOKEN) { throw 'Login failed: no access_token (set REQMANGO_EMAIL / REQMANGO_PASSWORD or start backend with seed data)' }
$H = @{ 'Authorization' = "Bearer $TOKEN" }
$ISSUE = 990

Write-Host "===== Re-test Watcher ====="

Write-Host "Add Watcher:"
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$ISSUE/watch" -Headers $H -Method Post
    Write-Host "  [PASS] $($resp.message)"
} catch {
    Write-Host "  [FAIL] $($_.Exception.Message)"
}

Write-Host "List Watchers:"
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$ISSUE/watchers" -Headers $H -Method Get
    Write-Host "  [PASS] watchers=$($resp.Count)"
} catch {
    Write-Host "  [FAIL] $($_.Exception.Message)"
}

Write-Host "Remove Watcher:"
try {
    $resp = Invoke-RestMethod -Uri "$BASE/issues/$ISSUE/watch" -Headers $H -Method Delete
    Write-Host "  [PASS] $($resp.message)"
} catch {
    Write-Host "  [FAIL] $($_.Exception.Message)"
}