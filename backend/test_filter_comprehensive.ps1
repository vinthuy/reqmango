# Comprehensive Filter Function Test
# Project: CORE (id=15, identifier=CORE, workspace_id=8)
# Total issues: 60
# States: 85(backlog), 86(unstarted), 87(started), 88(started), 89(completed), 90(cancelled)
# Priorities: none(4), low(15), medium(26), high(12), urgent(3)
# Labels: 163-177 (15 labels)
# Cycles: 67-72 (6 sprints)
# Modules: 88-96 (9 modules)
# Members: user_id 49-58+ (workspace 8)
#
# NOTE: Chinese keywords are built via [char] codes to avoid PowerShell 5.1
# UTF-8-no-BOM file encoding issues.

$ErrorActionPreference = 'Continue'
$BASE = 'http://localhost:8000/api/v1'
$PASS = 0
$FAIL = 0
$SKIP = 0

# Build Chinese keywords via Unicode codepoints (avoids file encoding issues)
$KW_MONITOR = [char]0x76D1 + [char]0x63A7        # jian kong
$KW_AS_USER = [char]0x4F5C + [char]0x4E3A         # zuo wei
$KW_DASHBOARD = 'Dashboard'
$KW_DOCKER = 'Docker'

# Login
$loginResp = Invoke-RestMethod -Uri "$BASE/auth/login" -Method Post -ContentType 'application/json' -Body '{"email":"admin@reqmango.com","password":"demo1234"}'
$HEADERS = @{ Authorization = "Bearer $($loginResp.access_token)"; 'Content-Type' = 'application/json; charset=utf-8' }
$PROJECT_ID = 15

function Send-RQL {
    param([string]$RQL)
    $bodyObj = @{ entity = 'issue'; project_id = $script:PROJECT_ID; rql = $RQL; page = 1; page_size = 100 }
    $body = $bodyObj | ConvertTo-Json -Compress
    $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($body)
    return Invoke-RestMethod -Uri "$BASE/rql/search" -Method Post -Headers $HEADERS -Body $bodyBytes -ContentType 'application/json; charset=utf-8'
}

function Run-Test {
    param(
        [string]$Name,
        [string]$RQL,
        [int]$ExpectedMin = -1,
        [int]$ExpectedMax = -1,
        [int]$ExpectedExact = -1
    )
    try {
        $resp = Send-RQL -RQL $RQL
        if (-not $resp.success) {
            Write-Host "  FAIL: $Name -> API error: $($resp.error.message)" -ForegroundColor Red
            $script:FAIL++
            return
        }
        $items = $resp.data.items
        $count = if ($items) { @($items).Count } else { 0 }
        $ok = $true
        if ($ExpectedExact -ge 0) {
            $ok = ($count -eq $ExpectedExact)
        } elseif ($ExpectedMax -ge 0) {
            $ok = ($count -ge $ExpectedMin) -and ($count -le $ExpectedMax)
        } elseif ($ExpectedMin -ge 0) {
            $ok = ($count -ge $ExpectedMin)
        }
        if ($ok) {
            Write-Host "  PASS: $Name (count=$count)" -ForegroundColor Green
            $script:PASS++
        } else {
            $rangeStr = if ($ExpectedExact -ge 0) { "exact=$ExpectedExact" } else { "min=$ExpectedMin max=$ExpectedMax" }
            Write-Host "  FAIL: $Name (count=$count, expected $rangeStr)" -ForegroundColor Red
            $script:FAIL++
        }
    } catch {
        Write-Host "  FAIL: $Name -> Exception: $_" -ForegroundColor Red
        $script:FAIL++
    }
}

Write-Host "=== Comprehensive Filter Function Test ===" -ForegroundColor Cyan
Write-Host ""

# ========== 1. sequence_id (Number field) ==========
Write-Host "[1] sequence_id filter (number field)" -ForegroundColor Yellow
Run-Test "sequence_id = 1" "sequence_id = 1" -ExpectedExact 1
Run-Test "sequence_id = 2" "sequence_id = 2" -ExpectedExact 1
Run-Test "sequence_id = 60 (last)" "sequence_id = 60" -ExpectedMin 1
Run-Test "sequence_id != 1" "sequence_id != 1" -ExpectedMin 2
Run-Test "sequence_id = 999 (not exist)" "sequence_id = 999" -ExpectedExact 0
Run-Test "sequence_id > 58" "sequence_id > 58" -ExpectedMin 1
Run-Test "sequence_id < 3" "sequence_id < 3" -ExpectedExact 2
Run-Test "sequence_id >= 59" "sequence_id >= 59" -ExpectedMin 1
Run-Test "sequence_id <= 2" "sequence_id <= 2" -ExpectedExact 2
Write-Host ""

