#!/usr/bin/env python3
"""Dashboard API验收测试"""

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
    print("Dashboard API验收测试")
    print("=" * 50)

    token = login()
    ws_id, proj_id = get_ids(token)
    print(f"Workspace: {ws_id}, Project: {proj_id}\n")

    # 1. Dashboard CRUD
    print("1. Dashboard CRUD")
    dash = test("创建Dashboard", "post", f"{BASE_URL}/projects/{proj_id}/dashboards", token, {
        "name": "Test Dashboard",
        "description": "验收测试Dashboard"
    })
    test("列出Dashboard", "get", f"{BASE_URL}/projects/{proj_id}/dashboards", token)
    if dash:
        test("获取Dashboard", "get", f"{BASE_URL}/projects/{proj_id}/dashboards/{dash['id']}", token)
        test("更新Dashboard", "put", f"{BASE_URL}/projects/{proj_id}/dashboards/{dash['id']}", token, {
            "name": "Updated Dashboard"
        })
        test("设为默认", "post", f"{BASE_URL}/projects/{proj_id}/dashboards/{dash['id']}/set-default", token)
        test("复制Dashboard", "post", f"{BASE_URL}/projects/{proj_id}/dashboards/{dash['id']}/duplicate", token)
        test("获取完整Dashboard", "get", f"{BASE_URL}/projects/{proj_id}/dashboards/{dash['id']}/full", token)
    print()

    # 2. Widget CRUD
    print("2. Widget CRUD")
    if dash:
        widget = test("添加Widget", "post", f"{BASE_URL}/projects/{proj_id}/dashboards/{dash['id']}/widgets", token, {
            "widget_type": "issue_list",
            "title": "My Issues",
            "config": {"project_id": proj_id, "limit": 10}
        })
        if widget:
            test("更新Widget", "put", f"{BASE_URL}/projects/{proj_id}/dashboards/{dash['id']}/widgets/{widget['id']}", token, {
                "title": "Updated Widget"
            })
            test("删除Widget", "delete", f"{BASE_URL}/projects/{proj_id}/dashboards/{dash['id']}/widgets/{widget['id']}", token, expect=204)
    print()

    # 3. Cleanup
    print("3. 清理")
    if dash:
        test("删除Dashboard", "delete", f"{BASE_URL}/projects/{proj_id}/dashboards/{dash['id']}", token, expect=204)

    print("\n" + "=" * 50)
    print("Dashboard验收完成！")
    print("=" * 50)

if __name__ == "__main__":
    main()
