import type { APIRequestContext } from '@playwright/test'

const API = 'http://localhost:8000/api/v1'
const ADMIN = { email: 'admin@reqmango.com', password: 'demo1234' }

export type DemoContext = {
  token: string
  headers: Record<string, string>
  workspaceId: number
  workspaceSlug: string
  projectId: number
  projectIdentifier: string
}

/** Resolve the seeded reqmango-dev / CORE project (IDs vary across DBs). */
export async function resolveDemo(request: APIRequestContext): Promise<DemoContext> {
  const login = await request.post(`${API}/auth/login`, { data: ADMIN })
  if (!login.ok()) {
    throw new Error(`login failed: ${login.status()} ${await login.text()}`)
  }
  const { access_token: token } = await login.json()
  const headers = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }

  const wsRes = await request.get(`${API}/workspaces`, { headers })
  const wsBody = await wsRes.json()
  const workspaces = Array.isArray(wsBody) ? wsBody : (wsBody.data || [])
  const workspace = workspaces.find((w: any) => w.slug === 'reqmango-dev') || workspaces[0]
  if (!workspace?.id) throw new Error('no workspace found for e2e')

  const projRes = await request.get(`${API}/projects?workspace_id=${workspace.id}`, { headers })
  const projBody = await projRes.json()
  const projects = Array.isArray(projBody) ? projBody : (projBody.data || [])
  const project = projects.find((p: any) => p.identifier === 'CORE') || projects[0]
  if (!project?.id) throw new Error('no project found for e2e')

  return {
    token,
    headers,
    workspaceId: workspace.id,
    workspaceSlug: workspace.slug,
    projectId: project.id,
    projectIdentifier: project.identifier || 'CORE',
  }
}

export async function loginPage(page: any, token: string) {
  await page.goto('http://localhost:5173/login')
  await page.evaluate((t: string) => localStorage.setItem('token', t), token)
}
