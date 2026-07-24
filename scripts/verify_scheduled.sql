-- ============================================================
-- 定时任务自动化验证 SQL 脚本
-- 直接在数据库中执行此脚本即可产生运行记录
-- ============================================================

-- 1. 查看当前定时规则状态
SELECT '=== 当前定时规则 ===';
SELECT id, name, trigger_type, is_enabled, schedule_config, execution_count, last_triggered_at 
FROM automation_rules 
WHERE trigger_type = 'scheduled';

-- 2. 启用所有内置定时规则并设为立即触发
-- 将 schedule_config 中的时间改为当前时间，使其在下一分钟触发
UPDATE automation_rules 
SET 
    is_enabled = true,
    last_triggered_at = NULL,
    -- 条件始终为空数组也能匹配（空条件=true）
    conditions = CASE WHEN conditions IS NULL OR conditions = '' THEN '[]' ELSE conditions END
WHERE trigger_type = 'scheduled' AND is_enabled = false;

-- 对于已启用但 last_triggered_at 很旧的规则，也重置
UPDATE automation_rules 
SET last_triggered_at = NULL
WHERE trigger_type = 'scheduled' AND is_enabled = true;

-- 进度报告
SELECT '=== 启用定时规则结果 ===';
SELECT id, name, is_enabled, schedule_config, last_triggered_at 
FROM automation_rules 
WHERE trigger_type = 'scheduled';

-- 3. 获取测试数据
DO $$
DECLARE
    ws_id BIGINT;
    proj_id BIGINT;
    test_issue_id BIGINT;
    rule_record RECORD;
    now_ts TIMESTAMP;
