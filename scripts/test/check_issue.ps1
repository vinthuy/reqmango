$BASE = 'http://localhost:8000/api/v1'
$TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODUwMjY5MTUsImlhdCI6MTc4NDQyMjExNSwic3ViIjoiMSJ9.hMyKlBim8wLpqFoKg4Ig0e-Xdd3Plf3J4G4yFFnI1TQ'
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