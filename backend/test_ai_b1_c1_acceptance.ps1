# B1+C1 AI automation templates + project summary automated acceptance
# Usage: powershell -NoProfile -ExecutionPolicy Bypass -File backend/test_ai_b1_c1_acceptance.ps1

$ErrorActionPreference = 'Continue'
$apiBase = if ($env:REQMANGO_API_URL) { $env:REQMANGO_API_URL.TrimEnd('/') } else { 'http://localhost:8000/api/v1' }
$email = if ($env:REQMANGO_EMAIL) { $env:REQMANGO_EMAIL } else { 'admin@reqmango.com' }
$password = if ($env:REQMANGO_PASSWORD) { $env:REQMANGO_PASSWORD } else { 'demo1234' }
$repoRoot = Split-Path $PSScriptRoot -Parent
if (-not (Test-Path (Join-Path $repoRoot 'frontend'))) { $repoRoot = 'D:\code\reqmango' }
$feRoot = if ($env:REQMANGO_FE_ROOT) { $env:REQMANGO_FE_ROOT } else { Join-Path $repoRoot 'frontend' }
$outDir = Join-Path $repoRoot 'docs\dev\acceptance'
$stamp = Get-Date -Format 'yyyy-MM-dd'
$jsonOut = Join-Path $outDir "$stamp-ai-b1-c1-auto-run.json"
$mdOut = Join-Path $outDir "$stamp-ai-b1-c1.md"

$results = New-Object System.Collections.Generic.List[object]

function Add-Result($id, $status, $note) {
  $results.Add([pscustomobject]@{ Id = $id; Status = $status; Note = $note })
  Write-Host ("[{0}] {1}: {2}" -f $status, $id, $note)
}

function Get-ErrorBody($ex) {
  try {
    if ($ex.ErrorDetails -and $ex.ErrorDetails.Message) { return $ex.ErrorDetails.Message }
    return $ex.Message
  } catch { return $ex.Message }
}

function Invoke-Api($method, $path, $bodyObj = $null) {
  $uri = if ($path.StartsWith('http')) { $path } else { "$apiBase$path" }
  $params = @{ Uri = $uri; Method = $method; Headers = $script:h; TimeoutSec = 180 }
  if ($null -ne $bodyObj) {
    $json = if ($bodyObj -is [string]) { $bodyObj } else { ($bodyObj | ConvertTo-Json -Depth 12 -Compress) }
    $params.Body = [System.Text.Encoding]::UTF8.GetBytes($json)
    $params.ContentType = 'application/json; charset=utf-8'
  }
  try { return Invoke-RestMethod @params }
  catch {
    $code = 0
    try { $code = [int]$_.Exception.Response.StatusCode } catch {}
    return @{ __error = $_.Exception.Message; __body = (Get-ErrorBody $_); __status = $code }
  }
}

function Test-Ok($r) { return -not ($r -is [hashtable] -and $r.ContainsKey('__error')) }

function To-List($r, $wrapKey = $null) {
  if ($null -eq $r) { return @() }
  if ($r -is [hashtable] -and $r.ContainsKey('__error')) { return @() }
  if ($r -is [System.Array]) { return @($r) }
  if ($wrapKey -and $r.PSObject.Properties.Name -contains $wrapKey -and $null -ne $r.$wrapKey) {
    return @($r.$wrapKey)
  }
  return @($r)
}

function File-Has($rel, $pattern) {
  $p = Join-Path $feRoot $rel
  if (-not (Test-Path $p)) { return $false }
  return [bool](Select-String -Path $p -Pattern $pattern -Quiet)
}

function File-Text($rel) {
  $p = Join-Path $feRoot $rel
  if (-not (Test-Path $p)) { return '' }
  return Get-Content -Path $p -Raw -Encoding UTF8
}

Write-Host "=== B1+C1 automated acceptance ==="
Write-Host "API: $apiBase"
Write-Host "FE:  $feRoot"
Write-Host ""

# ---------- Static evidence ----------
$settingsRel = 'src\views\ProjectSettings.vue'
$settingsText = File-Text $settingsRel
$aiNames = @('intakeTriage', 'riskOnCreate', 'specDraftOnCreate')
$aiFound = @($aiNames | Where-Object { $settingsText -match [regex]::Escape("name: '$_'") })
if ($aiFound.Count -ge 3) {
  Add-Result 'B1-1' 'PASS' ("AI templates present: " + ($aiFound -join ', '))
} else {
  Add-Result 'B1-1' 'FAIL' ("expected 3 AI templates, found: " + ($aiFound -join ', '))
}

