$ErrorActionPreference = "Continue"
Set-Location d:\code\reqmango\frontend

Write-Host "=== Running workspace-settings tests ==="
npx playwright test e2e/workspace-settings-e2e.spec.ts --reporter=list 2>&1 | Out-File -FilePath d:\code\reqmango\test_results_ws.txt
Get-Content d:\code\reqmango\test_results_ws.txt | Select-String "passed|failed|ok \d"

Write-Host "=== Running i18n tests ==="
npx playwright test e2e/i18n-full-e2e.spec.ts --reporter=list 2>&1 | Out-File -FilePath d:\code\reqmango\test_results_i18n.txt
Get-Content d:\code\reqmango\test_results_i18n.txt | Select-String "passed|failed|ok \d"
