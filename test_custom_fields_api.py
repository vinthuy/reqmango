#!/usr/bin/env python3
"""自定义字段全链路API测试"""

import requests
import json
import sys
import time

BASE_URL = "http://localhost:8000/api/v1"

def login():
    resp = requests.post(f"{BASE_URL}/auth/login", json={
        "email": "admin@reqmango.com",
        "password": "demo1234"
    })
    data = resp.json()
    return data["access_token"]

def get_workspace_and_project(token):
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(f"{BASE_URL}/workspaces", headers=headers)
    ws = resp.json()[0]
    resp = requests.get(f"{BASE_URL}/projects?workspace_id={ws['id']}", headers=headers)
    proj = resp.json()[0]
    return ws["id"], proj["id"]

def test_api(name, method, url, token, json_data=None, expected_status=200):
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    try:
        resp = getattr(requests, method)(url, headers=headers, json=json_data)
        if resp.status_code in (200, 201, 204) if isinstance(expected_status, int) and expected_status < 300 else resp.status_code == expected_status:
            print(f"  ✓ {name}")
            try:
                return resp.json()
            except:
                return None
        else:
            print(f"  ✗ {name} - Expected {expected_status}, got {resp.status_code}")
            print(f"    Response: {resp.text[:200]}")
            return None
    except Exception as e:
        print(f"  ✗ {name} - Error: {e}")
        return None

