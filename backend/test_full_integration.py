"""
前后端集成测试脚本
测试所有API端点，验证前后端数据交互
"""

import requests
import json
from datetime import datetime

BASE_URL = "http://localhost:8000/api/v1"
results = []

def log_test(test_name, success, message=""):
    """记录测试结果"""
    status = "[PASS]" if success else "[FAIL]"
    results.append({"name": test_name, "success": success, "message": message})
    print(f"{status} {test_name}: {message}")

def make_request(method, endpoint, data=None, token=None, params=None, expected_status=200):
    """发送API请求"""
    url = f"{BASE_URL}{endpoint}"
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"

    try:
        if method == "GET":
            response = requests.get(url, headers=headers, params=params)
        elif method == "POST":
            response = requests.post(url, json=data, headers=headers, params=params)
        elif method == "PUT":
            response = requests.put(url, json=data, headers=headers, params=params)
        elif method == "PATCH":
            response = requests.patch(url, json=data, headers=headers, params=params)
        elif method == "DELETE":
            response = requests.delete(url, headers=headers, params=params)

        if response.status_code == expected_status or (200 <= response.status_code < 300):
            try:
                return response.json()
            except:
                return {"message": "Success"}
        else:
            try:
                error_detail = response.json().get("detail", str(response.status_code))
            except:
                error_detail = str(response.status_code)
            return {"error": error_detail, "status_code": response.status_code}
    except Exception as e:
        return {"error": str(e)}

print("\n" + "="*70)
print("前后端集成测试 - API 端点验证")
print("="*70)

# 1. 用户认证测试
print("\n[1] 用户认证功能测试")
print("-"*50)

test_email = f"integ_{datetime.now().strftime('%Y%m%d%H%M%S')}@test.com"
test_username = f"integuser_{datetime.now().strftime('%Y%m%d%H%M%S')}"

# 注册
result = make_request("POST", "/auth/register", {
    "email": test_email,
    "username": test_username,
    "password": "TestPass123!",
    "display_name": "Integration Test User"
}, expected_status=201)
if "error" not in result and "id" in result:
    log_test("用户注册", True, f"ID: {result.get('id', '')[:8]}...")
else:
    log_test("用户注册", False, result.get("error", "Unknown error"))

# 登录
result = make_request("POST", "/auth/login", {
    "email": test_email,
    "password": "TestPass123!"
})
if "error" not in result and "access_token" in result:
    log_test("用户登录", True, "Token获取成功")
    access_token = result.get("access_token")
else:
    log_test("用户登录", False, result.get("error", "Unknown error"))
    access_token = None

# 获取当前用户
if access_token:
    result = make_request("GET", "/auth/me", token=access_token)
    if "error" not in result:
        log_test("获取用户信息", True, f"用户名: {result.get('username')}")
    else:
        log_test("获取用户信息", False, result.get("error"))

# 2. 工作空间测试
print("\n[2] 工作空间功能测试")
print("-"*50)

if access_token:
    workspace_slug = f"test-ws-{datetime.now().strftime('%H%M%S')}"
    result = make_request("POST", "/workspaces/", {
        "name": "Integration Test Workspace",
        "slug": workspace_slug,
        "timezone": "Asia/Shanghai"
    }, access_token)
    if "error" not in result:
        log_test("创建工作空间", True, f"ID: {result.get('id', '')[:8]}...")
        workspace_id = result.get("id")
    else:
        log_test("创建工作空间", False, result.get("error"))
        workspace_id = None

    # 获取工作空间详情
    if workspace_id:
        result = make_request("GET", f"/workspaces/{workspace_slug}", token=access_token)
        if "error" not in result:
            log_test("获取工作空间", True, f"名称: {result.get('name')}")
        else:
            log_test("获取工作空间", False, result.get("error"))
else:
    workspace_id = None

# 3. 项目测试
print("\n[3] 项目功能测试")
print("-"*50)

if workspace_id:
    result = make_request("POST", "/projects/", {
        "name": "Integration Test Project",
        "identifier": "ITP",
        "description": "Test project for integration",
        "is_public": True,
        "timezone": "Asia/Shanghai"
    }, access_token, params={"workspace_id": workspace_id})
    if "error" not in result:
        log_test("创建项目", True, f"ID: {result.get('id', '')[:8]}...")
        project_id = result.get("id")
    else:
        log_test("创建项目", False, result.get("error"))
        project_id = None

    # 获取项目列表
    if workspace_id:
        result = make_request("GET", "/projects/", token=access_token, params={"workspace_id": workspace_id})
        if "error" not in result and isinstance(result, list):
            log_test("获取项目列表", True, f"共 {len(result)} 个项目")
        else:
            log_test("获取项目列表", False, result.get("error"))

    # 获取项目详情
    if project_id:
        result = make_request("GET", f"/projects/{project_id}", token=access_token)
        if "error" not in result:
            log_test("获取项目详情", True, f"名称: {result.get('name')}")
        else:
            log_test("获取项目详情", False, result.get("error"))
else:
    project_id = None

# 4. 工作项测试
print("\n[4] 工作项功能测试")
print("-"*50)

