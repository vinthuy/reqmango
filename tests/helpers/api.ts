import { type APIRequestContext } from '@playwright/test';

const BASE_URL = 'http://localhost:8000/api/v1';

export async function apiLogin(request: APIRequestContext, email: string, password: string) {
  const res = await request.post(`${BASE_URL}/auth/login`, {
    data: { email, password },
  });
  const body = await res.json();
  return body.token;
}

export async function apiRegister(request: APIRequestContext, email: string, password: string, name: string) {
  const res = await request.post(`${BASE_URL}/auth/register`, {
    data: { email, password, name },
  });
  return res.json();
}

export async function apiCreateWorkspace(request: APIRequestContext, token: string, name: string, slug: string) {
  const res = await request.post(`${BASE_URL}/workspaces`, {
    headers: { Authorization: `Bearer ${token}` },
    data: { name, slug },
  });
  return res.json();
}

export async function apiCreateProject(request: APIRequestContext, token: string, workspaceId: number, data: { name: string; identifier: string; description: string }) {
  const res = await request.post(`${BASE_URL}/projects?workspace_id=${workspaceId}`, {
    headers: { Authorization: `Bearer ${token}` },
    data,
  });
  return res.json();
}
