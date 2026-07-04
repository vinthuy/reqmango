#!/usr/bin/env python3
"""通知系统验收测试"""

import requests
import json

BASE_URL = "http://localhost:8000/api/v1"

def login():
    resp = requests.post(f"{BASE_URL}/auth/login", json={
        "email": "admin@reqmango.com",
        "password": "demo1234"
    })
    return resp.json()["access_token"]

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
    print("通知系统验收测试")
    print("=" * 50)

    token = login()

    # 1. List notifications
    print("1. 通知列表")
    notifs = test("列出通知", "get", f"{BASE_URL}/notifications", token)
    if notifs:
        if isinstance(notifs, list):
            print(f"    通知数量: {len(notifs)}")
        elif isinstance(notifs, dict):
            print(f"    通知: {list(notifs.keys())[:5]}")
    print()

    # 2. Summary
    print("2. 通知摘要")
    summary = test("获取摘要", "get", f"{BASE_URL}/notifications/summary", token)
    if summary:
        print(f"    摘要: {summary}")
    print()

    # 3. Create notification
    print("3. 创建通知")
    notif = test("创建通知", "post", f"{BASE_URL}/notifications", token, {
        "title": "Test Notification",
        "message": "This is a test notification",
        "type": "info"
    })
    if notif:
        test("获取通知", "get", f"{BASE_URL}/notifications/{notif['id']}", token)
        test("标记已读", "patch", f"{BASE_URL}/notifications/{notif['id']}/read", token)
        test("删除通知", "delete", f"{BASE_URL}/notifications/{notif['id']}", token, expect=204)
    print()

    # 4. Bulk operations
    print("4. 批量操作")
    test("全部标记已读", "post", f"{BASE_URL}/notifications/read-all", token)
    test("检查提醒", "post", f"{BASE_URL}/notifications/check-due-reminders", token)
    print()

    print("=" * 50)
    print("通知系统验收完成！")
    print("=" * 50)

if __name__ == "__main__":
    main()