# ========== 2. name (Title field, LIKE) ==========
Write-Host "[2] name filter (title LIKE)" -ForegroundColor Yellow
Run-Test "name LIKE %Docker%" "name LIKE `"%$KW_DOCKER%`"" -ExpectedMin 1
Run-Test "name LIKE %monitor%" "name LIKE `"%$KW_MONITOR%`"" -ExpectedMin 1
Run-Test "name LIKE %asuser%" "name LIKE `"%$KW_AS_USER%`"" -ExpectedMin 1
Run-Test "name LIKE %Dashboard%" "name LIKE `"%$KW_DASHBOARD%`"" -ExpectedMin 1
Run-Test "name LIKE %nonexistent%" "name LIKE `"%nonexistent_xyz%`"" -ExpectedExact 0
Run-Test "name NOT LIKE %Docker%" "name NOT LIKE `"%$KW_DOCKER%`"" -ExpectedMin 1
Run-Test "name = exact title" "name = `"Docker`"" -ExpectedMin 0
Write-Host ""

# ========== 3. state_id (Number field) ==========
Write-Host "[3] state_id filter" -ForegroundColor Yellow
Run-Test "state_id = 85 (backlog)" "state_id = 85" -ExpectedMin 1
Run-Test "state_id = 87 (in progress)" "state_id = 87" -ExpectedMin 1
Run-Test "state_id != 85" "state_id != 85" -ExpectedMin 2
Run-Test "state_id = 999 (not exist)" "state_id = 999" -ExpectedExact 0
Run-Test "state_id IN (85, 86)" "state_id IN (85, 86)" -ExpectedMin 2
Run-Test "state_id IN (85, 86, 87)" "state_id IN (85, 86, 87)" -ExpectedMin 3
Run-Test "state_id NOT IN (85)" "state_id NOT IN (85)" -ExpectedMin 2
Write-Host ""

# ========== 4. state_group (Special field with JOIN) ==========
Write-Host "[4] state_group filter (JOIN states)" -ForegroundColor Yellow
Run-Test "state_group = backlog" "state_group = `"backlog`"" -ExpectedMin 1
Run-Test "state_group = unstarted" "state_group = `"unstarted`"" -ExpectedMin 1
Run-Test "state_group = started" "state_group = `"started`"" -ExpectedMin 1
Run-Test "state_group = completed" "state_group = `"completed`"" -ExpectedMin 1
Run-Test "state_group = cancelled" "state_group = `"cancelled`"" -ExpectedMin 1
Run-Test "state_group = nonexistent" "state_group = `"nonexistent`"" -ExpectedExact 0
Run-Test "state_group != backlog" "state_group != `"backlog`"" -ExpectedMin 1
Run-Test "state_group IN (backlog, started)" "state_group IN (`"backlog`", `"started`")" -ExpectedMin 2
Run-Test "state_group NOT IN (backlog)" "state_group NOT IN (`"backlog`")" -ExpectedMin 1
Run-Test "state_group IN (all 5 groups)" "state_group IN (`"backlog`", `"unstarted`", `"started`", `"completed`", `"cancelled`")" -ExpectedExact 60
Write-Host ""

