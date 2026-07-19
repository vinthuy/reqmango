"""
种子数据自动化测试 - 产生丰富的 automation_executions 记录
测试场景：
  1. any: 任意子项 In Progress → 父级 In Progress
  2. all: 所有子项 Done → 父级 Done  
  3. any: 任意子项 In Review → 父级 In Review
  4. 混合: all Done → parent Done + any In Progress → parent In Progress (2条规则同时)
"""
import requests
import json
import time
import sys

BASE = "http://localhost:8000/api/v1"
EMAIL = "admin@reqmango.com"
PASSWORD = "demo1234"

def header(token):
    return {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}

def login():
    r = requests.post(f"{BASE}/auth/login", json={"email": EMAIL, "password": PASSWORD})
    return r.json()["access_token"]

def main():
    print("=" * 60)
    print("  Seed Data Automation Test Suite")
    print("=" * 60)
    
    token = login()
    h = header(token)
    print("\n[OK] Login successful")

    # Get workspace
    pj = requests.get(f"{BASE}/projects/1", headers=h).json()
    ws = pj["workspace_id"]

    # =============================================
    # STEP 1: Clean up old E2E data
    # =============================================
    print("\n" + "=" * 40)
    print("  STEP 1: Cleanup old E2E data")
    print("=" * 40)
    
    # Delete old automation rules
    rules_resp = requests.get(f"{BASE}/projects/1/automations", headers=h)
    for rule in rules_resp.json():
        if "E2E" in rule.get("name", "") or "Seed" in rule.get("name", ""):
            requests.delete(f"{BASE}/projects/1/automations/{rule['id']}", headers=h)
            print(f"  Deleted rule #{rule['id']} '{rule['name']}'")

    # Delete old E2E issues
    issues_resp = requests.get(f"{BASE}/issues?project_id=1&limit=200", headers=h)
    for iss in issues_resp.json():
        if "E2E" in iss.get("name", "") or "Seed" in iss.get("name", ""):
            requests.delete(f"{BASE}/issues/{iss['id']}", headers=h)
            print(f"  Deleted issue #{iss['id']} '{iss['name']}'")

    # =============================================
    # STEP 2: Create test structures
    # =============================================
    print("\n" + "=" * 40)
    print("  STEP 2: Create test structures")
    print("=" * 40)

    # Create parent issues (3 groups)
    parents = []
    for g in range(1, 4):
        pname = f"(Seed) Parent-{g}"
        p = requests.post(f"{BASE}/issues?project_id=1&workspace_id={ws}", 
                         headers=h, json={"name": pname, "state_id": 2}).json()
        p["children"] = []
        parents.append(p)
        print(f"  Created Parent #{p['id']} '{pname}' (Todo)")

    # Create 3 children per parent
    all_children = []
    for pg, parent in enumerate(parents):
        pid = parent["id"]
        for c in range(1, 4):
            cname = f"(Seed) Child-{pg+1}-{c}"
            child = requests.post(f"{BASE}/issues?project_id=1&workspace_id={ws}",
                                 headers=h, json={"name": cname, "state_id": 2, "parent_id": pid}).json()
            parent["children"].append(child)
            all_children.append(child)
            print(f"  Created Child #{child['id']} '{cname}' (Todo, parent=#{pid})")

    # =============================================
    # STEP 3: Create automation rules with raw JSON
    # =============================================
    print("\n" + "=" * 40)
    print("  STEP 3: Create automation rules")
    print("=" * 40)

    rules_config = [
        # Rule 1: any In Progress → parent In Progress
        {
            "name": "(Seed) R1: any InProgress",
            "actions": '[{"type":"rollup_to_parent","value":{"rules":[{"condition":"any","child_state":"In Progress","parent_state":"In Progress"}]}}]'
        },
        # Rule 2: all Done → parent Done
        {
            "name": "(Seed) R2: all Done",
            "actions": '[{"type":"rollup_to_parent","value":{"rules":[{"condition":"all","child_state":"Done","parent_state":"Done"}]}}]'
        },
        # Rule 3: any In Review → parent In Review  
        {
            "name": "(Seed) R3: any InReview",
            "actions": '[{"type":"rollup_to_parent","value":{"rules":[{"condition":"any","child_state":"In Review","parent_state":"In Review"}]}}]'
        },
        # Rule 4: combined - anyInProgress + allDone
        {
            "name": "(Seed) R4: combined any+all",
            "actions": '[{"type":"rollup_to_parent","value":{"rules":[{"condition":"any","child_state":"In Progress","parent_state":"In Progress"},{"condition":"all","child_state":"Done","parent_state":"Done"}]}}]'
        },
    ]

    rules = []
    for rc in rules_config:
        body = {
            "name": rc["name"],
            "description": "Seed data test",
            "trigger_type": "issue.state_changed",
            "is_enabled": True,
            "actions": rc["actions"]
        }
        r = requests.post(f"{BASE}/projects/1/automations", headers=h, json=body)
        rd = r.json()
        rules.append(rd)
        # Verify storage
        if rd.get("actions", "").startswith("["):
            print(f"  Created Rule #{rd['id']} '{rd['name']}' [OK]")
        else:
            print(f"  Created Rule #{rd['id']} '{rd['name']}' [WARN: actions not array!]")

    # =============================================
    # STEP 4: Trigger state changes
    # =============================================
    print("\n" + "=" * 40)
    print("  STEP 4: Trigger state changes")
    print("=" * 40)

    # Group 1: Test "any InProgress" rule
    # Change Child-1-1 -> InProgress (3), should trigger parent -> InProgress
    print("\n  --- Group 1: any InProgress ---")
    c = parents[0]["children"]
    print(f"  Child #{c[0]['id']} -> InProgress (3)")
    requests.put(f"{BASE}/issues/{c[0]['id']}", headers=h, json={"state_id": 3})
    time.sleep(2)
    p1 = requests.get(f"{BASE}/issues/{parents[0]['id']}", headers=h).json()
    print(f"  Parent #{p1['id']} state_id={p1['state_id']} {'PASS' if p1['state_id']==3 else 'FAIL'}")

    # Change Child-1-2 -> InProgress too (still any, parent already InProgress)
    print(f"  Child #{c[1]['id']} -> InProgress (3)")
    requests.put(f"{BASE}/issues/{c[1]['id']}", headers=h, json={"state_id": 3})
    time.sleep(2)

    # Group 2: Test "all Done" rule  
    # First change Child-2-1 -> InProgress -> Done
    print("\n  --- Group 2: all Done ---")
    c = parents[1]["children"]
    # Need to go through workflow: Todo(2) -> InProgress(3) -> Done(5) directly
    for i, child in enumerate(c):
        print(f"  Child #{child['id']} -> InProgress(3)")
        requests.put(f"{BASE}/issues/{child['id']}", headers=h, json={"state_id": 3})
        time.sleep(0.5)
    time.sleep(2)
    # Parent should now be InProgress (R1 rule triggered)
    p2 = requests.get(f"{BASE}/issues/{parents[1]['id']}", headers=h).json()
    print(f"  Parent #{p2['id']} state_id={p2['state_id']} (any InProgress)")

    # Now try Done
    for i, child in enumerate(c):
        print(f"  Child #{child['id']} -> Done(5)")
        try:
            requests.put(f"{BASE}/issues/{child['id']}", headers=h, json={"state_id": 5})
        except Exception as e:
            print(f"    (workflow restriction, skipping)")
        time.sleep(0.5)
    time.sleep(3)
    p2 = requests.get(f"{BASE}/issues/{parents[1]['id']}", headers=h).json()
    print(f"  Parent #{p2['id']} state_id={p2['state_id']} (expect Done=5 if all children Done)")

    # Group 3: Test "any In Review" rule
    print("\n  --- Group 3: any InReview ---")
    c = parents[2]["children"]
    print(f"  Child #{c[0]['id']} -> InProgress(3)")
    requests.put(f"{BASE}/issues/{c[0]['id']}", headers=h, json={"state_id": 3})
    time.sleep(1)
    print(f"  Child #{c[0]['id']} -> InReview(4)")
    requests.put(f"{BASE}/issues/{c[0]['id']}", headers=h, json={"state_id": 4})
    time.sleep(2)
    p3 = requests.get(f"{BASE}/issues/{parents[2]['id']}", headers=h).json()
    print(f"  Parent #{p3['id']} state_id={p3['state_id']} {'PASS' if p3['state_id']==4 else 'INFO'} (any InReview)")

    # Also change other children to trigger more executions
    print(f"  Child #{c[1]['id']} -> InProgress(3)")
    requests.put(f"{BASE}/issues/{c[1]['id']}", headers=h, json={"state_id": 3})
    time.sleep(1)
    print(f"  Child #{c[2]['id']} -> InProgress(3)")
    requests.put(f"{BASE}/issues/{c[2]['id']}", headers=h, json={"state_id": 3})
    time.sleep(2)

    # =============================================
    # STEP 5: Verify execution records
    # =============================================
    print("\n" + "=" * 40)
    print("  STEP 5: Execution history verification")
    print("=" * 40)

    for rule in rules:
        rid = rule["id"]
        try:
            resp = requests.get(f"{BASE}/projects/1/automations/{rid}/execution-history", headers=h)
            data = resp.json()
            if isinstance(data, dict):
                total = data.get("total", len(data.get("executions", [])))
                execs = data.get("executions", [])
            elif isinstance(data, list):
                execs = data
                total = len(data)
            else:
                total = 0
                execs = []
            print(f"  Rule #{rid} ({rule['name']}): {total} executions")
            # Show first execution
            if execs:
                e = execs[0]
                st = e.get("status", "?")
                dur = e.get("duration", 0)
                iss = e.get("issue_id", "?")
                print(f"    Latest: status={st} issue=#{iss} duration={dur}ms")
        except Exception as e:
            print(f"  Rule #{rid}: ERROR - {e}")

    # =============================================
    # SUMMARY
    # =============================================
    print("\n" + "=" * 60)
    print("  SUMMARY")
    print("=" * 60)
    print(f"  Parents created: {len(parents)}")
    print(f"  Children created: {len(all_children)}")
    print(f"  Rules created: {len(rules)}")
    print(f"  Rules / Issues will persist for manual UI verification")
    print(f"  Automation execution records generated in DB")
    print("=" * 60)

if __name__ == "__main__":
    main()
