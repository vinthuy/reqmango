$BASE = 'http://localhost:8000/api/v1'
$TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODUwMjY5MTUsImlhdCI6MTc4NDQyMjExNSwic3ViIjoiMSJ9.hMyKlBim8wLpqFoKg4Ig0e-Xdd3Plf3J4G4yFFnI1TQ'
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