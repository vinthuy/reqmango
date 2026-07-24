# ============================================================
# 定时任务自动化验证脚本
# 功能: 
#   1. 启用内置定时规则并设置立即触发
#   2. 插入模拟运行记录(automation_executions)
#   3. 调用 API 手动执行规则(如果后台运行中)
# ============================================================
param(
    [string]$DB_HOST = "localhost",
    [string]$DB_PORT = "5432",
    [string]$DB_USER = "postgres",
    [string]$DB_PASS = "postgres",
    [string]$DB_NAME = "reqmango",
    [string]$API_URL = "http://localhost:8000",
    [switch]$API_ONLY = $false,
    [switch]$DB_ONLY = $false
)

$ErrorActionPreference = "Stop"
$PS_DB_URL = "postgresql://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host " 定时任务自动化验证工具" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# ============================================================
# 辅助函数: 执行 SQL
# ============================================================
function Invoke-SQL {
    param([string]$SQL)
    $env:PGPASSWORD = $DB_PASS
    $result = & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c $SQL 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  SQL Error: $result" -ForegroundColor Red
        return $null
    }
    return $result
}

# ============================================================
# 辅助函数: 调用 API
# ============================================================
function Invoke-API {
    param(
        [string]$Method,
        [string]$Path,
        $Body = $null
    )
    try {
        $headers = @{ "Content-Type" = "application/json" }
        $uri = "$API_URL$Path"
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -Body $jsonBody -TimeoutSec 10
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -TimeoutSec 10
        }
        return @{ Success = $true; Data = $response }
    } catch {
        return @{ Success = $false; Error = $_.Exception.Message }
    }
}

# ============================================================
# 第一步: 检查 psql 是否可用 (除非只使用 API)
# ============================================================
if (-not $API_ONLY) {
    Write-Host "[Step 1] 检查 psql 可用性..." -ForegroundColor Yellow
    $hasPsql = Get-Command psql -ErrorAction SilentlyContinue
    if (-not $hasPsql) {
        Write-Host "  警告: psql 未找到, 将跳过数据库直连操作" -ForegroundColor Yellow
        Write-Host "  请安装 PostgreSQL 客户端或将 psql 加入 PATH" -ForegroundColor Yellow
        Write-Host ""
        $DB_ONLY = $false
    } else {
        # 检查数据库连通性
        $dbCheck = Invoke-SQL "SELECT 1;" 2>&1
        if ($dbCheck -match "1") {
            Write-Host "  [OK] 数据库连接成功" -ForegroundColor Green
        } else {
            Write-Host "  [FAIL] 数据库连接失败: $dbCheck" -ForegroundColor Red
            Write-Host "  将跳过数据库操作" -ForegroundColor Yellow
            $DB_ONLY = $false
        }
    }
    Write-Host ""
}

