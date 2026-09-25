# C2 Dashboard ai_summary widget smoke
# Usage: powershell -NoProfile -ExecutionPolicy Bypass -File backend/test_ai_c2_smoke.ps1
$ErrorActionPreference = 'Continue'
$apiBase = 'http://localhost:8000/api/v1'
$projectId = 4

function Invoke-Api($method, $path, $bodyObj = $null) {
  $uri = $apiBase + $path
  $params = @{ Uri = $uri; Method = $method; Headers = $script:h; TimeoutSec = 180 }
  if ($null -ne $bodyObj) {
    $json = ($bodyObj | ConvertTo-Json -Depth 12 -Compress)
    $params.Body = [System.Text.Encoding]::UTF8.GetBytes($json)
    $params.ContentType = 'application/json; charset=utf-8'
  }
  try { return Invoke-RestMethod @params }
  catch {
    return @{ __error = $_.Exception.Message; __body = $_.ErrorDetails.Message }
  }
}

$login = Invoke-Api POST '/auth/login' @{ email = 'admin@reqmango.com'; password = 'demo1234' }
if (-not $login.access_token) { Write-Host "LOGIN FAIL"; exit 2 }
$script:h = @{ Authorization = ('Bearer ' + $login.access_token) }

$dashboards = Invoke-Api GET "/projects/$projectId/dashboards"
if ($dashboards.__error) { Write-Host "FAIL list dashboards $($dashboards.__body)"; exit 1 }
$list = @($dashboards)
if ($list.Count -eq 0 -or -not $list[0].id) {
  $created = Invoke-Api POST "/projects/$projectId/dashboards" @{ name = 'C2 Smoke Dashboard' }
  if ($created.__error -or -not $created.id) { Write-Host "FAIL create dashboard $($created.__body)"; exit 1 }
  $dashId = $created.id
} else {
  $dashId = $list[0].id
}
Write-Host ("dashboard_id=" + $dashId)

$widget = Invoke-Api POST "/projects/$projectId/dashboards/$dashId/widgets" @{
  widget_type = 'ai_summary'
  title = 'AI Project Summary'
  config = @{}
  position = @{ x = 0; y = 0; w = 6; h = 4 }
}
if ($widget.__error -or -not $widget.id) {
  Write-Host "FAIL add widget $($widget.__body)"
  exit 1
}
Write-Host ("widget_id=" + $widget.id)
Write-Host ("widget_type=" + $widget.widget_type)
if ($widget.widget_type -ne 'ai_summary') { Write-Host 'FAIL wrong type'; exit 1 }

$full = Invoke-Api GET "/projects/$projectId/dashboards/$dashId/full"
if ($full.__error) { Write-Host "FAIL get full $($full.__body)"; exit 1 }
$wd = @($full.widget_data) | Where-Object { $_.widget_id -eq $widget.id }
if (-not $wd) { Write-Host 'FAIL missing widget_data entry'; exit 1 }
$data = $wd.data
if (-not $data) { Write-Host 'FAIL empty data'; exit 1 }

# Accept either successful summary payload or readable error (AI may be offline)
$hasSummary = $null -ne $data.summary
$hasError = $null -ne $data.error
if (-not $hasSummary -and -not $hasError) {
  Write-Host ("FAIL unexpected data keys: " + (($data | Get-Member -MemberType NoteProperty).Name -join ','))
  exit 1
}
if ($hasSummary) {
  Write-Host ("summary_len=" + ([string]$data.summary).Length)
  Write-Host ("insights_count=" + @($data.insights).Count)
} else {
  Write-Host ("error_ok=" + $data.error)
}

# Cleanup smoke widget
$null = Invoke-Api DELETE "/projects/$projectId/dashboards/$dashId/widgets/$($widget.id)"

Write-Host 'C2_SMOKE_PASS'
exit 0