# Templates always shown: empty-state is a sibling; template loop is not nested under it.
$tmplIdx = $settingsText.IndexOf('v-for="template in automationTemplates"')
$emptyIdx = $settingsText.IndexOf('v-if="automations.length === 0"')
$spaceIdx = $settingsText.IndexOf('class="space-y-3"')
# Prefer structural: empty block then space-y-3 section containing templates
$okSibling = ($tmplIdx -ge 0) -and ($emptyIdx -ge 0) -and ($spaceIdx -gt $emptyIdx) -and ($tmplIdx -gt $spaceIdx)
if ($okSibling) {
  Add-Result 'B1-2' 'PASS' 'template grid sibling of empty-state (always shown)'
} elseif ($tmplIdx -ge 0 -and ($emptyIdx -lt 0 -or $tmplIdx -lt $emptyIdx)) {
  Add-Result 'B1-2' 'PASS' 'template grid present without empty-state gate'
} else {
  Add-Result 'B1-2' 'FAIL' 'template visibility still depends on empty automations list'
}

if (File-Has 'src\components\AutomationRuleBuilder.vue' 'dispatch_agent') {
  Add-Result 'B1-3' 'PASS' 'AutomationRuleBuilder exposes dispatch_agent (DEF-03)'
} else {
  Add-Result 'B1-3' 'FAIL' 'AutomationRuleBuilder missing dispatch_agent'
}

$svcPath = Join-Path $repoRoot 'backend\internal\ai\service\ai_service.go'
$svcText = if (Test-Path $svcPath) { Get-Content $svcPath -Raw -Encoding UTF8 } else { '' }
if ($svcText -match 'func \(s \*AIService\) Analyze' -and $svcText -match 'analyzeIssue' -and $svcText -match 'analyzeProject' -and $svcText -match 'IssueID > 0') {
  Add-Result 'DEF-02' 'PASS' 'Analyze routes issue vs project by IssueID'
} else {
  Add-Result 'DEF-02' 'FAIL' 'Analyze issue-scoped branch missing'
}

$projText = File-Text 'src\views\Project.vue'
if (($projText -match 'analyzeWithAI\(projectId\.value,\s*0\)') -and ($projText -match 'createPage') -and ($projText -match 'aiSummarySavePage')) {
  Add-Result 'C1-1' 'PASS' 'Project.vue has project AI summary + save-as-page'
} else {
  Add-Result 'C1-1' 'FAIL' 'Project.vue missing summary/save-as-page wiring'
}

# No primary product entry revival for Loop/Harness in project/settings surfaces
$navHits = @()
foreach ($rel in @(
  'src\views\Project.vue',
  'src\views\ProjectSettings.vue',
  'src\locales\zh-CN.json',
  'src\locales\en-US.json'
)) {
  $t = File-Text $rel
  if ($t -match '(?i)Harness|agent-loop|Agent Loop|对抗评审') { $navHits += $rel }
}
if ($navHits.Count -eq 0) {
  Add-Result 'B1C1-5' 'PASS' 'no Loop/Harness product copy in Project/Settings/i18n primary surfaces'
} else {
  Add-Result 'B1C1-5' 'FAIL' ("Loop/Harness copy found in: " + ($navHits -join ', '))
}

$phase12 = Join-Path $outDir '2026-09-25-ai-pm-phase1-2-auto-run.json'
if (Test-Path $phase12) {
  Add-Result 'B1C1-4' 'PASS' 'Phase 1-2 auto-run artifact present (prior conditional pass)'
} else {
  Add-Result 'B1C1-4' 'PARTIAL' 'Phase 1-2 auto-run JSON missing; prior acceptance doc may still apply'
}

# ---------- API ----------
$login = Invoke-Api POST '/auth/login' @{ email = $email; password = $password }
if (-not (Test-Ok $login) -or -not $login.access_token) {
  Write-Host "FATAL: login failed for $email :: $($login.__body)"
  exit 2
}
$script:h = @{ Authorization = ("Bearer " + $login.access_token) }
Write-Host "Logged in as $email"

$workspaces = Invoke-Api GET '/workspaces'
$wsList = To-List $workspaces 'workspaces'
if ($wsList.Count -eq 0) { Write-Host 'FATAL no workspaces'; exit 2 }
$ws = $wsList[0]
$wsId = [int]$ws.id
$wsSlug = $ws.slug
Write-Host "Workspace id=$wsId slug=$wsSlug"

$projects = Invoke-Api GET "/workspaces/$wsId/projects"
$projList = To-List $projects 'projects'
if ($projList.Count -eq 0) {
  $projects = Invoke-Api GET "/projects?workspace_id=$wsId"
  $projList = To-List $projects 'projects'
}
if ($projList.Count -eq 0) { Write-Host 'FATAL no projects'; exit 2 }
$project = $projList[0]
$projectId = [int]$project.id
Write-Host "Project id=$projectId name=$($project.name)"