def main():
    print("=" * 60)
    print("自定义字段全链路API测试")
    print("=" * 60)

    # Login
    token = login()
    ws_id, proj_id = get_workspace_and_project(token)
    print(f"Workspace: {ws_id}, Project: {proj_id}\n")

    # 1. Custom Field CRUD
    print("1. 自定义字段CRUD")
    field_text = test_api("创建文本字段", "post", f"{BASE_URL}/custom-fields?workspace_id={ws_id}", token, {
        "name": "Test-Text-Field",
        "field_type": "text",
        "project_id": proj_id,
        "is_active": True
    })
    field_number = test_api("创建数字字段", "post", f"{BASE_URL}/custom-fields?workspace_id={ws_id}", token, {
        "name": "Test-Number-Field",
        "field_type": "number",
        "project_id": proj_id,
        "is_active": True
    })
    field_select = test_api("创建下拉字段", "post", f"{BASE_URL}/custom-fields?workspace_id={ws_id}", token, {
        "name": "Test-Select-Field",
        "field_type": "dropdown",
        "project_id": proj_id,
        "is_active": True,
        "options": [{"label": "Option A", "value": "option_a"}, {"label": "Option B", "value": "option_b"}]
    })
    field_bool = test_api("创建布尔字段", "post", f"{BASE_URL}/custom-fields?workspace_id={ws_id}", token, {
        "name": "Test-Bool-Field",
        "field_type": "boolean",
        "project_id": proj_id,
        "is_active": True
    })
    field_date = test_api("创建日期字段", "post", f"{BASE_URL}/custom-fields?workspace_id={ws_id}", token, {
        "name": "Test-Date-Field",
        "field_type": "date",
        "project_id": proj_id,
        "is_active": True
    })
    field_url = test_api("创建URL字段", "post", f"{BASE_URL}/custom-fields?workspace_id={ws_id}", token, {
        "name": "Test-URL-Field",
        "field_type": "url",
        "project_id": proj_id,
        "is_active": True
    })
    print()

    # 2. List & Query
    print("2. 列表和查询")
    test_api("列出所有字段", "get", f"{BASE_URL}/custom-fields?workspace_id={ws_id}", token)
    if field_text:
        test_api("按ID查询字段", "get", f"{BASE_URL}/custom-fields/{field_text['id']}", token)
    print()

    # 3. Link to Issue Type
    print("3. 关联工作项类型")
    issue_types = test_api("获取工作项类型", "get", f"{BASE_URL}/issue-types?project_id={proj_id}&workspace_id={ws_id}", token)
    if issue_types and len(issue_types) > 0:
        it_id = issue_types[0]["id"]
        if field_text:
            test_api("链接字段到类型", "post", f"{BASE_URL}/issue-types/{it_id}/fields", token, {"field_id": field_text["id"], "is_required": False})
        if field_number:
            test_api("链接数字字段", "post", f"{BASE_URL}/issue-types/{it_id}/fields", token, {"field_id": field_number["id"], "is_required": False})
        if field_select:
            test_api("链接下拉字段", "post", f"{BASE_URL}/issue-types/{it_id}/fields", token, {"field_id": field_select["id"], "is_required": False})
        if field_bool:
            test_api("链接布尔字段", "post", f"{BASE_URL}/issue-types/{it_id}/fields", token, {"field_id": field_bool["id"], "is_required": False})
        if field_date:
            test_api("链接日期字段", "post", f"{BASE_URL}/issue-types/{it_id}/fields", token, {"field_id": field_date["id"], "is_required": False})
        if field_url:
            test_api("链接URL字段", "post", f"{BASE_URL}/issue-types/{it_id}/fields", token, {"field_id": field_url["id"], "is_required": False})
    print()

    # 4. Create Issues with Custom Field Values
    print("4. 创建工作项（带自定义字段值）")
    states = test_api("获取状态", "get", f"{BASE_URL}/projects/{proj_id}/settings/states", token)
    state_id = states[0]["id"] if states else 1

    issue1 = test_api("创建工作项1", "post", f"{BASE_URL}/issues?project_id={proj_id}&workspace_id={ws_id}", token, {
        "name": "Test Issue with Custom Fields",
        "project_id": proj_id,
        "state_id": state_id,
        "issue_type_id": issue_types[0]["id"] if issue_types else 1,
        "custom_field_values": {
            str(field_text["id"]): "Hello World" if field_text else "",
            str(field_number["id"]): "42" if field_number else "",
            str(field_select["id"]): "option_a" if field_select else "",
            str(field_bool["id"]): "true" if field_bool else "",
            str(field_date["id"]): "2026-07-04" if field_date else ""
        }
    })
    issue2 = test_api("创建工作项2", "post", f"{BASE_URL}/issues?project_id={proj_id}&workspace_id={ws_id}", token, {
        "name": "Test Issue with Different Values",
        "project_id": proj_id,
        "state_id": state_id,
        "issue_type_id": issue_types[0]["id"] if issue_types else 1,
        "custom_field_values": {
            str(field_text["id"]): "Goodbye World" if field_text else "",
            str(field_number["id"]): "100" if field_number else "",
            str(field_select["id"]): "option_b" if field_select else "",
            str(field_bool["id"]): "false" if field_bool else "",
            str(field_date["id"]): "2026-12-31" if field_date else ""
        }
    })
    print()

    # 5. Query Custom Field Values
    print("5. 查询自定义字段值")
    if issue1:
        test_api("查询工作项1字段值", "get", f"{BASE_URL}/issues/{issue1['id']}", token)
    if issue2:
        test_api("查询工作项2字段值", "get", f"{BASE_URL}/issues/{issue2['id']}", token)
    print()

    # 6. RQL Queries
    print("6. RQL查询")
    if field_text and issue1:
        test_api("文本精确匹配", "post", f"{BASE_URL}/rql/search", token, {
            "entity": "issue",
            "project_id": proj_id,
            "rql": f'cf_{field_text["name"]} = "Hello World"'
        })
        test_api("文本模糊匹配", "post", f"{BASE_URL}/rql/search", token, {
            "entity": "issue",
            "project_id": proj_id,
            "rql": f'cf_{field_text["name"]} ~ "Hello"'
        })
    if field_number:
        test_api("数字大于", "post", f"{BASE_URL}/rql/search", token, {
            "entity": "issue",
            "project_id": proj_id,
            "rql": f'cf_{field_number["name"]} > 50'
        })
        test_api("数字小于", "post", f"{BASE_URL}/rql/search", token, {
            "entity": "issue",
            "project_id": proj_id,
            "rql": f'cf_{field_number["name"]} < 50'
        })
        test_api("数字等于", "post", f"{BASE_URL}/rql/search", token, {
            "entity": "issue",
            "project_id": proj_id,
            "rql": f'cf_{field_number["name"]} = 42'
        })
    if field_select:
        test_api("下拉字段匹配", "post", f"{BASE_URL}/rql/search", token, {
            "entity": "issue",
            "project_id": proj_id,
            "rql": f'cf_{field_select["name"]} = "option_a"'
        })
    if field_bool:
        test_api("布尔字段匹配", "post", f"{BASE_URL}/rql/search", token, {
            "entity": "issue",
            "project_id": proj_id,
            "rql": f'cf_{field_bool["name"]} = true'
        })
    if field_text and field_number:
        test_api("多条件AND查询", "post", f"{BASE_URL}/rql/search", token, {
            "entity": "issue",
            "project_id": proj_id,
            "rql": f'cf_{field_text["name"]} ~ "Hello" AND cf_{field_number["name"]} > 10'
        })
    print()

    # 7. Update Custom Field Values
    print("7. 更新自定义字段值")
    if issue1 and field_text:
        test_api("更新单个字段值", "put", f"{BASE_URL}/custom-fields/issues/{issue1['id']}/values/{field_text['id']}", token, {
            "value": "Updated Hello World"
        })
    if issue1 and field_number:
        test_api("批量更新字段值", "post", f"{BASE_URL}/custom-fields/issues/{issue1['id']}/values/bulk", token, {
            "issue_id": issue1["id"],
            "values": [{"field_id": field_number["id"], "value": "999"}]
        })
    print()

    # 8. Cleanup - Delete Issues and Fields
    print("8. 清理测试数据")
    if issue1:
        test_api("删除工作项1", "delete", f"{BASE_URL}/issues/{issue1['id']}", token, expected_status=204)
    if issue2:
        test_api("删除工作项2", "delete", f"{BASE_URL}/issues/{issue2['id']}", token, expected_status=204)
    for f in [field_text, field_number, field_select, field_bool, field_date, field_url]:
        if f:
            test_api(f"删除字段 {f['name']}", "delete", f"{BASE_URL}/custom-fields/{f['id']}", token, expected_status=204)
    print()

    print("=" * 60)
    print("测试完成！")
    print("=" * 60)

if __name__ == "__main__":
    main()