if project_id and workspace_id:
    result = make_request("POST", "/issues/", {
        "name": "Integration Test Issue",
        "description_html": "<p>Test issue description</p>",
        "description_json": {"type": "doc", "content": []},
        "priority": "high"
    }, access_token, params={"workspace_id": workspace_id, "project_id": project_id})
    if "error" not in result:
        log_test("创建工作项", True, f"ID: {result.get('id', '')[:8]}...")
        issue_id = result.get("id")
    else:
        log_test("创建工作项", False, result.get("error"))
        issue_id = None

    # 获取工作项列表
    result = make_request("GET", "/issues/", token=access_token, params={
        "workspace_id": workspace_id,
        "project_id": project_id
    })
    if "error" not in result and isinstance(result, list):
        log_test("获取工作项列表", True, f"共 {len(result)} 个工作项")
    else:
        log_test("获取工作项列表", False, result.get("error"))

    # 获取工作项详情
    if issue_id:
        result = make_request("GET", f"/issues/{issue_id}", token=access_token)
        if "error" not in result:
            log_test("获取工作项详情", True, f"名称: {result.get('name')}")
        else:
            log_test("获取工作项详情", False, result.get("error"))
else:
    issue_id = None

# 5. 状态管理测试
print("\n[5] 状态管理功能测试")
print("-"*50)

if project_id and workspace_id:
    # 创建新状态
    result = make_request("POST", f"/projects/{project_id}/settings/states/", {
        "name": "In Testing",
        "color": "#8B5CF6",
        "group": "in_progress",
        "sequence": 3,
        "project_id": project_id
    }, access_token, params={"workspace_id": workspace_id})
    if "error" not in result:
        log_test("创建状态", True, f"ID: {result.get('id', '')[:8]}...")
        state_id = result.get("id")
    else:
        log_test("创建状态", False, result.get("error"))
        state_id = None

    # 获取状态列表
    result = make_request("GET", f"/projects/{project_id}/settings/states/", token=access_token)
    if "error" not in result and isinstance(result, list):
        log_test("获取状态列表", True, f"共 {len(result)} 个状态")
    else:
        log_test("获取状态列表", False, result.get("error"))
else:
    state_id = None

# 6. 自动化规则测试
print("\n[6] 自动化规则功能测试")
print("-"*50)

if project_id and workspace_id:
    result = make_request("POST", f"/projects/{project_id}/workflow/automations/", {
        "name": "Auto Assign Bug",
        "is_enabled": True,
        "trigger": {"type": "issue.created"},
        "conditions": [{"field": "type", "operator": "equals", "value": "bug"}],
        "actions": [{"type": "notification.create", "config": {}}],
        "project_id": project_id
    }, access_token, params={"workspace_id": workspace_id})
    if "error" not in result:
        log_test("创建自动化规则", True, f"ID: {result.get('id', '')[:8]}...")
    else:
        log_test("创建自动化规则", False, result.get("error"))

    # 获取自动化列表
    result = make_request("GET", f"/projects/{project_id}/workflow/automations/", token=access_token)
    if "error" not in result and isinstance(result, list):
        log_test("获取自动化列表", True, f"共 {len(result)} 条规则")
    else:
        log_test("获取自动化列表", False, result.get("error"))

# 7. 通知测试
print("\n[7] 通知功能测试")
print("-"*50)

if access_token:
    result = make_request("GET", "/notifications/", token=access_token)
    if "error" not in result and isinstance(result, list):
        log_test("获取通知列表", True, f"共 {len(result)} 条通知")
    else:
        log_test("获取通知列表", False, result.get("error"))

    # 获取通知摘要
    result = make_request("GET", "/notifications/summary", token=access_token)
    if "error" not in result:
        log_test("获取通知摘要", True, f"未读: {result.get('unread_count', 0)}")
    else:
        log_test("获取通知摘要", False, result.get("error"))

# 8. 评论测试
print("\n[8] 评论功能测试")
print("-"*50)

if issue_id:
    result = make_request("POST", "/comments/", {
        "issue_id": issue_id,
        "content": "This is a test comment",
        "html_content": "<p>This is a test comment</p>"
    }, access_token)
    if "error" not in result:
        log_test("创建评论", True, f"ID: {result.get('id', '')[:8]}...")
        comment_id = result.get("id")
    else:
        log_test("创建评论", False, result.get("error"))
        comment_id = None

    # 获取评论列表
    if issue_id:
        result = make_request("GET", f"/comments/issue/{issue_id}", token=access_token)
        if "error" not in result:
            log_test("获取评论列表", True, f"共 {result.get('total', 0)} 条评论")
        else:
            log_test("获取评论列表", False, result.get("error"))
else:
    comment_id = None

# 9. 自定义字段测试
print("\n[9] 自定义字段功能测试")
print("-"*50)

if project_id and workspace_id:
    result = make_request("POST", "/custom-fields/", {
        "name": "Story Points",
        "field_type": "number",
        "is_required": False,
        "is_active": True
    }, access_token, params={"workspace_id": workspace_id, "project_id": project_id})
    if "error" not in result:
        log_test("创建自定义字段", True, f"ID: {result.get('id', '')[:8]}...")
    else:
        log_test("创建自定义字段", False, result.get("error"))

    # 获取自定义字段列表
    result = make_request("GET", "/custom-fields/", token=access_token, params={
        "workspace_id": workspace_id,
        "project_id": project_id
    })
    if "error" not in result and isinstance(result, list):
        log_test("获取自定义字段列表", True, f"共 {len(result)} 个字段")
    else:
        log_test("获取自定义字段列表", False, result.get("error"))

# 测试结果汇总
print("\n" + "="*70)
print("测试结果汇总")
print("="*70)

passed = sum(1 for r in results if r["success"])
failed = len(results) - passed
total = len(results)

print(f"总测试数: {total}")
print(f"通过: {passed}")
print(f"失败: {failed}")
print(f"通过率: {passed/total*100:.1f}%" if total > 0 else "N/A")

if failed > 0:
    print("\n失败测试:")
    for r in results:
        if not r["success"]:
            print(f"  - {r['name']}: {r['message']}")

print("\n" + "="*70)
print("前后端集成测试完成!")
print("="*70)