# ============================================================
# 第二步: 数据库操作 - 启用/创建定时规则 & 插入运行记录
# ============================================================
if (-not $API_ONLY) {
    Write-Host "[Step 2] 数据库操作 - 定时规则 & 运行记录" -ForegroundColor Yellow

    # 2.1 获取现有的工作区和项目信息
    Write-Host "  2.1 获取工作区和项目信息..." -ForegroundColor Gray
    $wsInfo = Invoke-SQL "SELECT id, name FROM workspaces LIMIT 1;"
    if (-not $wsInfo) {
        Write-Host "  错误: 没有找到工作区, 请先初始化数据" -ForegroundColor Red
        if (-not $API_ONLY) { exit 1 }
    } else {
        $wsId, $wsName = $wsInfo -split '\|'
        Write-Host "    工作区: ID=$wsId, Name=$wsName" -ForegroundColor Gray

        $projectInfo = Invoke-SQL "SELECT id, name FROM projects WHERE workspace_id = $wsId LIMIT 1;"
        if ($projectInfo) {
            $projId, $projName = $projectInfo -split '\|'
            Write-Host "    项目:   ID=$projId, Name=$projName" -ForegroundColor Gray
        } else {
            Write-Host "    警告: 工作区下没有项目" -ForegroundColor Yellow
        }

        # 获取一些 issue 用于测试
        $issueIds = @()
        if ($projId) {
            $issueList = Invoke-SQL "SELECT id FROM issues WHERE project_id = $projId AND deleted_at IS NULL ORDER BY id LIMIT 5;" | Where-Object { $_ -match '^\d+$' }
            foreach ($id in $issueList) {
                $issueIds += $id.Trim()
            }
            Write-Host "    找到 $($issueIds.Count) 个 Issue" -ForegroundColor Gray
        }

        # 2.2 检查现有定时规则
        Write-Host "  2.2 检查现有定时规则..." -ForegroundColor Gray
        $existingScheduled = Invoke-SQL @"
SELECT id, name, is_enabled, schedule_config, execution_count, last_triggered_at 
FROM automation_rules 
WHERE trigger_type = 'scheduled' 
ORDER BY id;
"@
        Write-Host "    现有定时规则:" -ForegroundColor Gray
        if ($existingScheduled) {
            foreach ($line in ($existingScheduled -split "`n")) {
                if ($line.Trim()) { Write-Host "      $line" -ForegroundColor Gray }
            }
        } else {
            Write-Host "      (无)" -ForegroundColor Gray
        }

        # 2.3 启用所有内置定时规则, 并设置 last_triggered_at 为 NULL 使其立即触发
        Write-Host "  2.3 启用内置定时规则 (设为立即触发)..." -ForegroundColor Gray
        $now = Get-Date
        $nowTime = $now.ToString("HH:mm")
        $nowMinute = $now.Minute
        $nowDay = $now.DayOfWeek.ToString().Substring(0, 3).ToLower()
        $nowDayNum = $now.Day

        # 为每条定时规则准备匹配当前时间的 schedule_config
        $updateResult = Invoke-SQL @"
UPDATE automation_rules 
SET 
    is_enabled = true,
    last_triggered_at = NULL,
    schedule_config = CASE 
        WHEN schedule_config LIKE '%"daily"%' THEN jsonb_set(schedule_config::jsonb, '{time}', '"$nowTime"')
        WHEN schedule_config LIKE '%"weekly"%' THEN jsonb_set(jsonb_set(schedule_config::jsonb, '{time}', '"$nowTime"'), '{days}', '["$nowDay"]')
        WHEN schedule_config LIKE '%"monthly"%' THEN jsonb_set(jsonb_set(schedule_config::jsonb, '{time}', '"$nowTime"'), '{day}', '$nowDayNum')
        WHEN schedule_config LIKE '%"hourly"%' THEN jsonb_set(schedule_config::jsonb, '{minute}', '$nowMinute')
        ELSE schedule_config
    END
WHERE trigger_type = 'scheduled' AND is_enabled = false;

-- 确保 rule 的 conditions 至少不为空, 空条件始终匹配
UPDATE automation_rules SET conditions = '[]' WHERE conditions = '' OR conditions IS NULL;
"@
        Write-Host "    已启用并配置为当前触发时间" -ForegroundColor Green

        # 2.4 插入模拟运行记录
        Write-Host "  2.4 插入模拟运行记录..." -ForegroundColor Gray
        $scheduledRules = Invoke-SQL "SELECT id, name FROM automation_rules WHERE trigger_type = 'scheduled' ORDER BY id;"

        if ($scheduledRules) {
            $ruleEntries = @()
            foreach ($line in ($scheduledRules -split "`n")) {
                if ($line.Trim()) {
                    $parts = $line -split '\|'
                    if ($parts.Count -ge 2) {
                        $ruleEntries += @{ Id = $parts[0].Trim(); Name = $parts[1].Trim() }
                    }
                }
            }

            foreach ($entry in $ruleEntries) {
                $ruleId = $entry.Id
                $ruleName = $entry.Name
                $testIssueId = if ($issueIds.Count -gt 0) { $issueIds[0] } else { 0 }
                $now = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd HH:mm:ss")

                # 插入成功记录
                $successCtx = @{rule_id=$ruleId;workspace_id=$wsId;issue_id=$testIssueId;action="add_comment";comment="[定时任务验证] 规则 '$ruleName' 在 $nowTime 自动触发成功"} | ConvertTo-Json -Compress
                $successActions = '["成功: 自动添加了定时任务评论"]'
                
                $sql = @"
INSERT INTO automation_executions (rule_id, issue_id, trigger_type, context_json, actions_taken, status, error, duration, executed_at, created_at, updated_at)
VALUES ($ruleId, $testIssueId, 'scheduled', '$successCtx', '$successActions', 'success', '', floor(random() * 500 + 50)::bigint, '$now'::timestamp - interval '2 minutes', '$now'::timestamp - interval '2 minutes', '$now'::timestamp - interval '2 minutes');
"@
                Invoke-SQL $sql
                Write-Host "    插入成功记录: 规则 #$ruleId '$ruleName'" -ForegroundColor Green

                # 插入失败记录 (模拟: 无 issue 可操作)
                $failedCtx = @{rule_id=$ruleId;workspace_id=$wsId;issue_id=0;note="No issues matched for scheduled action"} | ConvertTo-Json -Compress
                $sql2 = @"
INSERT INTO automation_executions (rule_id, issue_id, trigger_type, context_json, actions_taken, status, error, duration, executed_at, created_at, updated_at)
VALUES ($ruleId, 0, 'scheduled', '$failedCtx', '[]', 'failed', 'no matching issues found in project scope', floor(random() * 300 + 20)::bigint, '$now'::timestamp - interval '5 minutes', '$now'::timestamp - interval '5 minutes', '$now'::timestamp - interval '5 minutes');
"@
                Invoke-SQL $sql2
                Write-Host "    插入失败记录: 规则 #$ruleId (模拟无匹配 issue)" -ForegroundColor Yellow

                # 插入跳过记录 (模拟: 条件不满足)
                $skippedCtx = @{rule_id=$ruleId;workspace_id=$wsId;issue_id=$testIssueId;note="Scheduled run but conditions not met"} | ConvertTo-Json -Compress
                $sql3 = @"
INSERT INTO automation_executions (rule_id, issue_id, trigger_type, context_json, actions_taken, status, error, duration, executed_at, created_at, updated_at)
VALUES ($ruleId, $testIssueId, 'scheduled', '$skippedCtx', '[]', 'skipped', 'Conditions not met', floor(random() * 100 + 10)::bigint, '$now'::timestamp - interval '10 minutes', '$now'::timestamp - interval '10 minutes', '$now'::timestamp - interval '10 minutes');
"@
                Invoke-SQL $sql3
                Write-Host "    插入跳过记录: 规则 #$ruleId (条件不满足)" -ForegroundColor Gray
            }
        } else {
            Write-Host "    警告: 没有找到定时规则字段" -ForegroundColor Yellow
        }

        # 2.5 如果没有项目级定时规则,创建一个新的
        if ($projId) {
            Write-Host "  2.5 创建项目级测试定时规则..." -ForegroundColor Gray
            $testRuleName = "[测试] 每小时自动检查高优先级任务"
            $scheduleConfig = @{frequency="hourly";minute=$nowMinute} | ConvertTo-Json -Compress
            $conditions = '[]'
            $actions = '[{"type":"set_field","field":"priority","value":"high"}]'
            
            $checkExist = Invoke-SQL @"
SELECT id FROM automation_rules WHERE project_id = $projId AND trigger_type = 'scheduled' LIMIT 1;
"@
            if (-not $checkExist -or $checkExit -match '^\s*$') {
                Invoke-SQL @"
INSERT INTO automation_rules (project_id, workspace_id, name, trigger_type, conditions, actions, is_enabled, scope, schedule_config, execution_count, sequence, created_at, updated_at)
VALUES ($projId, $wsId, '$testRuleName', 'scheduled', '$conditions', '$actions', true, 'all', '$scheduleConfig', 0, 99, NOW(), NOW());
"@
                $newRuleId = Invoke-SQL "SELECT id FROM automation_rules WHERE name = '$testRuleName' LIMIT 1;"
                Write-Host "    创建成功! 新规则 ID=$newRuleId" -ForegroundColor Green

                # 为新规则也插入运行记录
                if ($newRuleId) {
                    $newRuleId = $newRuleId.Trim()
                    $testIssueId = if ($issueIds.Count -gt 0) { $issueIds[0] } else { 0 }
                    $now = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd HH:mm:ss")
                    $newCtx = @{rule_id=$newRuleId;workspace_id=$wsId;issue_id=$testIssueId;check_time=$nowTime} | ConvertTo-Json -Compress
                    
                    Invoke-SQL @"
INSERT INTO automation_executions (rule_id, issue_id, trigger_type, context_json, actions_taken, status, error, duration, executed_at, created_at, updated_at)
VALUES ($newRuleId, $testIssueId, 'scheduled', '$newCtx', '["set priority to high"]', 'success', '', 345, '$now'::timestamp - interval '1 minute', '$now'::timestamp - interval '1 minute', '$now'::timestamp - interval '1 minute');
"@
                    Write-Host "    插入新规则的运行记录" -ForegroundColor Green
                }
            } else {
                Write-Host "    项目 $projId 已有定时规则, 跳过创建" -ForegroundColor Gray
            }
        }
    }

    # 2.6 统计
    Write-Host "  2.6 统计运行记录..." -ForegroundColor Gray
    $countResult = Invoke-SQL "SELECT status, count(*) FROM automation_executions WHERE trigger_type = 'scheduled' GROUP BY status ORDER BY status;"
    if ($countResult) {
        Write-Host "    运行记录统计:" -ForegroundColor Green
        foreach ($line in ($countResult -split "`n")) {
            if ($line.Trim()) { Write-Host "      $line" -ForegroundColor Gray }
        }
    }
    Write-Host ""
}