# Ensure PM agents + create one AI template rule (riskOnCreate - unique suffix)
$agents = Invoke-Api POST "/workspaces/$wsId/agents/ensure-pm" @{}
$agentList = To-List $agents 'agents'
if ($agentList.Count -eq 0) { $agentList = To-List $agents }
$riskAgent = $agentList | Where-Object { $_.name -eq '交付风险' -or $_.name -match 'risk' } | Select-Object -First 1
if (-not $riskAgent) { $riskAgent = $agentList | Select-Object -First 1 }
if ($riskAgent -and $riskAgent.id) {
  Add-Result 'B1-API-0' 'PASS' ("ensure-pm ok; agent id=$($riskAgent.id) name=$($riskAgent.name)")
} else {
  Add-Result 'B1-API-0' 'FAIL' "ensure-pm returned no agents: $($agents.__body)"
}

$ruleName = "B1C1-auto-risk-$(Get-Date -Format 'HHmmss')"
$actions = @(
  @{ type = 'dispatch_agent'; value = [int]$riskAgent.id; field = 'B1C1 acceptance: analyze delivery risk for high-priority issues.' }
)
$createRule = Invoke-Api POST "/projects/$projectId/automations" @{
  name = $ruleName
  description = 'B1+C1 automated acceptance rule'
  trigger_type = (@{ type = 'issue.created' } | ConvertTo-Json -Compress)
  conditions = (@(
    @{ field = 'priority'; operator = 'in'; value = @('urgent', 'high') }
  ) | ConvertTo-Json -Compress -Depth 6)
  actions = ($actions | ConvertTo-Json -Compress -Depth 6)
}
if (Test-Ok $createRule) {
  $ruleId = $createRule.id
  if (-not $ruleId -and $createRule.automation) { $ruleId = $createRule.automation.id }
  Add-Result 'B1-API-1' 'PASS' ("created automation id=$ruleId name=$ruleName with dispatch_agent")
  if ($ruleId) {
    $null = Invoke-Api DELETE "/projects/$projectId/automations/$ruleId"
  }
} else {
  Add-Result 'B1-API-1' 'FAIL' ("create automation failed: $($createRule.__body)")
}

# Project analyze (issue_id=0)
$analyzeProj = Invoke-Api POST "/projects/$projectId/ai/analyze?issue_id=0&workspace_id=$wsId" @{}
if (Test-Ok $analyzeProj -and ($analyzeProj.summary -or $analyzeProj.insights)) {
  $sumLen = 0
  if ($analyzeProj.summary) { $sumLen = ([string]$analyzeProj.summary).Length }
  Add-Result 'C1-API-1' 'PASS' ("project analyze ok; summary_len=$sumLen")
} elseif (Test-Ok $analyzeProj) {
  Add-Result 'C1-API-1' 'PARTIAL' 'project analyze HTTP 200 but empty summary/insights'
} else {
  Add-Result 'C1-API-1' 'FAIL' ("project analyze failed: $($analyzeProj.__body)")
}

# Issue-scoped analyze (DEF-02 runtime)
$issues = Invoke-Api GET "/projects/$projectId/issues?workspace_id=$wsId&per_page=5"
$issueList = To-List $issues 'issues'
if ($issueList.Count -eq 0) {
  $issues = Invoke-Api GET "/issues?project_id=$projectId&workspace_id=$wsId&per_page=5"
  $issueList = To-List $issues 'issues'
}
if ($issueList.Count -gt 0) {
  $issueId = [int]$issueList[0].id
  $analyzeIssue = Invoke-Api POST "/projects/$projectId/ai/analyze?issue_id=$issueId&workspace_id=$wsId" @{}
  if (Test-Ok $analyzeIssue -and $analyzeIssue.summary) {
    $looksProjectWide = ([string]$analyzeIssue.summary) -match '(?i)整个项目|project-wide|overall project'
    if ($looksProjectWide) {
      Add-Result 'C1-API-2' 'PARTIAL' "issue analyze ok but summary may be project-wide (issue_id=$issueId)"
    } else {
      Add-Result 'C1-API-2' 'PASS' ("issue-scoped analyze ok issue_id=$issueId")
    }
  } elseif (Test-Ok $analyzeIssue) {
    Add-Result 'C1-API-2' 'PARTIAL' "issue analyze 200 empty body issue_id=$issueId"
  } else {
    Add-Result 'C1-API-2' 'FAIL' ("issue analyze failed: $($analyzeIssue.__body)")
  }
} else {
  Add-Result 'C1-API-2' 'SKIP' 'no issues in project to test issue-scoped analyze'
}