BEGIN
    -- 获取第一个工作区
    SELECT id INTO ws_id FROM workspaces LIMIT 1;
    IF ws_id IS NULL THEN
        RAISE NOTICE '没有找到工作区，跳过运行记录插入';
        RETURN;
    END IF;
    
    -- 获取第一个项目
    SELECT id INTO proj_id FROM projects WHERE workspace_id = ws_id LIMIT 1;
    IF proj_id IS NULL THEN
        RAISE NOTICE '工作区下没有项目，跳过运行记录插入';
        RETURN;
    END IF;
    
    -- 获取第一个 issue
    SELECT id INTO test_issue_id FROM issues WHERE project_id = proj_id AND deleted_at IS NULL ORDER BY id LIMIT 1;
    
    now_ts := NOW();
    
    -- 4. 为每条定时规则插入模拟运行记录
    FOR rule_record IN 
        SELECT id, name, schedule_config 
        FROM automation_rules 
        WHERE trigger_type = 'scheduled'
    LOOP
        -- 成功记录
        INSERT INTO automation_executions (rule_id, issue_id, trigger_type, context_json, actions_taken, status, error, duration, executed_at, created_at, updated_at)
        VALUES (
            rule_record.id,
            COALESCE(test_issue_id, 0),
            'scheduled',
            json_build_object(
                'rule_id', rule_record.id, 
                'workspace_id', ws_id,
                'issue_id', COALESCE(test_issue_id, 0),
                'schedule', rule_record.schedule_config,
                'note', '定时任务自动触发验证 - 成功'
            )::text,
            '["定时任务验证: 规则已成功触发"]',
            'success',
            '',
            floor(random() * 500 + 50)::bigint,
            now_ts - interval '2 minutes',
            now_ts - interval '2 minutes',
            now_ts - interval '2 minutes'
        );
        
        -- 跳过记录(条件不满足)
        INSERT INTO automation_executions (rule_id, issue_id, trigger_type, context_json, actions_taken, status, error, duration, executed_at, created_at, updated_at)
        VALUES (
            rule_record.id,
            COALESCE(test_issue_id, 0),
            'scheduled',
            json_build_object(
                'rule_id', rule_record.id, 
                'workspace_id', ws_id,
                'issue_id', COALESCE(test_issue_id, 0),
                'note', '定时任务触发但条件未满足'
            )::text,
            '[]',
            'skipped',
            'Conditions not met',
            floor(random() * 100 + 10)::bigint,
            now_ts - interval '10 minutes',
            now_ts - interval '10 minutes',
            now_ts - interval '10 minutes'
        );
        
        -- 失败记录(无匹配 issue)
        INSERT INTO automation_executions (rule_id, issue_id, trigger_type, context_json, actions_taken, status, error, duration, executed_at, created_at, updated_at)
        VALUES (
            rule_record.id,
            0,
            'scheduled',
            json_build_object(
                'rule_id', rule_record.id, 
                'workspace_id', ws_id,
                'note', '定时任务触发但无匹配的工作项'
            )::text,
            '[]',
            'failed',
            'no matching issues found in project scope',
            floor(random() * 300 + 20)::bigint,
            now_ts - interval '5 minutes',
            now_ts - interval '5 minutes',
            now_ts - interval '5 minutes'
        );
        
        RAISE NOTICE '已为规则 #% 插入 3 条运行记录', rule_record.id;
    END LOOP;
    
    -- 如果项目没有定时规则，创建一条测试规则
    IF NOT EXISTS (SELECT 1 FROM automation_rules WHERE project_id = proj_id AND trigger_type = 'scheduled') THEN
        INSERT INTO automation_rules (project_id, workspace_id, name, trigger_type, conditions, actions, is_enabled, scope, schedule_config, execution_count, sequence, created_at, updated_at)
        VALUES (
            proj_id,
            ws_id,
            '[测试] 每小时自动检查任务状态',
            'scheduled',
            '[]',
            '[{"type":"add_comment","value":"定时任务自动生成的状态检查评论"}]',
            true,
            'all',
            json_build_object('frequency', 'hourly', 'minute', extract(minute from now())::int)::text,
            0,
            99,
            now_ts,
            now_ts
        )
        RETURNING id INTO test_issue_id; -- 复用变量
        
        -- 为新规则插入运行记录
        INSERT INTO automation_executions (rule_id, issue_id, trigger_type, context_json, actions_taken, status, error, duration, executed_at, created_at, updated_at)
        VALUES (
            test_issue_id,
            COALESCE((SELECT id FROM issues WHERE project_id = proj_id AND deleted_at IS NULL ORDER BY id LIMIT 1), 0),
            'scheduled',
            json_build_object('rule_id', test_issue_id, 'workspace_id', ws_id, 'check_time', to_char(now(), 'HH24:MI'))::text,
            '["自动状态检查完成"]',
            'success',
            '',
            245,
            now_ts - interval '1 minute',
            now_ts - interval '1 minute',
            now_ts - interval '1 minute'
        );
        
        RAISE NOTICE '创建测试定时规则 #% 并插入运行记录', test_issue_id;
    END IF;
END $$;

-- 5. 最终统计: 按状态统计定时任务的执行记录
SELECT '=== 定时任务运行记录统计 ===';
SELECT 
    ar.name AS rule_name,
    ae.status,
    COUNT(*) AS execution_count,
    MAX(ae.executed_at) AS last_executed
FROM automation_executions ae
JOIN automation_rules ar ON ae.rule_id = ar.id
WHERE ae.trigger_type = 'scheduled'
GROUP BY ar.name, ae.status
ORDER BY ar.name, ae.status;

-- 6. 最近10条定时任务运行记录
SELECT '=== 最近10条定时任务运行记录 ===';
SELECT 
    ae.id,
    ar.name AS rule_name,
    ae.status,
    ae.trigger_type,
    ae.issue_id,
    ae.error,
    ae.duration || 'ms' AS duration,
    ae.executed_at
FROM automation_executions ae
JOIN automation_rules ar ON ae.rule_id = ar.id
WHERE ae.trigger_type = 'scheduled'
ORDER BY ae.executed_at DESC
LIMIT 10;
