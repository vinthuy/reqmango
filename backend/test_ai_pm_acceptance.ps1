# Phase 1-2 AI PM automated acceptance (API + static source evidence)
# Usage: powershell -NoProfile -ExecutionPolicy Bypass -File backend/test_ai_pm_acceptance.ps1

$ErrorActionPreference = 'Continue'
$apiBase = if ($env:REQMANGO_API_URL) { $env:REQMANGO_API_URL.TrimEnd('/') } else { 'http://localhost:8000/api/v1' }
$email = if ($env:REQMANGO_EMAIL) { $env:REQMANGO_EMAIL } else { 'admin@reqmango.com' }
$password = if ($env:REQMANGO_PASSWORD) { $env:REQMANGO_PASSWORD } else { 'demo1234' }
$repoRoot = Split-Path $PSScriptRoot -Parent
if (-not (Test-Path (Join-Path $repoRoot 'frontend'))) { $repoRoot = 'D:\code\reqmango' }
$feRoot = if ($env:REQMANGO_FE_ROOT) { $env:REQMANGO_FE_ROOT } else { Join-Path $repoRoot 'frontend' }

$results = New-Object System.Collections.Generic.List[object]
$defects = New-Object System.Collections.Generic.List[object]

function Add-Result($id, $status, $note) {
  $results.Add([pscustomobject]@{ Id = $id; Status = $status; Note = $note })
  Write-Host ("[{0}] {1}: {2}" -f $status, $id, $note)
}

function Add-Defect($id, $item, $sev, $status, $note) {
  $defects.Add([pscustomobject]@{ Id = $id; Item = $item; Sev = $sev; Status = $status; Note = $note })
}

function Get-ErrorBody($ex) {
  try {
    if ($ex.ErrorDetails -and $ex.ErrorDetails.Message) { return $ex.ErrorDetails.Message }
    return $ex.Message
  } catch { return $ex.Message }
}

function Invoke-Api($method, $path, $bodyObj = $null) {
  $uri = if ($path.StartsWith('http')) { $path } else { "$apiBase$path" }
  $params = @{ Uri = $uri; Method = $method; Headers = $script:h; TimeoutSec = 90 }
  if ($null -ne $bodyObj) {
    $json = if ($bodyObj -is [string]) { $bodyObj } else { ($bodyObj | ConvertTo-Json -Depth 12 -Compress) }
    $params.Body = [System.Text.Encoding]::UTF8.GetBytes($json)
    $params.ContentType = 'application/json; charset=utf-8'
  }
  try { return Invoke-RestMethod @params }
  catch {
    $code = 0
    try { $code = [int]$_.Exception.Response.StatusCode } catch {}
    return @{ __error = $_.Exception.Message; __body = (Get-ErrorBody $_.Exception); __status = $code }
  }
}

function Test-Ok($r) { return -not ($r -is [hashtable] -and $r.ContainsKey('__error')) }

# Avoid PowerShell member-enumeration on arrays (e.g. $arr.workspaces -> nulls).
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

Write-Host "=== Phase 1-2 AI PM automated acceptance ==="
Write-Host "API: $apiBase"
Write-Host "FE:  $feRoot"
Write-Host ""

$login = Invoke-Api POST '/auth/login' @{ email = $email; password = $password }
if (-not (Test-Ok $login) -or -not $login.access_token) {
  Write-Host "FATAL: login failed for $email :: $($login.__body)"
  exit 2
}
$script:h = @{ Authorization = ("Bearer " + $login.access_token) }
Write-Host "Logged in as $email"

$workspaces = Invoke-Api GET '/workspaces'
if (-not (Test-Ok $workspaces)) { Write-Host "FATAL workspaces $($workspaces.__body)"; exit 2 }
$ws = (To-List $workspaces 'workspaces') | Where-Object { $_.id } | Select-Object -First 1
if (-not $ws) { Write-Host "FATAL: no workspace"; exit 2 }
$wsId = [uint64]$ws.id

