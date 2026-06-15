"""
完整的API集成测试脚本
测试所有主要功能：用户认证、工作空间、项目、工作项、自定义字段、工作流等
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
        
        if response.status_code == expected_status or (response.status_code >= 200 and response.status_code < 300):
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

# ========================================
# 1. 用户认证功能测试
# ========================================
print("\n" + "="*60)
print("1. 用户认证功能测试")
print("="*60)

# 注册新用户
test_email = f"test_{datetime.now().strftime('%Y%m%d%H%M%S')}@example.com"
test_username = f"testuser_{datetime.now().strftime('%Y%m%d%H%M%S')}"
register_data = {
    "email": test_email,
    "username": test_username,
    "password": "testpass123",
    "display_name": "Test User"
}
result = make_request("POST", "/auth/register", register_data, expected_status=201)
if "error" not in result:
    log_test("用户注册", True, f"用户ID: {result.get('id', '')[:8]}...")
    user_id = result.get('id')
else:
    log_test("用户注册", False, result.get('error'))

# 登录
login_data = {"email": test_email, "password": "testpass123"}
result = make_request("POST", "/auth/login", login_data)
if "error" not in result and "access_token" in result:
    log_test("用户登录", True, "获取Token成功")
    access_token = result.get('access_token')
else:
    log_test("用户登录", False, result.get('error', 'Unknown error'))
    access_token = None

# 获取当前用户信息
if access_token:
    result = make_request("GET", "/auth/me", token=access_token)
    if "error" not in result:
        log_test("获取用户信息", True, f"用户名: {result.get('username')}")
    else:
        log_test("获取用户信息", False, result.get('error'))

# ========================================
# 2. 工作空间功能测试
# ========================================
print("\n" + "="*60)
print("2. 工作空间功能测试")
print("="*60)

workspace_slug = f"test-workspace-{datetime.now().strftime('%Y%m%d%H%M%S')}"
workspace_data = {
    "name": "Test Workspace",
    "slug": workspace_slug,
    "timezone": "Asia/Shanghai"
}
result = make_request("POST", "/workspaces/", workspace_data, access_token)
if "error" not in result:
    log_test("创建工作空间", True, f"工作空间ID: {result.get('id', '')[:8]}...")
    workspace_id = result.get('id')
else:
    log_test("创建工作空间", False, result.get('error'))
    workspace_id = None

# 获取单个工作空间（通过slug）
if workspace_slug:
    result = make_request("GET", f"/workspaces/{workspace_slug}", token=access_token)
    if "error" not in result:
        log_test("获取单个工作空间", True, f"名称: {result.get('name')}")
    else:
        log_test("获取单个工作空间", False, result.get('error'))

# ========================================
# 3. 项目功能测试
# ========================================
print("\n" + "="*60)
print("3. 项目功能测试")
print("="*60)

if workspace_id:
    project_data = {
        "name": "Test Project",
        "identifier": "TESTPRJ",
        "description": "A test project for integration testing",
        "is_public": True,
        "timezone": "Asia/Shanghai"
    }
    result = make_request("POST", "/projects/", project_data, access_token, params={"workspace_id": workspace_id})
    if "error" not in result:
        log_test("创建项目", True, f"项目ID: {result.get('id', '')[:8]}...")
        project_id = result.get('id')
    else:
        log_test("创建项目", False, result.get('error'))
        project_id = None

    # 获取项目列表
    result = make_request("GET", "/projects/", token=access_token, params={"workspace_id": workspace_id})
    if "error" not in result:
        count = len(result) if isinstance(result, list) else 0
        log_test("获取项目列表", True, f"共 {count} 个项目")
    else:
        log_test("获取项目列表", False, result.get('error'))

    # 获取单个项目
    if project_id:
        result = make_request("GET", f"/projects/{project_id}", token=access_token)
        if "error" not in result:
            log_test("获取单个项目", True, f"名称: {result.get('name')}")
        else:
            log_test("获取单个项目", False, result.get('error'))

else:
    log_test("创建项目", False, "需要先创建工作空间")
    project_id = None

# ========================================
# 4. 工作项(Issue)功能测试
# ========================================
print("\n" + "="*60)
print("4. 工作项(Issue)功能测试")
print("="*60)

if project_id and workspace_id:
    # 创建工作项
    issue_data = {
        "name": "Test Issue",
        "description_html": "<p>This is a test issue</p>",
        "description_json": {"type": "doc", "content": []},
        "priority": "high"
    }
    result = make_request("POST", "/issues/", issue_data, access_token, params={"workspace_id": workspace_id, "project_id": project_id})
    if "error" not in result:
        log_test("创建工作项", True, f"工作项ID: {result.get('id', '')[:8]}...")
        issue_id = result.get('id')
    else:
        log_test("创建工作项", False, result.get('error'))
        issue_id = None

    # 获取工作项列表
    result = make_request("GET", "/issues/", token=access_token, params={"workspace_id": workspace_id, "project_id": project_id})
    if "error" not in result:
        count = len(result) if isinstance(result, list) else 0
        log_test("获取工作项列表", True, f"共 {count} 个工作项")
    else:
        log_test("获取工作项列表", False, result.get('error'))

    # 获取单个工作项
    if issue_id:
        result = make_request("GET", f"/issues/{issue_id}", token=access_token)
        if "error" not in result:
            log_test("获取单个工作项", True, f"名称: {result.get('name')}")
        else:
            log_test("获取单个工作项", False, result.get('error'))

else:
    log_test("创建工作项", False, "需要先创建工作空间和项目")
    issue_id = None

# ========================================
# 5. 自定义字段功能测试
# ========================================
print("\n" + "="*60)
print("5. 自定义字段功能测试")
print("="*60)

if project_id and workspace_id:
    # 创建自定义字段
    field_data = {
        "name": "Test Field",
        "field_type": "text",
        "is_required": False,
        "is_active": True
    }
    result = make_request("POST", "/custom-fields/", field_data, access_token, params={"workspace_id": workspace_id, "project_id": project_id})
    if "error" not in result:
        log_test("创建自定义字段", True, f"字段ID: {result.get('id', '')[:8]}...")
        field_id = result.get('id')
    else:
        log_test("创建自定义字段", False, result.get('error'))
        field_id = None

    # 获取自定义字段列表
    result = make_request("GET", "/custom-fields/", token=access_token, params={"workspace_id": workspace_id, "project_id": project_id})
    if "error" not in result:
        count = len(result) if isinstance(result, list) else 0
        log_test("获取自定义字段列表", True, f"共 {count} 个字段")
    else:
        log_test("获取自定义字段列表", False, result.get('error'))

else:
    log_test("创建自定义字段", False, "需要先创建工作空间和项目")
    field_id = None

# ========================================
# 6. 工作流功能测试
# ========================================
print("\n" + "="*60)
print("6. 工作流(Workflow)功能测试")
print("="*60)

if project_id:
    # 创建状态（状态管理在 settings 路由下）
    state_data = {
        "name": "In Progress",
        "color": "#3B82F6",
        "group": "in_progress",
        "sequence": 2,
        "project_id": str(project_id)
    }
    result = make_request("POST", f"/projects/{project_id}/settings/states/", state_data, access_token, params={"workspace_id": str(workspace_id)})
    if "error" not in result:
        log_test("创建状态", True, f"状态ID: {result.get('id', '')[:8]}...")
        state_id = result.get('id')
    else:
        log_test("创建状态", False, result.get('error'))
        state_id = None

    # 获取状态列表
    result = make_request("GET", f"/projects/{project_id}/settings/states/", token=access_token)
    if "error" not in result:
        count = len(result) if isinstance(result, list) else 0
        log_test("获取状态列表", True, f"共 {count} 个状态")
    else:
        log_test("获取状态列表", False, result.get('error'))

else:
    log_test("创建状态", False, "需要先创建工作空间和项目")
    state_id = None

# ========================================
# 7. 自动化功能测试
# ========================================
print("\n" + "="*60)
print("7. 自动化(Automation)功能测试")
print("="*60)

if project_id:
    automation_data = {
        "name": "Test Automation",
        "is_enabled": True,
        "trigger": {"type": "issue_created"},
        "conditions": [],
        "actions": [{"type": "send_notification", "config": {}}],
        "project_id": str(project_id)
    }
    result = make_request("POST", f"/projects/{project_id}/workflow/automations/", automation_data, access_token, params={"workspace_id": str(workspace_id)})
    if "error" not in result:
        log_test("创建自动化规则", True, f"规则ID: {result.get('id', '')[:8]}...")
    else:
        log_test("创建自动化规则", False, result.get('error'))

    # 获取自动化列表
    result = make_request("GET", f"/projects/{project_id}/workflow/automations/", token=access_token)
    if "error" not in result:
        count = len(result) if isinstance(result, list) else 0
        log_test("获取自动化列表", True, f"共 {count} 个规则")
    else:
        log_test("获取自动化列表", False, result.get('error'))

else:
    log_test("创建自动化规则", False, "需要先创建工作空间和项目")

# ========================================
# 8. 通知功能测试
# ========================================
print("\n" + "="*60)
print("8. 通知(Notification)功能测试")
print("="*60)

if access_token:
    result = make_request("GET", "/notifications/", token=access_token)
    if "error" not in result:
        count = len(result) if isinstance(result, list) else 0
        log_test("获取通知列表", True, f"共 {count} 条通知")
    else:
        log_test("获取通知列表", False, result.get('error'))
else:
    log_test("获取通知列表", False, "需要先登录")

# ========================================
# 9. 评论功能测试
# ========================================
print("\n" + "="*60)
print("9. 评论(Comment)功能测试")
print("="*60)

if issue_id:
    comment_data = {
        "issue_id": issue_id,
        "content": "This is a test comment"
    }
    result = make_request("POST", "/comments/", comment_data, access_token)
    if "error" not in result:
        log_test("创建评论", True, f"评论ID: {result.get('id', '')[:8]}...")
    else:
        log_test("创建评论", False, result.get('error'))
else:
    log_test("创建评论", False, "需要先创建工作项")

# ========================================
# 测试结果汇总
# ========================================
print("\n" + "="*60)
print("测试结果汇总")
print("="*60)

passed = sum(1 for r in results if r['success'])
failed = len(results) - passed
print(f"总测试数: {len(results)}")
print(f"通过: {passed}")
print(f"失败: {failed}")
print(f"通过率: {passed/len(results)*100:.1f}%" if results else "N/A")

if failed > 0:
    print("\n失败测试:")
    for r in results:
        if not r['success']:
            print(f"  - {r['name']}: {r['message']}")