# ========== 5. priority (String field) ==========
Write-Host "[5] priority filter" -ForegroundColor Yellow
Run-Test "priority = low" "priority = `"low`"" -ExpectedExact 15
Run-Test "priority = medium" "priority = `"medium`"" -ExpectedExact 26
Run-Test "priority = high" "priority = `"high`"" -ExpectedExact 12
Run-Test "priority = urgent" "priority = `"urgent`"" -ExpectedExact 3
Run-Test "priority = none" "priority = `"none`"" -ExpectedExact 4
Run-Test "priority != low" "priority != `"low`"" -ExpectedMin 1
Run-Test "priority = nonexistent" "priority = `"nonexistent`"" -ExpectedExact 0
Run-Test "priority IN (low, high)" "priority IN (`"low`", `"high`")" -ExpectedExact 27
Run-Test "priority IN (urgent, none)" "priority IN (`"urgent`", `"none`")" -ExpectedExact 7
Run-Test "priority NOT IN (low)" "priority NOT IN (`"low`")" -ExpectedExact 45
Write-Host ""

# ========== 6. assignee_id (User field with JOIN) ==========
Write-Host "[6] assignee_id filter (JOIN issue_assignees)" -ForegroundColor Yellow
Run-Test "assignee_id = 49 (admin)" "assignee_id = 49" -ExpectedMin 0
Run-Test "assignee_id = 51 (lisi)" "assignee_id = 51" -ExpectedMin 1
Run-Test "assignee_id = 64" "assignee_id = 64" -ExpectedMin 1
Run-Test "assignee_id = 999 (not exist)" "assignee_id = 999" -ExpectedExact 0
Run-Test "assignee_id IS NULL" "assignee_id IS NULL" -ExpectedMin 0
Run-Test "assignee_id IS NOT NULL" "assignee_id IS NOT NULL" -ExpectedMin 1
Run-Test "assignee_id IN (49, 51)" "assignee_id IN (49, 51)" -ExpectedMin 1
Run-Test "assignee_id NOT IN (49)" "assignee_id NOT IN (49)" -ExpectedMin 0
Write-Host ""

# ========== 7. label (Label field with JOIN) ==========
Write-Host "[7] label filter (JOIN issue_labels)" -ForegroundColor Yellow
Run-Test "label = 163 (frontend)" "label = 163" -ExpectedMin 1
Run-Test "label = 166 (DevOps)" "label = 166" -ExpectedMin 1
Run-Test "label = 999 (not exist)" "label = 999" -ExpectedExact 0
Run-Test "label IS NULL" "label IS NULL" -ExpectedMin 0
Run-Test "label IS NOT NULL" "label IS NOT NULL" -ExpectedMin 1
Run-Test "label IN (163, 166)" "label IN (163, 166)" -ExpectedMin 1
Write-Host ""

# ========== 8. cycle_id (Cycle field with JOIN) ==========
Write-Host "[8] cycle_id filter (JOIN issue_cycles)" -ForegroundColor Yellow
Run-Test "cycle_id = 72 (Sprint 6)" "cycle_id = 72" -ExpectedMin 1
Run-Test "cycle_id = 71 (Sprint 5)" "cycle_id = 71" -ExpectedMin 1
Run-Test "cycle_id = 999 (not exist)" "cycle_id = 999" -ExpectedExact 0
Run-Test "cycle_id IS NULL" "cycle_id IS NULL" -ExpectedMin 0
Run-Test "cycle_id IS NOT NULL" "cycle_id IS NOT NULL" -ExpectedMin 1
Run-Test "cycle_id IN (71, 72)" "cycle_id IN (71, 72)" -ExpectedMin 1
Write-Host ""

# ========== 9. module_id (Module field with JOIN) ==========
Write-Host "[9] module_id filter (JOIN module_issues)" -ForegroundColor Yellow
Run-Test "module_id = 91 (inventory)" "module_id = 91" -ExpectedMin 1
Run-Test "module_id = 94 (member center)" "module_id = 94" -ExpectedMin 1
Run-Test "module_id = 999 (not exist)" "module_id = 999" -ExpectedExact 0
Run-Test "module_id IS NULL" "module_id IS NULL" -ExpectedMin 0
Run-Test "module_id IS NOT NULL" "module_id IS NOT NULL" -ExpectedMin 1
Run-Test "module_id IN (91, 94)" "module_id IN (91, 94)" -ExpectedMin 1
Write-Host ""

# ========== 10. Date fields ==========
Write-Host "[10] date fields filter" -ForegroundColor Yellow
Run-Test "created_at > 2026-01-01" "created_at > `"2026-01-01`"" -ExpectedMin 1
Run-Test "created_at > 2026-12-31" "created_at > `"2026-12-31`"" -ExpectedExact 0
Run-Test "created_at < 2026-12-31" "created_at < `"2026-12-31`"" -ExpectedMin 1
Run-Test "created_at >= 2026-01-01" "created_at >= `"2026-01-01`"" -ExpectedMin 1
Run-Test "created_at <= 2026-12-31" "created_at <= `"2026-12-31`"" -ExpectedMin 1
Run-Test "start_date IS NULL" "start_date IS NULL" -ExpectedMin 0
Run-Test "start_date IS NOT NULL" "start_date IS NOT NULL" -ExpectedMin 0
Run-Test "target_date IS NULL" "target_date IS NULL" -ExpectedMin 0
Run-Test "target_date IS NOT NULL" "target_date IS NOT NULL" -ExpectedMin 0
Write-Host ""

