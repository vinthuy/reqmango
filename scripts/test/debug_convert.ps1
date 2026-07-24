$BASE = 'http://localhost:8000/api/v1'
$TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODUwMjY5MTUsImlhdCI6MTc4NDQyMjExNSwic3ViIjoiMSJ9.hMyKlBim8wLpqFoKg4Ig0e-Xdd3Plf3J4G4yFFnI1TQ'
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