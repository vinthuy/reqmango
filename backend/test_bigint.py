import requests
import json
import time

# 1. 注册用户
print('=== 1. 注册用户 ===')
response = requests.post('http://localhost:8001/api/v1/auth/register', json={
    'email': 'test2@example.com',
    'username': 'testuser2',
    'password': 'test123456',
    'display_name': 'Test User 2'
})
print('Status:', response.status_code)
if response.status_code == 201:
    print('注册成功!')
elif response.status_code == 409:
    print('用户已存在，跳过注册')
else:
    print('Response:', response.text[:200])

# 2. 登录
print('\n=== 2. 登录 ===')
response = requests.post('http://localhost:8001/api/v1/auth/login', json={
    'email': 'test2@example.com',
    'password': 'test123456'
})
print('Status:', response.status_code)
if response.status_code == 200:
    token = response.json().get('access_token')
    print('登录成功!')
    headers = {'Authorization': f'Bearer {token}'}
else:
    print('登录失败:', response.text)
    exit(1)

# 3. 创建工作空间
print('\n=== 3. 创建工作空间 ===')
response = requests.post('http://localhost:8001/api/v1/workspaces', headers=headers, json={
    'name': 'demo2',
    'slug': 'demo2-' + str(int(time.time()))
})
print('Status:', response.status_code)
if response.status_code == 201:
    workspace = response.json()
    workspace_id = workspace['id']
    print(f'工作空间创建成功! ID: {workspace_id}')
else:
    print('Response:', response.text[:500])
    exit(1)

# 4. 获取工作空间列表
print('\n=== 4. 获取工作空间列表 ===')
response = requests.get('http://localhost:8001/api/v1/workspaces', headers=headers)
print('Status:', response.status_code)
if response.status_code == 200:
    workspaces = response.json()
    print(f'工作空间数量: {len(workspaces)}')
    for ws in workspaces:
        print(f'  - {ws["name"]} (slug: {ws["slug"]})')

# 5. 创建项目
print('\n=== 5. 创建项目 ===')
response = requests.post(f'http://localhost:8001/api/v1/projects?workspace_id={workspace_id}', headers=headers, json={
    'name': 'Test Project',
    'identifier': 'TP'
})
print('Status:', response.status_code)
if response.status_code == 201:
    project = response.json()
    project_id = project['id']
    print(f'项目创建成功! ID: {project_id}')
else:
    print('Response:', response.text[:300])
    project_id = None

# 6. 获取项目列表
print('\n=== 6. 获取项目列表 ===')
response = requests.get(f'http://localhost:8001/api/v1/projects?workspace_id={workspace_id}', headers=headers)
print('Status:', response.status_code)
if response.status_code == 200:
    projects = response.json()
    print(f'项目数量: {len(projects)}')

# 7. 获取单个项目
print('\n=== 7. 获取单个项目 ===')
if project_id:
    response = requests.get(f'http://localhost:8001/api/v1/projects/{project_id}', headers=headers)
    print('Status:', response.status_code)
    if response.status_code == 200:
        print('获取项目成功!')
        print('项目详情:', response.json())

print('\n=== 测试完成 ===')
