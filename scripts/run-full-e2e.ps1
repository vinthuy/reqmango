<#
.SYNOPSIS
  ReqMango 全量端到端覆盖测试 —— 一键运行。

.DESCRIPTION
  拉起完整服务栈并运行全部测试套件：
    1. 检查 PostgreSQL（127.0.0.1:5432），未运行则尝试启动服务 PostgreSQL-18
    2. 启动 Go 后端（backend/，:8000）
    3. 构建前端并启动 vite preview（frontend/，:5173）
    4. 运行后端 Go 全量测试
    5. 运行 tests/ Playwright 模块 E2E 套件（chrome）
    6. 运行 frontend/e2e Playwright 套件（默认仅 chromium）
    7. 运行 frontend vitest 单元测试
    8. 聚合 Playwright JSON 报告为覆盖矩阵（scripts/e2e-coverage-report.mjs）

.PARAMETER FrontendProject
  frontend/e2e 使用的 Playwright project，默认 chromium。设为 all 则跑 chromium+firefox+webkit。

.PARAMETER Workers
  frontend/e2e 的并行 worker 数，默认 2。历史经验：worker 过多会压垮 Vite。

.PARAMETER SkipFrontendE2E
  跳过 frontend/e2e 套件。

.PARAMETER SkipVitest
  跳过前端单元测试。

.EXAMPLE
  pwsh -File scripts/run-full-e2e.ps1
  pwsh -File scripts/run-full-e2e.ps1 -FrontendProject all -Workers 1

.NOTES
  本脚本在「主机无法创建子进程」的会话中编写，因此**尚未端到端实测**。
  首次运行时请留意每个阶段打印的 PASS/FAIL 与 test-artifacts/ 下的日志。
  前置：backend/.env 存在；PostgreSQL 中有 reqmango 库与测试账号。
#>

[CmdletBinding()]
param(
  [string]$FrontendProject = 'chromium',
  [int]$Workers = 2,
  [switch]$SkipFrontendE2E,
  [switch]$SkipVitest
)

$ErrorActionPreference = 'Continue'

$Root      = Split-Path -Parent $PSScriptRoot
$Backend   = Join-Path $Root 'backend'
$Frontend  = Join-Path $Root 'frontend'
$Tests     = Join-Path $Root 'tests'
$Artifacts = Join-Path $Root 'test-artifacts'
New-Item -ItemType Directory -Force -Path $Artifacts | Out-Null

$script:Failed = @()

function Write-Stage($msg) {
  Write-Host ''
  Write-Host ('=' * 72) -ForegroundColor Cyan
  Write-Host "  $msg" -ForegroundColor Cyan
  Write-Host ('=' * 72) -ForegroundColor Cyan
}

function Wait-Port {
  param([int]$Port, [int]$TimeoutSec = 120, [string]$Name = 'service')
  $deadline = (Get-Date).AddSeconds($TimeoutSec)
  while ((Get-Date) -lt $deadline) {
    $c = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($c) { Write-Host "  [ok] $Name 已监听 :$Port"; return $true }
    Start-Sleep -Seconds 2
  }
  Write-Host "  [FAIL] $Name 未能在 ${TimeoutSec}s 内监听 :$Port" -ForegroundColor Red
  return $false
}

# ---------------------------------------------------------------- 1. PostgreSQL
Write-Stage '1/8 检查 PostgreSQL'
$pg = Get-NetTCPConnection -LocalPort 5432 -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
if (-not $pg) {
  Write-Host '  5432 未监听，尝试启动服务 PostgreSQL-18 ...'
  try { Start-Service -Name 'PostgreSQL-18' -ErrorAction Stop } catch { Write-Host "  启动失败: $_" -ForegroundColor Yellow }
  Start-Sleep -Seconds 5
}
if (Wait-Port -Port 5432 -TimeoutSec 30 -Name 'PostgreSQL') {
  Write-Host '  PostgreSQL 就绪'
} else {
  $script:Failed += 'PostgreSQL'
  Write-Host '  无法继续：数据库不可用' -ForegroundColor Red
  exit 1
}

