#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import requests, json, sys
BASE = "http://localhost:8000/api/v1"
PASS = 0; FAIL = 0; RESULTS = []
HEADERS = {}
TOKEN = None; WS_ID = None; WS_SLUG = None; PROJ_ID = None
ISSUE_ID = None; STATE_ID = None; CYCLE_ID = None; MOD_ID = None
RT_ID = None; REL_ID = None; SV_ID = None

def test(name, fn):
    global PASS, FAIL
    try:
        r = fn()
        if r is not False:
            PASS += 1; RESULTS.append(f"  [PASS] {name}")
        else:
            FAIL += 1; RESULTS.append(f"  [FAIL] {name}")
    except Exception as e:
        FAIL += 1; RESULTS.append(f"  [ERROR] {name}: {str(e)[:120]}")

def api_raw(method, path, data=None, params=None):
    url = f"{BASE}{path}"
    try:
        resp = requests.request(method, url, json=data, params=params, headers=HEADERS, timeout=15)
        return resp
    except Exception as e:
        print(f"    Connection error: {e}")
        return None

# Accept 200/201/204 as success
def api_ok(method, path, data=None, params=None):
    resp = api_raw(method, path, data, params)
    if resp is None: return False
    if resp.status_code not in [200,201,204]:
        print(f"    HTTP {resp.status_code}: {resp.text[:200]}")
        return False
    return resp

def aj(method, path, data=None, params=None):
    resp = api_ok(method, path, data, params)
    if resp is False: return False
    try: return resp.json()
    except: return resp.text

def first_id(items):
    if isinstance(items, list) and items: return items[0].get("id")
    if isinstance(items, dict):
        for k in ("items","data","workspaces","projects","modules","cycles","issues"):
            if k in items and items[k] and isinstance(items[k],list):
                return items[k][0].get("id")
    return None

print("=" * 60)
print("ReqMango Full Feature Test")
print("=" * 60)

# ===== 1. Auth =====
print("\n[1] Auth")
def t_login():
    global TOKEN, HEADERS
    resp = aj("POST","/auth/login",{"email":"admin@reqmango.com","password":"demo1234"})
    if not resp: return False
    TOKEN = resp.get("access_token") or resp.get("token")
    if not TOKEN:
        for k,v in (resp.items() if isinstance(resp,dict) else []):
            if isinstance(v,str) and len(v)>20: TOKEN=v; break
    if not TOKEN: return False
    HEADERS = {"Authorization": f"Bearer {TOKEN}"}; return True

def t_me():
    resp = aj("GET","/auth/me")
    return resp is not False

test("Login (admin@reqmango.com / demo1234)", t_login)
test("Get current user (/auth/me)", t_me)
if not TOKEN: print("\n[FAIL] Auth failed."); sys.exit(1)

# ===== 2. Workspace =====
print("\n[2] Workspace")
def t_ws_list():
    global WS_ID, WS_SLUG
    resp = aj("GET","/workspaces")
    if not resp: return False
    items = resp if isinstance(resp,list) else resp.get("workspaces",resp.get("data",[]))
    if not items: return False
    WS_ID = items[0].get("id"); WS_SLUG = items[0].get("slug","")
    return WS_ID is not None

def t_ws_get():
    # Use slug if available, otherwise numeric ID
    slug = WS_SLUG if WS_SLUG else str(WS_ID)
    resp = aj("GET",f"/workspaces/{slug}")
    if resp and resp.get("name"): return True
    # Fallback: try numeric ID
    resp2 = aj("GET",f"/workspaces/{WS_ID}")
    return resp2 is not False and resp2.get("name") is not None

def t_ws_members():
    slug = WS_SLUG if WS_SLUG else str(WS_ID)
    resp = aj("GET",f"/workspaces/{slug}/members")
    return resp is not False

test("List workspaces", t_ws_list)
test("Get workspace detail", t_ws_get)
test("List workspace members", t_ws_members)

# ===== 3. Project =====
print("\n[3] Project")
def t_proj_list():
    global PROJ_ID
    resp = aj("GET","/projects",params={"workspace_id":WS_ID})
    fid = first_id(resp)
    if fid: PROJ_ID = fid
    return PROJ_ID is not None

def t_proj_get():
    resp = aj("GET",f"/projects/{PROJ_ID}")
    return resp and resp.get("name")

def t_proj_members():
    return aj("GET",f"/projects/{PROJ_ID}/members") is not False

def t_proj_states():
    global STATE_ID
    resp = aj("GET",f"/projects/{PROJ_ID}/settings/states")
    fid = first_id(resp)
    if fid: STATE_ID = fid
    return fid is not None

def t_statistics():
    return aj("GET","/issues/statistics",params={"project_id":PROJ_ID}) is not False

def t_flow_metrics():
    return aj("GET","/issues/flow-metrics",params={"project_id":PROJ_ID}) is not False

test("List projects", t_proj_list)
test("Get project detail", t_proj_get)
test("List project members", t_proj_members)
test("List project states", t_proj_states)
test("Issue statistics", t_statistics)
test("Flow metrics", t_flow_metrics)

