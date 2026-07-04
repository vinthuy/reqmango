#!/usr/bin/env python3
"""工作流状态机和权限系统验收测试"""

import requests
import json

BASE_URL = "http://localhost:8000/api/v1"

def login():
    resp = requests.post(f"{BASE_URL}/auth/login", json={
        "email": "admin@reqmango.com",
        "password": "demo1234"
    })
    return resp.json()["access_token"]

def get_ids(token):
    headers = {"Authorization": f"Bearer {token}"}
    ws = requests.get(f"{BASE_URL}/workspaces", headers=headers).json()[0]
    proj = requests.get(f"{BASE_URL}/projects?workspace_id={ws['id']}", headers=headers).json()[0]
    return ws["id"], proj["id"]

def test(name, method, url, token, data=None, expect=200):
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    try:
        resp = getattr(requests, method)(url, headers=headers, json=data)
        if resp.status_code in (200, 201, 204) if expect < 300 else resp.status_code == expect:
            print(f"  ✓ {name}")
            try: return resp.json()
            except: return None
        else:
            print(f"  ✗ {name} - {resp.status_code}: {resp.text[:150]}")
            return None
    except Exception as e:
        print(f"  ✗ {name} - {e}")
        return None

def main():
    print("=" * 50)
    print("工作流状态机和权限系统验收测试")
    print("=" * 50)

    token = login()
    ws_id, proj_id = get_ids(token)
    print(f"Workspace: {ws_id}, Project: {proj_id}\n")

    # 1. States
    print("1. 状态管理")
    states = test("列出状态", "get", f"{BASE_URL}/projects/{proj_id}/settings/states", token)
    if states:
        print(f"    当前状态数: {len(states)}")
        for s in states[:3]:
            print(f"    - {s['name']} (group: {s.get('group', 'N/A')})")
    state = test("创建状态", "post", f"{BASE_URL}/projects/{proj_id}/settings/states?workspace_id={ws_id}", token, {
        "name": "Testing",
        "group": "in_progress",
        "color": "#f59e0b"
    })
    if state:
        test("获取状态", "get", f"{BASE_URL}/projects/{proj_id}/settings/states/{state['id']}", token)
        test("更新状态", "put", f"{BASE_URL}/projects/{proj_id}/settings/states/{state['id']}", token, {
            "name": "In Testing"
        })
        test("删除状态", "delete", f"{BASE_URL}/projects/{proj_id}/settings/states/{state['id']}", token, expect=204)
    print()

    # 2. Workflows
    print("2. 工作流管理")
    workflows = test("列出工作流", "get", f"{BASE_URL}/projects/{proj_id}/workflows", token)
    if workflows:
        print(f"    当前工作流数: {len(workflows)}")
    workflow = test("创建工作流", "post", f"{BASE_URL}/projects/{proj_id}/workflows", token, {
        "name": "Test Workflow",
        "description": "验收测试工作流"
    })
    if workflow:
        test("获取工作流", "get", f"{BASE_URL}/projects/{proj_id}/workflows/{workflow['id']}", token)
        test("更新工作流", "put", f"{BASE_URL}/projects/{proj_id}/workflows/{workflow['id']}", token, {
            "name": "Updated Workflow"
        })
        # Add transition
        transition = test("添加转换", "post", f"{BASE_URL}/projects/{proj_id}/workflows/{workflow['id']}/transitions", token, {
            "name": "Start Testing",
            "from_state_id": states[0]["id"] if states else 1,
            "to_state_id": states[1]["id"] if states and len(states) > 1 else 2
        })
        if transition:
            test("更新转换", "put", f"{BASE_URL}/projects/{proj_id}/workflows/{workflow['id']}/transitions/{transition['id']}", token, {
                "name": "Begin Testing"
            })
            test("删除转换", "delete", f"{BASE_URL}/projects/{proj_id}/workflows/{workflow['id']}/transitions/{transition['id']}", token, expect=204)
        test("删除工作流", "delete", f"{BASE_URL}/projects/{proj_id}/workflows/{workflow['id']}", token, expect=204)
    print()

    # 3. Permissions
    print("3. 权限系统")
    perms = test("列出权限", "get", f"{BASE_URL}/permissions", token)
    if perms:
        if isinstance(perms, list):
            print(f"    权限数量: {len(perms)}")
            for p in perms[:5]:
                print(f"    - {p.get('name', p.get('codename', 'N/A'))}")
        elif isinstance(perms, dict):
            print(f"    权限: {list(perms.keys())[:5]}")
    print()

    # 4. Issue State Transitions
    print("4. 工作项状态转换")
    issues = test("获取工作项", "get", f"{BASE_URL}/issues?project_id={proj_id}&limit=1", token)
    if issues and len(issues) > 0:
        issue = issues[0]
        print(f"    工作项: {issue.get('identifier', 'N/A')} - {issue.get('name', 'N/A')[:30]}")
        print(f"    当前状态: {issue.get('state', {}).get('name', 'N/A')}")
    print()

    print("=" * 50)
    print("工作流和权限验收完成！")
    print("=" * 50)

if __name__ == "__main__":
    main()
