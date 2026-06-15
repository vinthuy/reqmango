"""
简化版API集成测试脚本
测试核心功能：用户认证、工作空间、项目、工作项
"""

import requests
from datetime import datetime

BASE_URL = "http://localhost:8000/api/v1"
access_token = None

def test_auth():
    print("\n=== 用户认证测试 ===")
    
    # 登录已存在的测试用户
    login_data = {"email": "pgtest@example.com", "password": "123"}
    response = requests.post(f"{BASE_URL}/auth/login", json=login_data)
    
    if response.status_code == 200:
        global access_token
        access_token = response.json().get('access_token')
        print("✅ 登录成功")
        print(f"   Token: {access_token[:30]}...")
        return True
    else:
        print(f"❌ 登录失败: {response.text}")
        return False

def test_workspace():
    print("\n=== 工作空间测试 ===")
    if not access_token:
        print("❌ 需要先登录")
        return None
    
    headers = {"Authorization": f"Bearer {access_token}", "Content-Type": "application/json"}
    
    # 创建工作空间
    workspace_data = {
        "name": "Test Workspace",
        "slug": f"test-ws-{datetime.now().strftime('%H%M%S')}",
        "timezone": "Asia/Shanghai"
    }
    response = requests.post(f"{BASE_URL}/workspaces/", json=workspace_data, headers=headers)
    
    if response.status_code == 201:
        workspace_id = response.json().get('id')
        print(f"✅ 创建工作空间成功: {workspace_id[:8]}...")
        return workspace_id
    else:
        print(f"❌ 创建工作空间失败: {response.text}")
        return None

def test_project(workspace_id):
    print("\n=== 项目测试 ===")
    if not access_token:
        print("❌ 需要先登录")
        return None
    
    headers = {"Authorization": f"Bearer {access_token}", "Content-Type": "application/json"}
    
    # 创建项目
    project_data = {
        "name": "Test Project",
        "identifier": f"TP{datetime.now().strftime('%H%M%S')}",
        "description": "Test project",
        "is_public": True,
        "timezone": "Asia/Shanghai"
    }
    response = requests.post(f"{BASE_URL}/projects/", json=project_data, headers=headers, params={"workspace_id": workspace_id})
    
    if response.status_code == 201:
        project_id = response.json().get('id')
        print(f"✅ 创建项目成功: {project_id[:8]}...")
        return project_id
    else:
        print(f"❌ 创建项目失败: {response.text}")
        return None

def test_issue(project_id, workspace_id):
    print("\n=== 工作项测试 ===")
    if not access_token:
        print("❌ 需要先登录")
        return None
    
    headers = {"Authorization": f"Bearer {access_token}", "Content-Type": "application/json"}
    
    # 创建工作项
    issue_data = {
        "name": "Test Issue",
        "description_html": "<p>Test</p>",
        "description_json": {"type": "doc", "content": []},
        "priority": "high"
    }
    response = requests.post(f"{BASE_URL}/issues/", json=issue_data, headers=headers, params={"workspace_id": workspace_id, "project_id": project_id})
    
    if response.status_code == 201:
        issue_id = response.json().get('id')
        print(f"✅ 创建工作项成功: {issue_id[:8]}...")
        return issue_id
    else:
        print(f"❌ 创建工作项失败: {response.text}")
        return None

def test_custom_field(project_id, workspace_id):
    print("\n=== 自定义字段测试 ===")
    if not access_token:
        print("❌ 需要先登录")
        return None
    
    headers = {"Authorization": f"Bearer {access_token}", "Content-Type": "application/json"}
    
    # 创建自定义字段
    field_data = {
        "name": "Test Field",
        "field_type": "text",
        "is_required": False,
        "is_active": True
    }
    response = requests.post(f"{BASE_URL}/custom-fields/", json=field_data, headers=headers, params={"workspace_id": workspace_id, "project_id": project_id})
    
    if response.status_code == 201:
        field_id = response.json().get('id')
        print(f"✅ 创建自定义字段成功: {field_id[:8]}...")
        return field_id
    else:
        print(f"❌ 创建自定义字段失败: {response.text}")
        return None

if __name__ == "__main__":
    print("🚀 开始集成测试...")
    
    # 运行测试
    if test_auth():
        workspace_id = test_workspace()
        if workspace_id:
            project_id = test_project(workspace_id)
            if project_id:
                issue_id = test_issue(project_id, workspace_id)
                test_custom_field(project_id, workspace_id)
    
    print("\n=== 测试完成 ===")