# ---------------------------------------------------------------- 2. 后端
Write-Stage '2/8 启动 Go 后端 (:8000)'
$backendOut = Join-Path $Artifacts 'backend.out.log'
$backendErr = Join-Path $Artifacts 'backend.err.log'
$backendProc = Start-Process -FilePath 'go' `
  -ArgumentList 'run', './cmd/server/' `
  -WorkingDirectory $Backend `
  -RedirectStandardOutput $backendOut -RedirectStandardError $backendErr `
  -PassThru -WindowStyle Hidden
Write-Host "  后端 PID = $($backendProc.Id)，日志 $backendOut"

if (-not (Wait-Port -Port 8000 -TimeoutSec 180 -Name '后端')) {
  $script:Failed += 'backend'
  Write-Host '  后端启动失败，末尾日志：' -ForegroundColor Red
  if (Test-Path $backendErr) { Get-Content $backendErr -Tail 30 }
  exit 1
}

# ---------------------------------------------------------------- 3. 前端
Write-Stage '3/8 构建并启动前端 (:5173, 生产构建 + vite preview)'
Push-Location $Frontend
if (-not (Test-Path (Join-Path $Frontend 'dist\index.html'))) {
  Write-Host '  dist/ 不存在，执行 vite build ...'
  & npx.cmd vite build 2>&1 | Tee-Object -FilePath (Join-Path $Artifacts 'vite-build.log') | Select-Object -Last 5
} else {
  Write-Host '  dist/ 已存在，跳过构建（如需重建请先删除 frontend/dist）'
}
Pop-Location

$previewOut = Join-Path $Artifacts 'preview.out.log'
$previewErr = Join-Path $Artifacts 'preview.err.log'
$previewProc = Start-Process -FilePath 'npx.cmd' `
  -ArgumentList 'vite', 'preview' `
  -WorkingDirectory $Frontend `
  -RedirectStandardOutput $previewOut -RedirectStandardError $previewErr `
  -PassThru -WindowStyle Hidden
Write-Host "  preview PID = $($previewProc.Id)，日志 $previewOut"

if (-not (Wait-Port -Port 5173 -TimeoutSec 120 -Name '前端 preview')) {
  $script:Failed += 'frontend'
  Write-Host '  前端启动失败，末尾日志：' -ForegroundColor Red
  if (Test-Path $previewErr) { Get-Content $previewErr -Tail 30 }
  exit 1
}

# 冒烟：确认页面与 /api 代理都通
try {
  $page = Invoke-WebRequest -Uri 'http://localhost:5173/' -UseBasicParsing -TimeoutSec 15
  Write-Host "  [ok] GET / -> $($page.StatusCode)"
  $login = Invoke-RestMethod -Uri 'http://localhost:5173/api/v1/auth/login' -Method POST `
    -Body (@{ email = 'qa_tester@reqmango.com'; password = 'Test@12345' } | ConvertTo-Json) `
    -ContentType 'application/json' -TimeoutSec 15
  Write-Host '  [ok] /api 代理可用（qa_tester 登录成功）'
} catch {
  Write-Host "  [FAIL] 冒烟失败: $_" -ForegroundColor Red
  $script:Failed += 'smoke'
}

try {
  # frontend/e2e 的 pages-e2e.spec.ts 需要 TEST_TOKEN，否则整体跳过
  $admin = Invoke-RestMethod -Uri 'http://localhost:8000/api/v1/auth/login' -Method POST `
    -Body (@{ email = 'admin@reqmango.com'; password = 'demo1234' } | ConvertTo-Json) `
    -ContentType 'application/json' -TimeoutSec 15
  $env:TEST_TOKEN = $admin.access_token
  Write-Host '  [ok] 已设置 TEST_TOKEN（admin@reqmango.com）'
} catch {
  Write-Host "  [warn] 无法获取 admin token，pages-e2e 将被跳过: $_" -ForegroundColor Yellow
}

# ---------------------------------------------------------------- 4. 后端测试
Write-Stage '4/8 后端 Go 全量测试'
Push-Location $Backend
& go test ./... -count=1 2>&1 | Tee-Object -FilePath (Join-Path $Artifacts 'go-test.log')
if ($LASTEXITCODE -ne 0) { $script:Failed += 'go-test' }
Pop-Location