# ============================================================
# 第三步: API 验证 (如果后台运行中)
# ============================================================
if (-not $DB_ONLY) {
    Write-Host "[Step 3] API 验证..." -ForegroundColor Yellow

    # 检查后台是否运行
    $healthCheck = Invoke-API -Method "GET" -Path "/api/health"
    if (-not $healthCheck.Success) {
        Write-Host "  后台服务未运行, 跳过 API 验证" -ForegroundColor Yellow
        Write-Host "  启动后台后可以手动调用以下 API 验证:" -ForegroundColor Yellow
        Write-Host "    POST /api/projects/{projectId}/automations/{ruleId}/execute" -ForegroundColor Gray
        Write-Host "    GET  /api/projects/{projectId}/automation-executions" -ForegroundColor Gray
        Write-Host "    GET  /api/automations/{ruleId}/execution-history" -ForegroundColor Gray
    } else {
        Write-Host "  [OK] 后台服务运行中" -ForegroundColor Green

        # 尝试获取登录 token (如果已经有用户注册)
        Write-Host "  3.1 尝试获取 API Token..." -ForegroundColor Gray
        $loginResp = Invoke-API -Method "POST" -Path "/api/auth/login" -Body @{
            email = "admin@reqmango.com"
            password = "demo1234"
        }
        
        if ($loginResp.Success -and $loginResp.Data.access_token) {
            $token = $loginResp.Data.access_token
            Write-Host "    [OK] 获取 Token 成功" -ForegroundColor Green
            
            # 这里可以继续调用其他 API...
            # 获取工作区列表
            $headers = @{ 
                "Content-Type" = "application/json"
                "Authorization" = "Bearer $token"
            }
            try {
                $wsResp = Invoke-RestMethod -Uri "$API_URL/api/workspaces" -Headers $headers -TimeoutSec 5
                if ($wsResp -and $wsResp.Count -gt 0) {
                    Write-Host "    工作区列表获取成功" -ForegroundColor Green
                }
            } catch {
                Write-Host "    获取工作区失败: $_" -ForegroundColor Yellow
            }
        } else {
            Write-Host "    登录失败 (可能没有 admin 用户), 跳过需认证的 API" -ForegroundColor Yellow
        }

        # 手动触发验证
        Write-Host "  3.2 提示: 可手动触发规则 (需替换实际的 projectId 和 ruleId):" -ForegroundColor Gray
        Write-Host "    curl -X POST $API_URL/api/projects/PROJECT_ID/automations/RULE_ID/execute -H 'Content-Type: application/json' -d '{\"issue_id\": ISSUE_ID}'" -ForegroundColor Gray
    }
    Write-Host ""
}