# ========== 11. Quick search (sequence_id) ==========
Write-Host "[11] quick search by issue key" -ForegroundColor Yellow
Run-Test "quick seq=1" "sequence_id = 1" -ExpectedExact 1
Run-Test "quick seq=3" "sequence_id = 3" -ExpectedExact 1
Run-Test "quick seq=5" "sequence_id = 5" -ExpectedExact 1
Run-Test "quick seq=999" "sequence_id = 999" -ExpectedExact 0
Write-Host ""

# ========== 12. Quick search (keyword LIKE OR) ==========
Write-Host "[12] quick search by keyword (LIKE OR)" -ForegroundColor Yellow
Run-Test "quick Docker" "(name LIKE `"%$KW_DOCKER%`" OR description LIKE `"%$KW_DOCKER%`")" -ExpectedMin 1
Run-Test "quick monitor" "(name LIKE `"%$KW_MONITOR%`" OR description LIKE `"%$KW_MONITOR%`")" -ExpectedMin 1
Run-Test "quick Dashboard" "(name LIKE `"%$KW_DASHBOARD%`" OR description LIKE `"%$KW_DASHBOARD%`")" -ExpectedMin 1
Run-Test "quick asuser" "(name LIKE `"%$KW_AS_USER%`" OR description LIKE `"%$KW_AS_USER%`")" -ExpectedMin 1
Run-Test "quick nonexistent" "(name LIKE `"%nonexistent_xyz%`" OR description LIKE `"%nonexistent_xyz%`")" -ExpectedExact 0
Write-Host ""

# ========== 13. Quick search + filter combo ==========
Write-Host "[13] quick search + filter combination" -ForegroundColor Yellow
Run-Test "seq=1 AND priority=low" "sequence_id = 1 AND priority = `"low`"" -ExpectedExact 1
Run-Test "seq=1 AND priority=high (no match)" "sequence_id = 1 AND priority = `"high`"" -ExpectedExact 0
Run-Test "Docker AND priority=low" "(name LIKE `"%$KW_DOCKER%`" OR description LIKE `"%$KW_DOCKER%`") AND priority = `"low`"" -ExpectedMin 1
Run-Test "Docker AND priority=high (no match)" "(name LIKE `"%$KW_DOCKER%`" OR description LIKE `"%$KW_DOCKER%`") AND priority = `"high`"" -ExpectedExact 0
Run-Test "seq=3 AND state_group=backlog" "sequence_id = 3 AND state_group = `"backlog`"" -ExpectedMin 0
Run-Test "seq=2 AND cycle_id=71" "sequence_id = 2 AND cycle_id = 71" -ExpectedExact 1
Write-Host ""

# ========== 14. Complex multi-condition queries ==========
Write-Host "[14] complex multi-condition queries" -ForegroundColor Yellow
Run-Test "priority=low AND state_group=backlog" "priority = `"low`" AND state_group = `"backlog`"" -ExpectedMin 1
Run-Test "priority=urgent AND state_group=started" "priority = `"urgent`" AND state_group = `"started`"" -ExpectedMin 0
Run-Test "assignee_id=51 AND priority=medium" "assignee_id = 51 AND priority = `"medium`"" -ExpectedMin 0
Run-Test "label=163 AND state_group=backlog" "label = 163 AND state_group = `"backlog`"" -ExpectedMin 0
Run-Test "name LIKE asuser AND priority=low" "(name LIKE `"%$KW_AS_USER%`" OR description LIKE `"%$KW_AS_USER%`") AND priority = `"low`"" -ExpectedMin 0
Run-Test "module_id=91 AND priority=low" "module_id = 91 AND priority = `"low`"" -ExpectedMin 0
Run-Test "cycle_id=72 AND state_group=backlog" "cycle_id = 72 AND state_group = `"backlog`"" -ExpectedMin 0
Run-Test "3 conditions: priority + state + assignee" "priority = `"high`" AND state_group = `"started`" AND assignee_id IS NOT NULL" -ExpectedMin 0
Write-Host ""