# ===== 4. Issue CRUD =====
print("\n[4] Issue CRUD")
def t_issue_list():
    global ISSUE_ID
    resp = aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"limit":5})
    fid = first_id(resp)
    if fid: ISSUE_ID = fid
    return ISSUE_ID is not None

def t_issue_get():
    resp = aj("GET",f"/issues/{ISSUE_ID}")
    return resp and resp.get("name")

def t_issue_create():
    resp = aj("POST","/issues",{
        "project_id":PROJ_ID,"workspace_id":WS_ID,
        "name":"AutoTest Issue","description_html":"<p>Auto test</p>",
        "priority":"medium","state_id":STATE_ID
    },params={"workspace_id":WS_ID,"project_id":PROJ_ID})
    if resp and resp.get("id"): return True
    return False

def t_issue_update():
    # Use PUT not PATCH
    resp = aj("PUT",f"/issues/{ISSUE_ID}",{"name":"AutoTest [Updated]","priority":"high"})
    return resp is not False

def t_issue_activities():
    return aj("GET",f"/issues/{ISSUE_ID}/activities") is not False

def t_issue_children():
    return aj("GET",f"/issues/{ISSUE_ID}/children") is not False

test("List issues (paginated)", t_issue_list)
test("Get issue detail", t_issue_get)
test("Create issue", t_issue_create)
test("Update issue", t_issue_update)
test("Issue activities", t_issue_activities)
test("Issue children (tree)", t_issue_children)

# ===== 5. Search & Sort =====
print("\n[5] Search & Sort")
test("Search issues", lambda: aj("GET","/issues/search",params={"workspace_id":WS_ID,"query":"AutoTest"}) is not False)
test("Issue suggest", lambda: aj("GET","/issues/suggest",params={"project_id":PROJ_ID,"query":"Au","limit":5}) is not False)
test("Bulk update", lambda: aj("POST","/issues/bulk/update",{"issue_ids":[ISSUE_ID],"priority":"low"},params={"project_id":PROJ_ID}) is not False)

# ===== 6. RQL =====
print("\n[6] RQL Filters")
test('RQL: name = "AutoTest"', lambda: aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"rql":'name = "AutoTest"',"limit":5}) is not False)
test('RQL: priority = "medium"', lambda: aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"rql":'priority = "medium"',"limit":5}) is not False)
test('RQL: state_group = "unstarted"', lambda: aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"rql":'state_group = "unstarted"',"limit":5}) is not False)
test('RQL: priority != "low"', lambda: aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"rql":'priority != "low"',"limit":5}) is not False)
test('RQL: name LIKE "Auto"', lambda: aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"rql":'name LIKE "Auto"',"limit":5}) is not False)

# ===== 7. Cycle & Module =====
print("\n[7] Cycle & Module")
def t_cycle_list():
    global CYCLE_ID
    resp = aj("GET",f"/projects/{PROJ_ID}/cycles")
    fid = first_id(resp)
    if fid: CYCLE_ID = fid
    return CYCLE_ID is not None

def t_cycle_get():
    return aj("GET",f"/cycles/{CYCLE_ID}") is not False

def t_mod_list():
    global MOD_ID
    resp = aj("GET","/modules",params={"project_id":PROJ_ID,"workspace_id":WS_ID})
    fid = first_id(resp)
    if fid: MOD_ID = fid
    return MOD_ID is not None

def t_mod_get():
    return aj("GET",f"/modules/{MOD_ID}") is not False

def t_mod_create():
    resp = aj("POST","/modules",{"name":"AutoTest Module","project_id":PROJ_ID,"workspace_id":WS_ID,"description":"test"},params={"workspace_id":WS_ID})
    if resp and resp.get("id"):
        global MOD_ID; MOD_ID = resp["id"]; return True
    return False

test("List cycles", t_cycle_list)
test("Get cycle detail", t_cycle_get)
test("List modules", t_mod_list)
test("Get module detail", t_mod_get)
test("Create module", t_mod_create)

# ===== 8. View Types =====
print("\n[8] Issue Views")
test("List view", lambda: aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"view_type":"list","limit":5}) is not False)
test("Kanban view (group_by=state_id)", lambda: aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"view_type":"kanban","group_by":"state_id","limit":30}) is not False)
test("Group by priority", lambda: aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"group_by":"priority","limit":20}) is not False)
test("Sort by priority desc", lambda: aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"sort_by":"priority","sort_dir":"desc","limit":10}) is not False)

# ===== 9. Relation Types =====
print("\n[9] Relation Types")
def t_rt_list():
    global RT_ID
    resp = aj("GET","/relations/types",params={"workspace_id":WS_ID})
    fid = first_id(resp)
    if fid: RT_ID = fid
    return RT_ID is not None

def t_rt_create():
    resp = aj("POST","/relations/types",{"name":"test-auto","inward_name":"tested by","outward_name":"tests"},params={"workspace_id":WS_ID})
    if resp and resp.get("id"): return True
    return False

test("List relation types", t_rt_list)
test("Create relation type", t_rt_create)

