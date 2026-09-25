# B2 NL automation-preview smoke (ASCII prompts for PowerShell encoding)
# Usage: powershell -NoProfile -ExecutionPolicy Bypass -File backend/test_ai_b2_smoke.ps1
$ErrorActionPreference = 'Continue'
$apiBase = 'http://localhost:8000/api/v1'

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
$null = Invoke-Api POST '/workspaces/1/agents/ensure-pm' @{}

$r = Invoke-Api POST '/projects/4/ai/automation-preview' @{
  prompt = 'When an urgent or high priority issue is created, dispatch the delivery-risk agent and add a reminder comment. Do not use webhook.'
}
if ($r.__error) { Write-Host "FAIL $($r.__body)"; exit 1 }
Write-Host ("name=" + $r.name)
Write-Host ("trigger=" + $r.trigger_type)
Write-Host ("actions_count=" + @($r.actions).Count)
if (-not $r.name -or @($r.actions).Count -lt 1) { Write-Host 'FAIL incomplete'; exit 1 }

$r2 = Invoke-Api POST '/projects/4/ai/automation-preview' @{
  prompt = 'Every morning at 9am send a webhook to https://example.com/hook'
}
if (-not $r2.__error) {
  $bad = @($r2.actions) | Where-Object { $_.type -eq 'call_webhook' }
  if ($bad -or $r2.trigger_type -eq 'scheduled') { Write-Host 'FAIL webhook/scheduled leaked'; exit 1 }
  Write-Host ("unsupported_ok trigger=" + $r2.trigger_type)
}

Write-Host 'B2_SMOKE_PASS'
exit 0