# Save as Page (API equivalent of C1 UI)
$pageTitle = "AI Project Summary (B1C1 auto) $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
$pageContent = if ($analyzeProj.summary) {
  $insights = @()
  if ($analyzeProj.insights) { $insights = @($analyzeProj.insights | ForEach-Object { "- $_" }) }
  @"
# $pageTitle

$($analyzeProj.summary)

## Insights
$($insights -join "`n")
"@
} else {
  "# $pageTitle`n`nAutomated acceptance placeholder page for B1+C1."
}
$createPage = Invoke-Api POST "/projects/$projectId/pages?workspace_id=$wsId" @{
  name = $pageTitle
  title = $pageTitle
  content = $pageContent
  description = $pageContent
}
$pageId = $null
if (Test-Ok $createPage) {
  $pageId = $createPage.id
  if (-not $pageId -and $createPage.page) { $pageId = $createPage.page.id }
  Add-Result 'C1-API-3' 'PASS' ("created page id=$pageId title=$pageTitle")
} else {
  Add-Result 'C1-API-3' 'FAIL' ("create page failed: $($createPage.__body)")
}

if ($pageId) {
  $pages = Invoke-Api GET "/projects/$projectId/pages?workspace_id=$wsId"
  $pageList = To-List $pages 'pages'
  $found = $pageList | Where-Object { $_.id -eq $pageId -or $_.title -eq $pageTitle -or $_.name -eq $pageTitle }
  if ($found) {
    Add-Result 'C1-API-4' 'PASS' "page visible in project pages list"
  } else {
    $getOne = Invoke-Api GET "/projects/$projectId/pages/$pageId"
    if (Test-Ok $getOne) {
      Add-Result 'C1-API-4' 'PASS' "page fetchable by id (list shape differed)"
    } else {
      Add-Result 'C1-API-4' 'FAIL' 'created page not found in list/get'
    }
  }
}

# ---------- Summary ----------
$pass = @($results | Where-Object { $_.Status -eq 'PASS' }).Count
$fail = @($results | Where-Object { $_.Status -eq 'FAIL' }).Count
$partial = @($results | Where-Object { $_.Status -eq 'PARTIAL' }).Count
$skip = @($results | Where-Object { $_.Status -eq 'SKIP' }).Count
$verdict = if ($fail -eq 0 -and $partial -eq 0) { 'PASS' } elseif ($fail -eq 0) { 'CONDITIONAL_PASS' } else { 'FAIL' }

Write-Host ""
Write-Host "=== SUMMARY pass=$pass fail=$fail partial=$partial skip=$skip verdict=$verdict ==="

$payload = [ordered]@{
  generated_at = (Get-Date).ToString('o')
  verdict = $verdict
  summary = @{ pass = $pass; fail = $fail; partial = $partial; skip = $skip }
  workspace_id = $wsId
  project_id = $projectId
  page_id = $pageId
  workspace_slug = $wsSlug
  results = $results
}
if (-not (Test-Path $outDir)) { New-Item -ItemType Directory -Path $outDir -Force | Out-Null }
($payload | ConvertTo-Json -Depth 8) | Set-Content -Path $jsonOut -Encoding UTF8

$md = @"
# B1+C1 Automated Acceptance ($stamp)

**Verdict:** $verdict  
**API:** $apiBase  
**Workspace:** $wsId ($wsSlug) / **Project:** $projectId  
**Counts:** pass=$pass fail=$fail partial=$partial skip=$skip

| Id | Status | Note |
|----|--------|------|
$(($results | ForEach-Object { "| $($_.Id) | $($_.Status) | $($_.Note) |" }) -join "`n")

## Criteria map

| # | Standard | Coverage |
|---|----------|----------|
| 1 | >=3 AI automation templates | B1-1, B1-API-1 |
| 2 | Templates always visible | B1-2 |
| 3 | Project summary -> Page | C1-1, C1-API-1, C1-API-3/4 |
| 4 | Phase 1-2 done | B1C1-4 |
| 5 | No Loop/Harness product revival | B1C1-5 |
| DEF-02/03 | Issue analyze + RuleBuilder | DEF-02, B1-3, C1-API-2 |

Artifact: ``$jsonOut``
"@
$md | Set-Content -Path $mdOut -Encoding UTF8
Write-Host "Wrote $jsonOut"
Write-Host "Wrote $mdOut"

if ($fail -gt 0) { exit 1 } else { exit 0 }