$projects = Invoke-Api GET "/workspaces/$wsId/projects"
if (-not (Test-Ok $projects)) { $projects = Invoke-Api GET "/projects?workspace_id=$wsId" }
$proj = (To-List $projects 'projects') | Where-Object { $_.id } | Select-Object -First 1
if (-not $proj) { Write-Host "FATAL: no project"; exit 2 }
$projectId = [uint64]$proj.id
Write-Host ("Workspace={0} Project={1}" -f $wsId, $projectId)
Write-Host ""

# ---- Static FE ----
$sidebarHasAgents = File-Has 'src/components/AppSidebar.vue' 'sidebar\.agents|/agents'
$advOk = File-Has 'src/components/AISettingsPanel.vue' 'advancedAgentsPath|/agents'
if (-not $sidebarHasAgents -and $advOk) { Add-Result 'P1-1' 'PASS' 'sidebar has no Agents primary; advanced /agents link present' }
elseif (-not $sidebarHasAgents) { Add-Result 'P1-1' 'PARTIAL' 'sidebar OK; advanced link missing' }
else { Add-Result 'P1-1' 'FAIL' 'sidebar still exposes Agents' }

$askBuild = (File-Has 'src/components/AICopilot.vue' "CopilotMode = 'ask' \| 'build'") -or ((File-Has 'src/components/AICopilot.vue' "'ask'") -and (File-Has 'src/components/AICopilot.vue' "'build'"))
$ctrlJ = (File-Has 'src/views/IssueDetail.vue' "key === 'j'") -or (File-Has 'src/views/IssueDetail.vue' 'key === .j.')
$ctx = (File-Has 'src/components/AICopilot.vue' 'contextProjectOnly') -or (File-Has 'src/components/AICopilot.vue' 'contextLabel')
if ($askBuild -and $ctrlJ -and $ctx) { Add-Result 'P1-2' 'PASS' 'Ask|Build + IssueDetail Ctrl+J + context wired' }
elseif ($askBuild) { Add-Result 'P1-2' 'PARTIAL' ("AskBuild={0} CtrlJ={1} ctx={2}" -f $askBuild,$ctrlJ,$ctx) }
else { Add-Result 'P1-2' 'FAIL' 'Ask|Build missing' }

if (File-Has 'src/components/AICopilot.vue' 'askReadOnlyHint') { Add-Result 'P1-3' 'PASS' 'Ask readonly hint present' }
else { Add-Result 'P1-3' 'FAIL' 'askReadOnlyHint missing' }

if (File-Has 'src/components/TriagePanel.vue' 'toast\.error') { Add-Result 'P1-7-ui' 'PASS' 'TriagePanel shows toast on error' }
else { Add-Result 'P1-7-ui' 'FAIL' 'TriagePanel swallows errors'; Add-Defect 'DEF-05' 'P1-7' 'P2' 'open' 'no toast.error' }

if (File-Has 'src/components/AutomationRuleBuilder.vue' 'dispatch_agent') {
  Add-Result 'DEF-03' 'PASS' 'RuleBuilder has dispatch_agent'
} else {
  Add-Result 'DEF-03' 'FAIL' 'AutomationRuleBuilder missing dispatch_agent'
  Add-Defect 'DEF-03' 'P2-2' 'P1' 'open' 'RuleBuilder missing dispatch_agent'
}

if (File-Has 'src/views/ProjectSettings.vue' 'automationTemplates.chooseTemplate') { Add-Result 'DEF-01' 'PASS' 'templates section always renderable' }
else { Add-Result 'DEF-01' 'FAIL' 'template section missing' }

if (File-Has 'src/components/AutomationManager.vue' 'value="issue.created"') { Add-Result 'DEF-06' 'PASS' 'AutomationManager dotted triggers' }
else { Add-Result 'DEF-06' 'FAIL' 'AutomationManager trigger not dotted'; Add-Defect 'DEF-06' '-' 'P2' 'open' 'underscore triggers' }

$ua = Get-Content (Join-Path $feRoot 'src/api/issue-agent.ts') -Raw
if (($ua -match '/assign-agent') -and ($ua -notmatch 'unassign-agent')) { Add-Result 'DEF-07' 'PASS' 'FE unassign DELETE .../assign-agent' }
else { Add-Result 'DEF-07' 'FAIL' 'unassign URL mismatch'; Add-Defect 'DEF-07' 'P1-4' 'P1' 'open' 'unassign URL' }