# ========== 15. IS NULL / IS NOT NULL ==========
Write-Host "[15] IS NULL / IS NOT NULL" -ForegroundColor Yellow
Run-Test "assignee_id IS NULL" "assignee_id IS NULL" -ExpectedMin 0
Run-Test "assignee_id IS NOT NULL" "assignee_id IS NOT NULL" -ExpectedMin 1
Run-Test "label IS NULL" "label IS NULL" -ExpectedMin 0
Run-Test "label IS NOT NULL" "label IS NOT NULL" -ExpectedMin 1
Run-Test "cycle_id IS NULL" "cycle_id IS NULL" -ExpectedMin 0
Run-Test "module_id IS NULL" "module_id IS NULL" -ExpectedMin 0
Run-Test "start_date IS NULL" "start_date IS NULL" -ExpectedMin 0
Run-Test "target_date IS NOT NULL" "target_date IS NOT NULL" -ExpectedMin 0
Write-Host ""

# ========== 16. IN / NOT IN operators ==========
Write-Host "[16] IN / NOT IN operators" -ForegroundColor Yellow
Run-Test "priority IN (low, medium, high)" "priority IN (`"low`", `"medium`", `"high`")" -ExpectedExact 53
Run-Test "priority NOT IN (low)" "priority NOT IN (`"low`")" -ExpectedExact 45
Run-Test "priority NOT IN (all)" "priority NOT IN (`"low`", `"medium`", `"high`", `"urgent`", `"none`")" -ExpectedExact 0
Run-Test "state_id IN (all 6)" "state_id IN (85, 86, 87, 88, 89, 90)" -ExpectedExact 60
Run-Test "state_group IN (all 5)" "state_group IN (`"backlog`", `"unstarted`", `"started`", `"completed`", `"cancelled`")" -ExpectedExact 60
Write-Host ""

# ========== 17. NOT operator ==========
Write-Host "[17] NOT operator" -ForegroundColor Yellow
Run-Test "NOT priority=low" "NOT priority = `"low`"" -ExpectedExact 45
Run-Test "NOT state_group=backlog" "NOT state_group = `"backlog`"" -ExpectedExact 48
Run-Test "NOT assignee_id IS NULL" "NOT assignee_id IS NULL" -ExpectedMin 1
Write-Host ""

# ========== 18. Edge cases ==========
Write-Host "[18] edge cases" -ForegroundColor Yellow
Run-Test "nonexistent field (silently ignored)" "nonexistent_field = 1" -ExpectedMin 1
Run-Test "priority=low AND nonexistent=1" "priority = `"low`" AND nonexistent_field = 1" -ExpectedMin 1
Write-Host ""

# ========== 19. Cross-table JOIN combinations ==========
Write-Host "[19] cross-table JOIN combinations" -ForegroundColor Yellow
Run-Test "assignee + label (two JOINs)" "assignee_id = 51 AND label = 163" -ExpectedMin 0
Run-Test "assignee + cycle + module (three JOINs)" "assignee_id = 51 AND cycle_id = 72 AND module_id = 91" -ExpectedMin 0
Run-Test "state_group + assignee (JOINs)" "state_group = `"backlog`" AND assignee_id IS NOT NULL" -ExpectedMin 0
Run-Test "label + cycle + module + state_group" "label = 163 AND cycle_id = 72 AND module_id = 91 AND state_group = `"backlog`"" -ExpectedMin 0
Write-Host ""

# ========== 20. Ordering ==========
Write-Host "[20] ordering (RQL with orderby)" -ForegroundColor Yellow
Run-Test "orderby priority desc" "priority = `"low`" orderby priority desc" -ExpectedMin 1
Write-Host ""

Write-Host "=== Test Results ===" -ForegroundColor Cyan
Write-Host "PASS: $PASS" -ForegroundColor Green
$failColor = if ($FAIL -gt 0) { 'Red' } else { 'Green' }
Write-Host "FAIL: $FAIL" -ForegroundColor $failColor
Write-Host "SKIP: $SKIP" -ForegroundColor DarkGray
$total = $PASS + $FAIL + $SKIP
Write-Host "TOTAL: $total" -ForegroundColor Cyan
exit $FAIL
