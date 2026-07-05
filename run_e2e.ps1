$ErrorActionPreference = "SilentlyContinue"
Set-Location d:\code\reqmango\frontend

$tests = @(
    @{Name="ws-api"; Pattern="API:"},
    @{Name="ws-members"; Pattern="members section|add member|member role"},
    @{Name="ws-roles"; Pattern="roles section|system roles"},
    @{Name="ws-rest"; Pattern="integrations section|plugins section|AI settings|work item types|custom fields|relations section|templates section"}
)

foreach ($t in $tests) {
    Write-Host "--- $($t.Name) ---"
    npx playwright test e2e/workspace-settings-e2e.spec.ts -g "$($t.Pattern)" --reporter=list *>$null
}
Write-Host "WS tests done"

npx playwright test e2e/i18n-full-e2e.spec.ts --reporter=list *>$null
Write-Host "i18n tests done"