# ===== 10. Issue Relations =====
print("\n[10] Issue Relations")
def t_rel_create():
    global REL_ID
    resp = aj("GET","/issues",params={"project_id":PROJ_ID,"workspace_id":WS_ID,"limit":3})
    items = resp if isinstance(resp,list) else resp.get("issues",resp.get("data",resp.get("items",[])))
    target_id = None
    for it in items:
        if it.get("id") != ISSUE_ID: target_id = it.get("id"); break
    if not target_id: return False
    resp2 = aj("POST",f"/issues/{ISSUE_ID}/relations",{"related_issue_id":target_id,"relation_type_id":RT_ID})
    if resp2 and resp2.get("id"): REL_ID = resp2["id"]; return True
    return False

def t_rel_list():
    return aj("GET",f"/issues/{ISSUE_ID}/relations") is not False

def t_rel_delete():
    if not REL_ID: return False
    return api_ok("DELETE",f"/relations/{REL_ID}") is not False

def t_rt_delete():
    return api_ok("DELETE",f"/relations/types/{RT_ID}") is not False

test("Create issue relation", t_rel_create)
test("List issue relations", t_rel_list)
test("Delete issue relation", t_rel_delete)
test("Delete relation type", t_rt_delete)

# ===== 11. Saved Views =====
print("\n[11] Saved Views")
def t_sv_create():
    global SV_ID
    resp = aj("POST",f"/projects/{PROJ_ID}/views",{
        "name":"Auto Test View","description":"Auto test",
        "view_type":"list","project_id":PROJ_ID,
        "filters":'[{"field":"priority","operator":"is","value":"high"}]',
        "rql":'priority = "high"',
        "sort_config":'[{"field":"created_at","dir":"desc"}]',
        "columns":'["name","state","priority","assignee"]',
        "group_by":"state_id"
    },params={"workspace_id":WS_ID})
    if resp and resp.get("id"): SV_ID = resp["id"]; return True
    return False

def t_sv_list():
    return aj("GET",f"/projects/{PROJ_ID}/views") is not False

def t_sv_get():
    if not SV_ID: return False
    return aj("GET",f"/projects/{PROJ_ID}/views/{SV_ID}") is not False

def t_sv_update():
    if not SV_ID: return False
    return aj("PUT",f"/projects/{PROJ_ID}/views/{SV_ID}",{"name":"Auto Test View [Updated]"}) is not False

def t_sv_delete():
    if not SV_ID: return False
    return api_ok("DELETE",f"/projects/{PROJ_ID}/views/{SV_ID}") is not False

test("Create saved view (sort_config, columns, group_by)", t_sv_create)
test("List saved views", t_sv_list)
test("Get saved view", t_sv_get)
test("Update saved view", t_sv_update)
test("Delete saved view", t_sv_delete)

# ===== 12. Custom Fields =====
print("\n[12] Custom Fields")
test("List custom fields", lambda: aj("GET","/custom-fields",params={"workspace_id":WS_ID}) is not False)
test("Issue custom field values", lambda: aj("GET",f"/custom-fields/issues/{ISSUE_ID}/values") is not False)

# ===== 13. Labels & Issue Types =====
print("\n[13] Labels & Issue Types")
test("List labels", lambda: aj("GET",f"/projects/{PROJ_ID}/settings/labels") is not False)
test("Workspace issue types", lambda: aj("GET","/issue-types",params={"workspace_id":WS_ID}) is not False)
test("Project issue types", lambda: aj("GET",f"/projects/{PROJ_ID}/issue-types",params={"workspace_id":WS_ID}) is not False)

# ===== 14. Comments =====
print("\n[14] Comments")
test("List comments", lambda: aj("GET",f"/comments/issue/{ISSUE_ID}") is not False)
test("Create comment", lambda: aj("POST","/comments",{"issue_id":ISSUE_ID,"body":"Auto test comment","activity_type":"comment"}) is not False)

# ===== 15. Attachments =====
print("\n[15] Attachments")
test("List attachments", lambda: aj("GET",f"/issues/{ISSUE_ID}/attachments") is not False)

# ===== 16. Pages =====
print("\n[16] Pages")
test("List pages", lambda: aj("GET",f"/projects/{PROJ_ID}/pages") is not False)

# ===== 17. Time Track =====
print("\n[17] Time Track")
test("List time tracks", lambda: aj("GET",f"/issues/{ISSUE_ID}/time-tracks") is not False)

# ===== 18. Cleanup =====
print("\n[18] Cleanup")
test("Delete test issue", lambda: api_ok("DELETE",f"/issues/{ISSUE_ID}") is not False)

# ===== Summary =====
print("\n" + "=" * 60)
total = PASS + FAIL
print(f"Results: {PASS}/{total} passed, {FAIL} failed")
print("=" * 60)
for r in RESULTS: print(r)
if FAIL > 0:
    print(f"\n[FAIL] {FAIL} test(s) failed!")
    sys.exit(1)
else:
    print("\n[OK] All tests passed!"); sys.exit(0)