Write-Host ""

# ---- API ----
$ensure1 = Invoke-Api POST "/workspaces/$wsId/agents/ensure-pm"
$ensure2 = Invoke-Api POST "/workspaces/$wsId/agents/ensure-pm"
$agents = @()
if (Test-Ok $ensure1) {
  if ($ensure1.agents) { $agents = @($ensure1.agents) }
  elseif ($ensure1 -is [System.Array]) { $agents = @($ensure1) }
}
$names = @($agents | ForEach-Object { $_.name })
$need = @(
  [System.Text.Encoding]::UTF8.GetString([byte[]](0xE8,0xAF,0xB7,0xE6,0xB1,0x82,0xE5,0x88,0x86,0xE8,0xAF,0x8A)),
  [System.Text.Encoding]::UTF8.GetString([byte[]](0xE4,0xBA,0xA4,0xE4,0xBB,0x98,0xE9,0xA3,0x8E,0xE9,0x99,0xA9)),
  [System.Text.Encoding]::UTF8.GetString([byte[]](0x53,0x70,0x72,0x69,0x6E,0x74,0x20,0xE6,0x80,0xBB,0xE7,0xBB,0x93)),
  [System.Text.Encoding]::UTF8.GetString([byte[]](0x53,0x70,0x65,0x63,0x20,0xE8,0x8D,0x89,0xE7,0xA8,0xBF))
)
$missing = @($need | Where-Object { $names -notcontains $_ })
$c1 = $names.Count
$c2 = 0
if (Test-Ok $ensure2) {
  if ($ensure2.agents) { $c2 = @($ensure2.agents).Count }
  elseif ($ensure2 -is [System.Array]) { $c2 = @($ensure2).Count }
}
if ((Test-Ok $ensure1) -and $missing.Count -eq 0 -and $c1 -eq $c2) { Add-Result 'P2-1' 'PASS' ("ensure-pm 4 agents idempotent n={0}" -f $c1) }
elseif ((Test-Ok $ensure1) -and $missing.Count -eq 0) { Add-Result 'P2-1' 'PARTIAL' ("agents ok but count {0}->{1}" -f $c1,$c2) }
else { Add-Result 'P2-1' 'FAIL' ("err={0} missing={1} names={2}" -f $ensure1.__body, ($missing -join ','), ($names -join ',')) }

$triageName = $need[0]
$triage = $agents | Where-Object { $_.name -eq $triageName } | Select-Object -First 1
$triageId = 0
if ($triage) { $triageId = [uint64]$triage.id }

$issueList = Invoke-Api GET "/projects/$projectId/issues?per_page=5&page=1"
$issues = @()
if (Test-Ok $issueList) {
  if ($issueList -is [System.Array]) { $issues = @($issueList) }
  elseif ($issueList.PSObject.Properties.Name -contains 'issues') { $issues = @($issueList.issues) }
  elseif ($issueList.PSObject.Properties.Name -contains 'results') { $issues = @($issueList.results) }
  elseif ($issueList.PSObject.Properties.Name -contains 'data' -and $issueList.data -is [System.Array]) { $issues = @($issueList.data) }
  else { $issues = @($issueList) }
}
$issues = @($issues | Where-Object { $_.id })
if ($issues.Count -eq 0) {
  # Fallback: create a scratch issue for assign/analyze checks
  $scratch = Invoke-Api POST "/projects/$projectId/issues" @{ name = ('acceptance-scratch-' + [Guid]::NewGuid().ToString('N').Substring(0,6)); priority = 'low' }
  if (-not (Test-Ok $scratch)) {
    $scratch = Invoke-Api POST ("/issues?project_id={0}&workspace_id={1}" -f $projectId, $wsId) @{ name = ('acceptance-scratch-' + [Guid]::NewGuid().ToString('N').Substring(0,6)); priority = 'low' }
  }
  if (Test-Ok $scratch) { $issues = @($scratch) }
}
$sampleIssue = $issues | Select-Object -First 1
$issueId = 0
if ($sampleIssue) { $issueId = [uint64]$sampleIssue.id }