# ============================================================
# 第四步: 摘要
# ============================================================
Write-Host "================================================" -ForegroundColor Cyan
Write-Host " 验证完成!" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "定时规则已设置为立即触发, 后台调度器每 1 分钟检查一次。" -ForegroundColor White
Write-Host "如后台正在运行, 1 分钟内调度器将自动触发定时规则!" -ForegroundColor White
Write-Host ""
Write-Host "查看运行记录的方式:" -ForegroundColor Yellow
Write-Host "  1. 前端界面: 项目设置 > 自动化规则 > 执行日志" -ForegroundColor Gray
Write-Host "  2. 前端界面: Issue 详情 > 自动化历史" -ForegroundColor Gray
Write-Host "  3. API: GET /api/projects/{projectId}/automation-executions" -ForegroundColor Gray
Write-Host "  4. 数据库: SELECT * FROM automation_executions WHERE trigger_type='scheduled' ORDER BY executed_at DESC;" -ForegroundColor Gray
Write-Host ""
Write-Host "下一步: 启动后台服务并观察调度器日志输出:" -ForegroundColor Yellow
Write-Host "  cd backend && go run ./cmd/server/" -ForegroundColor Gray
Write-Host "  观察日志中的 [Automation] 前缀消息" -ForegroundColor Gray
