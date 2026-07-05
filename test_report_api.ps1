$ErrorActionPreference = "Stop"
$ts = [DateTimeOffset]::Now.ToUnixTimeMilliseconds()
$email = "rpttest${ts}@t.com"
$username = "rpttest${ts}"
$password = "E2eTest123!"

# Register
Write-Host "1. Register..."
$regBody = @{ email=$email; username=$username; password=$password; display_name="RPT Test" } | ConvertTo-Json
$reg = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/auth/register" -Method POST -ContentType "application/json" -Body $regBody -UseBasicParsing
Write-Host "   Register: $($reg.StatusCode)"

# Login
Write-Host "2. Login..."
$loginBody = @{ email=$email; password=$password } | ConvertTo-Json
$login = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/auth/login" -Method POST -ContentType "application/json" -Body $loginBody -UseBasicParsing
$token = ($login.Content | ConvertFrom-Json).access_token
Write-Host "   Token: $($token.Substring(0,20))..."

# Create workspace
Write-Host "3. Create workspace..."
$wsBody = @{ name="RPT WS"; slug="rpt-ws-$ts" } | ConvertTo-Json
$ws = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/workspaces" -Method POST -ContentType "application/json" -Body $wsBody -Headers @{Authorization="Bearer $token"} -UseBasicParsing
$wsId = ($ws.Content | ConvertFrom-Json).id
Write-Host "   Workspace ID: $wsId"

# Create project
Write-Host "4. Create project..."
$projBody = @{ name="RPT Project"; identifier="RPT"; description="Test" } | ConvertTo-Json
$proj = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/projects?workspace_id=$wsId" -Method POST -ContentType "application/json" -Body $projBody -Headers @{Authorization="Bearer $token"} -UseBasicParsing
$projId = ($proj.Content | ConvertFrom-Json).id
Write-Host "   Project ID: $projId"

$auth = @{Authorization="Bearer $token"; "Content-Type"="application/json"}

# Test 1: Distribution by state (no RQL)
Write-Host "`n=== Test: Distribution by state ==="
$body = @{ report_type="distribution"; group_by="state"; chart="bar" } | ConvertTo-Json
$r = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/projects/$projId/reports" -Method POST -Body $body -Headers $auth -UseBasicParsing
Write-Host "   Status: $($r.StatusCode)"
$d = $r.Content | ConvertFrom-Json
Write-Host "   Total: $($d.total), Labels: $($d.labels.Count)"

# Test 2: RQL = operator
Write-Host "`n=== Test: RQL = operator ==="
$body = @{ report_type="distribution"; group_by="state"; chart="bar"; rql='state = "Todo"' } | ConvertTo-Json
try {
    $r = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/projects/$projId/reports" -Method POST -Body $body -Headers $auth -UseBasicParsing
    Write-Host "   Status: $($r.StatusCode)"
    $d = $r.Content | ConvertFrom-Json
    Write-Host "   Total: $($d.total)"
} catch {
    Write-Host "   ERROR: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "   Body: $($reader.ReadToEnd())"
    }
}

# Test 3: RQL LIKE operator
Write-Host "`n=== Test: RQL LIKE operator ==="
$body = @{ report_type="distribution"; group_by="state"; chart="bar"; rql='name LIKE "test"' } | ConvertTo-Json
try {
    $r = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/projects/$projId/reports" -Method POST -Body $body -Headers $auth -UseBasicParsing
    Write-Host "   Status: $($r.StatusCode)"
} catch {
    Write-Host "   ERROR: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "   Body: $($reader.ReadToEnd())"
    }
}

# Test 4: RQL IS NULL
Write-Host "`n=== Test: RQL IS NULL ==="
$body = @{ report_type="distribution"; group_by="state"; chart="bar"; rql='assignee IS NULL' } | ConvertTo-Json
try {
    $r = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/projects/$projId/reports" -Method POST -Body $body -Headers $auth -UseBasicParsing
    Write-Host "   Status: $($r.StatusCode)"
} catch {
    Write-Host "   ERROR: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "   Body: $($reader.ReadToEnd())"
    }
}

# Test 5: RQL combined AND
Write-Host "`n=== Test: RQL combined AND ==="
$body = @{ report_type="distribution"; group_by="state"; chart="bar"; rql='priority IN ["urgent"] AND state != "Done"' } | ConvertTo-Json
try {
    $r = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/projects/$projId/reports" -Method POST -Body $body -Headers $auth -UseBasicParsing
    Write-Host "   Status: $($r.StatusCode)"
} catch {
    Write-Host "   ERROR: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "   Body: $($reader.ReadToEnd())"
    }
}

# Test 6: Saved report CRUD
Write-Host "`n=== Test: Create saved report ==="
$body = @{ name="Test Report"; report_type="distribution"; group_by="state"; chart_type="bar"; rql='priority = "high"' } | ConvertTo-Json
try {
    $r = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/projects/$projId/saved-reports" -Method POST -Body $body -Headers $auth -UseBasicParsing
    Write-Host "   Status: $($r.StatusCode)"
    $d = $r.Content | ConvertFrom-Json
    Write-Host "   ID: $($d.id), Name: $($d.name)"
    $savedId = $d.id
} catch {
    Write-Host "   ERROR: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "   Body: $($reader.ReadToEnd())"
    }
}

Write-Host "`n=== Test: List saved reports ==="
try {
    $r = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/projects/$projId/saved-reports" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
    Write-Host "   Status: $($r.StatusCode)"
    $d = $r.Content | ConvertFrom-Json
    Write-Host "   Count: $($d.Count)"
} catch {
    Write-Host "   ERROR: $($_.Exception.Message)"
}

Write-Host "`n=== Test: Delete saved report ==="
try {
    $r = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/projects/$projId/saved-reports/$savedId" -Method DELETE -Headers @{Authorization="Bearer $token"} -UseBasicParsing
    Write-Host "   Status: $($r.StatusCode)"
} catch {
    Write-Host "   ERROR: $($_.Exception.Message)"
}

Write-Host "`n=== ALL TESTS DONE ==="