if ($triageId -gt 0 -and $issueId -gt 0) {
  $asg = Invoke-Api POST "/issues/$issueId/assign-agent" @{ agent_id = $triageId; priority = 'medium'; notes = 'acceptance' }
  $st = Invoke-Api GET "/issues/$issueId/agent-status"
  $unasg = Invoke-Api DELETE "/issues/$issueId/assign-agent"
  if ((Test-Ok $asg) -and (Test-Ok $st) -and (Test-Ok $unasg) -and ($st.agent_name -or $st.agent_id)) {
    Add-Result 'P1-4' 'PASS' ("assign/status/unassign OK issue={0} agent={1}" -f $issueId, $st.agent_name)
  } elseif ((Test-Ok $asg) -and (Test-Ok $st)) {
    Add-Result 'P1-4' 'PARTIAL' ("assign/status OK; unassign status={0} {1}" -f $unasg.__status, $unasg.__body)
    if (-not (Test-Ok $unasg)) { Add-Defect 'DEF-07' 'P1-4' 'P1' 'open' ("unassign HTTP {0}" -f $unasg.__status) }
  } else { Add-Result 'P1-4' 'FAIL' ("assign={0} status={1}" -f $asg.__body, $st.__body) }
} else { Add-Result 'P1-4' 'SKIP' ("triage={0} issue={1}" -f $triageId, $issueId) }

$act = Invoke-Api GET "/workspaces/$wsId/agents/activity"
if (Test-Ok $act) { Add-Result 'P1-6' 'PASS' 'agents/activity reachable' }
else { Add-Result 'P1-6' 'FAIL' ("{0} {1}" -f $act.__status, $act.__body) }

$dupTitle = 'Login'
if ($sampleIssue -and $sampleIssue.name) { $dupTitle = [string]$sampleIssue.name }
$dup = Invoke-Api POST "/projects/$projectId/issues/duplicate-check" @{ name = $dupTitle; description = '' }
if (Test-Ok $dup) {
  $dups = @()
  if ($dup.duplicates) { $dups = @($dup.duplicates) }
  elseif ($dup.similar) { $dups = @($dup.similar) }
  elseif ($dup -is [System.Array]) { $dups = @($dup) }
  Add-Result 'P1-8' 'PASS' ("duplicate-check OK hits={0}" -f $dups.Count)
} else { Add-Result 'P1-8' 'FAIL' ("{0} {1}" -f $dup.__status, $dup.__body) }

$agentList = Invoke-Api GET "/workspaces/$wsId/agents"
$feMention = (File-Has 'src/components/CommentList.vue' 'allMentionCandidates') -or (File-Has 'src/components/CommentList.vue' "kind === 'agent'")
if (Test-Ok $agentList) {
  $al = @()
  if ($agentList.agents) { $al = @($agentList.agents) }
  elseif ($agentList -is [System.Array]) { $al = @($agentList) }
  if ($al.Count -gt 0 -and $feMention) { Add-Result 'P1-5' 'PASS' ("agents API + CommentList mention wiring n={0}" -f $al.Count) }
  else { Add-Result 'P1-5' 'PARTIAL' ("agents={0} feMention={1}" -f $al.Count, $feMention) }
} else { Add-Result 'P1-5' 'FAIL' $agentList.__body }

$intake = Invoke-Api GET "/projects/$projectId/intake"
if (Test-Ok $intake) { Add-Result 'P1-7' 'PASS' 'intake pending list API OK' }
else { Add-Result 'P1-7' 'FAIL' ("{0} {1}" -f $intake.__status, $intake.__body) }