# ---------------------------------------------------------------- 5. tests/
Write-Stage '5/8 tests/ 模块 E2E 套件（chrome）'
$testsJson = Join-Path $Artifacts 'playwright-tests.json'
$env:PLAYWRIGHT_JSON_OUTPUT_NAME = $testsJson
Push-Location $Tests
& npx.cmd playwright test --reporter=list,json 2>&1 | Tee-Object -FilePath (Join-Path $Artifacts 'playwright-tests.log')
if ($LASTEXITCODE -ne 0) { $script:Failed += 'tests-e2e' }
Pop-Location
Remove-Item Env:\PLAYWRIGHT_JSON_OUTPUT_NAME -ErrorAction SilentlyContinue

# ---------------------------------------------------------------- 6. frontend/e2e
if (-not $SkipFrontendE2E) {
  Write-Stage "6/8 frontend/e2e 套件（project=$FrontendProject, workers=$Workers）"
  $feJson = Join-Path $Artifacts 'playwright-frontend.json'
  $env:PLAYWRIGHT_JSON_OUTPUT_NAME = $feJson
  Push-Location $Frontend
  if ($FrontendProject -eq 'all') {
    & npx.cmd playwright test --workers=$Workers --reporter=list,json 2>&1 |
      Tee-Object -FilePath (Join-Path $Artifacts 'playwright-frontend.log')
  } else {
    & npx.cmd playwright test --project=$FrontendProject --workers=$Workers --reporter=list,json 2>&1 |
      Tee-Object -FilePath (Join-Path $Artifacts 'playwright-frontend.log')
  }
  if ($LASTEXITCODE -ne 0) { $script:Failed += 'frontend-e2e' }
  Pop-Location
  Remove-Item Env:\PLAYWRIGHT_JSON_OUTPUT_NAME -ErrorAction SilentlyContinue
} else {
  Write-Stage '6/8 frontend/e2e —— 已按要求跳过'
}

# ---------------------------------------------------------------- 7. vitest
if (-not $SkipVitest) {
  Write-Stage '7/8 frontend vitest 单元测试'
  Push-Location $Frontend
  & npx.cmd vitest run 2>&1 | Tee-Object -FilePath (Join-Path $Artifacts 'vitest.log')
  if ($LASTEXITCODE -ne 0) { $script:Failed += 'vitest' }
  Pop-Location
} else {
  Write-Stage '7/8 vitest —— 已按要求跳过'
}

# ---------------------------------------------------------------- 8. 聚合
Write-Stage '8/8 聚合覆盖矩阵'
Push-Location $Root
if (Test-Path $testsJson) {
  node scripts/e2e-coverage-report.mjs $testsJson --label 'tests/ 模块 E2E 套件' `
    --json-out (Join-Path $Artifacts 'aggregate-tests.json') |
    Tee-Object -FilePath (Join-Path $Artifacts 'coverage-summary.md')
}
if ((-not $SkipFrontendE2E) -and (Test-Path $feJson)) {
  node scripts/e2e-coverage-report.mjs $feJson --label 'frontend/e2e 套件' `
    --json-out (Join-Path $Artifacts 'aggregate-frontend.json') |
    Tee-Object -FilePath (Join-Path $Artifacts 'coverage-summary.md') -Append
}
Pop-Location

# ---------------------------------------------------------------- 收尾
Write-Stage '清理后台进程'
foreach ($p in @($previewProc, $backendProc)) {
  if ($p -and -not $p.HasExited) {
    try { Stop-Process -Id $p.Id -Force; Write-Host "  已停止 PID $($p.Id)" } catch { Write-Host "  停止 PID $($p.Id) 失败: $_" -ForegroundColor Yellow }
  }
}
# vite preview / go run 可能派生子进程，按端口兜底清理
foreach ($port in @(5173, 8000)) {
  $c = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($c) { try { Stop-Process -Id $c.OwningProcess -Force; Write-Host "  已停止占用 :$port 的 PID $($c.OwningProcess)" } catch {} }
}

Write-Host ''
if ($script:Failed.Count -eq 0) {
  Write-Host '全部套件执行完毕，无失败阶段。' -ForegroundColor Green
  Write-Host "产物目录: $Artifacts"
  exit 0
} else {
  Write-Host ("存在失败阶段: " + ($script:Failed -join ', ')) -ForegroundColor Red
  Write-Host "请查看 $Artifacts 下的日志定位。" -ForegroundColor Red
  exit 1
}