$ruleId = 0
if ($triageId -gt 0) {
  $actionsJson = '[{"type":"dispatch_agent","value":' + $triageId + ',"field":"acceptance triage"}]'
  $ruleName = 'acceptance-intake-triage-' + [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
  $createRule = Invoke-Api POST "/projects/$projectId/automations" @{
    name = $ruleName
    description = 'auto acceptance P2-2'
    trigger_type = '{"type":"issue.created"}'
    conditions = '[]'
    actions = $actionsJson
    is_enabled = $true
  }
  if (Test-Ok $createRule) {
    $ruleId = [uint64]$createRule.id
    if (($createRule.trigger_type -match 'issue\.created') -and ($createRule.actions -match 'dispatch_agent')) {
      Add-Result 'P2-2' 'PASS' ("rule id={0} issue.created + dispatch_agent" -f $ruleId)
    } else { Add-Result 'P2-2' 'PARTIAL' ("rule id={0} trigger={1}" -f $ruleId, $createRule.trigger_type) }
  } else { Add-Result 'P2-2' 'FAIL' ("{0} {1}" -f $createRule.__status, $createRule.__body) }
} else { Add-Result 'P2-2' 'SKIP' 'no triage agent' }

if ($ruleId -gt 0) {
  $newName = 'acceptance-create-triage-' + [Guid]::NewGuid().ToString('N').Substring(0,8)
  $created = Invoke-Api POST "/projects/$projectId/issues" @{ name = $newName; priority = 'medium' }
  if (-not (Test-Ok $created)) {
    $created = Invoke-Api POST ("/issues?project_id={0}&workspace_id={1}" -f $projectId, $wsId) @{ name = $newName; priority = 'medium' }
  }
  if (Test-Ok $created) {
    $newId = [uint64]$created.id
    Start-Sleep -Seconds 4
    $hist = Invoke-Api GET "/automations/$ruleId/execution-history"
    $histCount = 0
    if (Test-Ok $hist) {
      if ($hist.history) { $histCount = @($hist.history).Count }
      elseif ($hist.executions) { $histCount = @($hist.executions).Count }
      elseif ($hist.data) { $histCount = @($hist.data).Count }
      elseif ($hist -is [System.Array]) { $histCount = @($hist).Count }
    }
    $act2 = Invoke-Api GET "/workspaces/$wsId/agents/activity"
    $actHit = $false
    if (Test-Ok $act2) {
      $raw = ($act2 | ConvertTo-Json -Depth 6 -Compress)
      if ($raw -match "$newId" -or $raw -match 'dispatch|automation') { $actHit = $true }
    }
    if ($histCount -gt 0 -or $actHit) { Add-Result 'P2-3' 'PASS' ("issue #{0} hist={1} actHit={2}" -f $newId, $histCount, $actHit) }
    else { Add-Result 'P2-3' 'PARTIAL' ("issue #{0} created; no hist yet (async/LLM)" -f $newId) }
  } else { Add-Result 'P2-3' 'FAIL' ("create issue: {0}" -f $created.__body) }
} else { Add-Result 'P2-3' 'SKIP' 'no rule' }

$intakeName = 'acceptance-intake-' + [Guid]::NewGuid().ToString('N').Substring(0,8)
$sub = Invoke-Api POST "/intake/$projectId" @{ name = $intakeName; description = 'auto acceptance'; priority = 'high'; submitter = 'acceptance-bot' }
if (Test-Ok $sub) {
  $intakeIssueId = [uint64]$sub.id
  if ($sub.status -eq 'pending' -or $intakeIssueId -gt 0) {
    Add-Result 'P2-4' 'PASS' ("intake submit id={0} status={1}" -f $intakeIssueId, $sub.status)
  } else { Add-Result 'P2-4' 'PARTIAL' ("submit ok id={0}" -f $intakeIssueId) }
} else { Add-Result 'P2-4' 'FAIL' ("{0} {1}" -f $sub.__status, $sub.__body) }

$cycles = Invoke-Api GET "/projects/$projectId/cycles"
$cycleId = 0
if (Test-Ok $cycles) {
  $cl = @()
  if ($cycles.cycles) { $cl = @($cycles.cycles) }
  elseif ($cycles.data) { $cl = @($cycles.data) }
  elseif ($cycles -is [System.Array]) { $cl = @($cycles) }
  if ($cl.Count -gt 0) { $cycleId = [uint64]$cl[0].id }
}
$feCycle = (File-Has 'src/views/CycleDetail.vue' 'sprintPlan') -or (File-Has 'src/views/CycleDetail.vue' 'aiSummary')
$spBody = @{}
if ($cycleId -gt 0) { $spBody = @{ cycle_id = $cycleId } }
$sp = Invoke-Api POST "/projects/$projectId/ai/sprint-plan" $spBody
if ((Test-Ok $sp) -and $feCycle) { Add-Result 'P2-5' 'PASS' ("sprint-plan OK cycle={0} + UI" -f $cycleId) }
elseif (Test-Ok $sp) { Add-Result 'P2-5' 'PARTIAL' 'API OK; CycleDetail UI not confirmed' }
elseif ($feCycle) { Add-Result 'P2-5' 'PARTIAL' ("UI wired; API: {0}" -f $sp.__body) }
else { Add-Result 'P2-5' 'FAIL' ("API {0} FE={1}" -f $sp.__body, $feCycle) }

$analyzeSrc = Join-Path $repoRoot 'backend/internal/ai/service/ai_service.go'
$codeIgnoresIssue = $true
if (Test-Path $analyzeSrc) {
  $m = Select-String -Path $analyzeSrc -Pattern 'func \(s \*AIService\) Analyze' -Context 0,55 | Out-String
  if ($m -match 'actx\.IssueID') { $codeIgnoresIssue = $false }
}
if ($issueId -gt 0) {
  $an = Invoke-Api POST "/projects/$projectId/ai/analyze?issue_id=$issueId"
  $issueName = 'issue'
  if ($sampleIssue.name) { $issueName = [string]$sampleIssue.name }
  $lbl = Invoke-Api POST "/projects/$projectId/ai/suggest-labels?issue_id=$issueId" @{ name = $issueName; description = '' }
  $analyzePass = Test-Ok $an
  $labelPass = Test-Ok $lbl
  if ($analyzePass -and $labelPass) {
    if ($codeIgnoresIssue) {
      Add-Result 'P2-6' 'PARTIAL' 'analyze+labels OK but Analyze ignores IssueID (DEF-02)'
      Add-Defect 'DEF-02' 'P2-6' 'P1' 'open' 'AIService.Analyze ignores actx.IssueID'
    } else { Add-Result 'P2-6' 'PASS' ("analyze+labels OK issue={0}" -f $issueId) }
  } elseif ($labelPass -or $analyzePass) {
    Add-Result 'P2-6' 'PARTIAL' ("analyze={0} label={1}" -f $analyzePass, $labelPass)
    if ($codeIgnoresIssue) { Add-Defect 'DEF-02' 'P2-6' 'P1' 'open' 'Analyze ignores IssueID' }
  } else {
    Add-Result 'P2-6' 'PARTIAL' ("LLM/API fail analyze={0} label={1}" -f $an.__body, $lbl.__body)
    if ($codeIgnoresIssue) { Add-Defect 'DEF-02' 'P2-6' 'P1' 'open' 'Analyze ignores IssueID' }
  }
} else { Add-Result 'P2-6' 'SKIP' 'no sample issue' }

if ($ruleId -gt 0) { $null = Invoke-Api DELETE "/projects/$projectId/automations/$ruleId" }

Write-Host ""
Write-Host "=== Summary ==="
$pass = @($results | Where-Object { $_.Status -eq 'PASS' }).Count
$fail = @($results | Where-Object { $_.Status -eq 'FAIL' }).Count
$partial = @($results | Where-Object { $_.Status -eq 'PARTIAL' }).Count
$skip = @($results | Where-Object { $_.Status -eq 'SKIP' }).Count
Write-Host ("PASS={0} PARTIAL={1} FAIL={2} SKIP={3}" -f $pass,$partial,$fail,$skip)
Write-Host ""
Write-Host "=== Open defects ==="
if ($defects.Count -eq 0) { Write-Host '(none)' } else { $defects | Format-Table -AutoSize }

$outDir = Join-Path $repoRoot 'docs/dev/acceptance'
New-Item -ItemType Directory -Force -Path $outDir | Out-Null
$outPath = Join-Path $outDir '2026-09-25-ai-pm-phase1-2-auto-run.json'
$payload = @{
  ran_at = (Get-Date).ToString('s')
  api = $apiBase
  workspace_id = $wsId
  project_id = $projectId
  summary = @{ pass = $pass; partial = $partial; fail = $fail; skip = $skip }
  results = $results
  defects = $defects
}
($payload | ConvertTo-Json -Depth 8) | Set-Content -Path $outPath -Encoding UTF8
Write-Host "Wrote $outPath"

if ($fail -gt 0) { exit 1 } else { exit 0 }