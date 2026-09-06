/**
 * reqmango Node.js SDK — client for the reqmango project management API.
 *
 * Zero runtime dependencies (uses native fetch, Node 18+).
 *
 * @example
 * ```ts
 * import { ReqMangoClient } from "@reqmango/sdk";
 *
 * const client = new ReqMangoClient({ token: "reqmango_pat_xxx" });
 * const issues = await client.listIssues({ projectId: 1 });
 * ```
 */

import { APIError, HTTPClient } from "./http.js";
import type {
  Agent,
  AgentActivity,
  AgentTask,
  AIChatReply,
  AISearchResponse,
  BurndownData,
  Chat,
  Comment,
  CreatePATRequest,
  CreatePATResponse,
  CreateIssueRequest,
  Cycle,
  CycleListResult,
  CycleProgress,
  Issue,
  IssueListOptions,
  IssueListResult,
  IssueSearchResult,
  IssueType,
  Label,
  Member,
  Message,
  Notification,
  Page,
  PATResponse,
  Project,
  ProjectCreateRequest,
  State,
  TaskLog,
  TokenResponse,
  UpdateIssueRequest,
  User,
  Workspace,
  WorkspaceCreateRequest,
} from "./types.js";

export { APIError } from "./http.js";
export type * from "./types.js";

export interface ReqMangoClientOptions {
  baseUrl?: string;
  token?: string;
}

export class ReqMangoClient {
  private http: HTTPClient;

  constructor(opts: ReqMangoClientOptions = {}) {
    this.http = new HTTPClient(opts.baseUrl, opts.token);
  }

  get baseURL(): string {
    return this.http.getBaseURL();
  }

  // --------------------------------------------------------------- Auth

  async login(email: string, password: string): Promise<TokenResponse> {
    return this.http.postJSON<TokenResponse>("/auth/login", { email, password });
  }

  async me(): Promise<User> {
    return this.http.getJSON<User>("/auth/me");
  }

  async createPAT(req: CreatePATRequest): Promise<CreatePATResponse> {
    return this.http.postJSON<CreatePATResponse>("/auth/tokens", req);
  }

  async listPATs(): Promise<PATResponse[]> {
    return this.http.getJSON<PATResponse[]>("/auth/tokens");
  }

  async revokePAT(id: number): Promise<void> {
    await this.http.deleteJSON(`/auth/tokens/${id}`);
  }

  // --------------------------------------------------------- Workspaces

  async listWorkspaces(): Promise<Workspace[]> {
    return this.http.getJSON<Workspace[]>("/workspaces");
  }

  async createWorkspace(req: WorkspaceCreateRequest): Promise<Workspace> {
    return this.http.postJSON<Workspace>("/workspaces", req);
  }

  async listMembers(workspaceId: number): Promise<Member[]> {
    return this.http.getJSON<Member[]>(`/workspaces/${workspaceId}/members`);
  }

  // ----------------------------------------------------------- Projects

  async listProjects(workspaceId: number): Promise<Project[]> {
    return this.http.getJSON<Project[]>("/projects", { workspace_id: workspaceId });
  }

  async getProject(projectId: number): Promise<Project> {
    return this.http.getJSON<Project>(`/projects/${projectId}`);
  }

  async createProject(workspaceId: number, req: ProjectCreateRequest): Promise<Project> {
    return this.http.postJSON<Project>("/projects", req, { workspace_id: workspaceId });
  }

  // ------------------------------------------------------------- Issues

  async listIssues(opts: IssueListOptions = {}): Promise<IssueListResult> {
    const query: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(opts)) {
      if (v !== undefined && v !== null && v !== "" && v !== 0) {
        query[k] = v;
      }
    }

    // ListIssues returns bare array + X-Total-Count header
    const resp = await this.http.getRaw("/issues", query);
    if (!resp.ok) {
      let body: Record<string, unknown> = {};
      try { body = (await resp.json()) as Record<string, unknown>; } catch { /* */ }
      throw new APIError(resp.status, (body.message as string) ?? resp.statusText, body);
    }
    const items = (await resp.json()) as Issue[];
    const total = parseInt(resp.headers.get("X-Total-Count") ?? "0", 10);
    return { items, total };
  }

  async createIssue(projectId: number, workspaceId: number, req: CreateIssueRequest): Promise<Issue> {
    return this.http.postJSON<Issue>("/issues", req, {
      project_id: projectId,
      workspace_id: workspaceId,
    });
  }

  async getIssue(issueId: number): Promise<Issue> {
    return this.http.getJSON<Issue>(`/issues/${issueId}`);
  }

  async updateIssue(issueId: number, req: UpdateIssueRequest): Promise<Issue> {
    return this.http.putJSON<Issue>(`/issues/${issueId}`, req);
  }

  async searchIssues(
    workspaceId: number,
    query: string,
    opts: { projectId?: number; limit?: number } = {},
  ): Promise<IssueSearchResult[]> {
    const params: Record<string, unknown> = { workspace_id: workspaceId, query };
    if (opts.projectId) params.project_id = opts.projectId;
    if (opts.limit) params.limit = opts.limit;
    return this.http.getJSON<IssueSearchResult[]>("/issues/search", params);
  }

  async resolveIssueCode(workspaceId: number, code: string): Promise<number> {
    const parts = code.split("-", 2);
    if (parts.length !== 2) {
      throw new Error(`invalid issue code "${code}" (expected IDENTIFIER-NUMBER)`);
    }
    const [identifier, seqStr] = parts;
    const seq = parseInt(seqStr, 10);

    const projects = await this.listProjects(workspaceId);
    const project = projects.find((p) => p.identifier.toUpperCase() === identifier.toUpperCase());
    if (!project) {
      throw new Error(`project with identifier "${identifier}" not found`);
    }

    const result = await this.listIssues({ project_id: project.id, search: seqStr, limit: 100 });
    const match = result.items.find((i) => i.sequence_id === seq);
    if (!match) {
      throw new Error(`issue ${code} not found`);
    }
    return match.id;
  }

  async addComment(issueId: number, body: string, parentId?: number): Promise<Comment> {
    const req: Record<string, unknown> = { issue_id: issueId, body };
    if (parentId) req.parent_id = parentId;
    return this.http.postJSON<Comment>("/comments", req);
  }

  async listComments(
    issueId: number,
    page = 1,
    pageSize = 20,
  ): Promise<{ comments: Comment[]; total: number }> {
    const query: Record<string, unknown> = {};
    if (page > 1) query.page = page;
    if (pageSize !== 20) query.page_size = pageSize;
    return this.http.getJSON<{ comments: Comment[]; total: number }>(
      `/comments/issue/${issueId}`,
      query,
    );
  }

  // ---------------------------------------------------------------- Cycles

  async listCycles(
    projectId: number,
    opts: { status?: string; limit?: number; offset?: number } = {},
  ): Promise<CycleListResult> {
    const query: Record<string, unknown> = {};
    if (opts.status) query.status = opts.status;
    if (opts.limit) query.limit = opts.limit;
    if (opts.offset) query.offset = opts.offset;
    return this.http.getJSON<CycleListResult>(`/projects/${projectId}/cycles`, query);
  }

  async getCycle(cycleId: number): Promise<Cycle> {
    return this.http.getJSON<Cycle>(`/cycles/${cycleId}`);
  }

  async getCycleProgress(cycleId: number): Promise<CycleProgress> {
    return this.http.getJSON<CycleProgress>(`/cycles/${cycleId}/progress`);
  }

  async getCycleBurndown(cycleId: number): Promise<BurndownData> {
    return this.http.getJSON<BurndownData>(`/cycles/${cycleId}/burndown`);
  }

  async addIssueToCycle(cycleId: number, issueId: number): Promise<void> {
    await this.http.postJSON(`/cycles/${cycleId}/issues`, undefined, { issue_id: issueId });
  }

  // ----------------------------------------------------------------- Meta

  async listStates(projectId: number): Promise<State[]> {
    return this.http.getJSON<State[]>(`/projects/${projectId}/settings/states`);
  }

  async listLabels(projectId: number): Promise<Label[]> {
    return this.http.getJSON<Label[]>(`/projects/${projectId}/settings/labels`);
  }

  async listIssueTypes(workspaceId: number, projectId?: number): Promise<IssueType[]> {
    const query: Record<string, unknown> = { workspace_id: workspaceId };
    if (projectId) query.project_id = projectId;
    return this.http.getJSON<IssueType[]>("/issue-types", query);
  }

  async listPages(projectId: number): Promise<Page[]> {
    return this.http.getJSON<Page[]>(`/projects/${projectId}/pages`);
  }

  async getPage(projectId: number, pageId: number): Promise<Page> {
    return this.http.getJSON<Page>(`/projects/${projectId}/pages/${pageId}`);
  }

  async listNotifications(
    opts: { unreadOnly?: boolean; limit?: number; offset?: number } = {},
  ): Promise<Notification[]> {
    const query: Record<string, unknown> = {};
    if (opts.unreadOnly) query.unread_only = "true";
    if (opts.limit) query.limit = opts.limit;
    if (opts.offset) query.offset = opts.offset;
    return this.http.getJSON<Notification[]>("/notifications", query);
  }

  // ------------------------------------------------------------------- AI

  async aiSearch(projectId: number, query: string): Promise<AISearchResponse> {
    return this.http.postJSON<AISearchResponse>(`/projects/${projectId}/ai/search`, { query });
  }

  async aiChat(projectId: number, message: string, threadId?: number): Promise<AIChatReply> {
    const body: Record<string, unknown> = { message };
    if (threadId) body.thread_id = threadId;

    const resp = await fetch(
      `${this.baseURL}/projects/${projectId}/ai/chat`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${(this.http as any).token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
        signal: AbortSignal.timeout(300_000),
      },
    );

    if (!resp.ok) {
      let errBody: Record<string, unknown> = {};
      try { errBody = (await resp.json()) as Record<string, unknown>; } catch { /* */ }
      throw new APIError(resp.status, (errBody.message as string) ?? resp.statusText, errBody);
    }

    const reply: AIChatReply = { text: "", thread_id: 0, tool_calls: [] };
    const reader = resp.body?.getReader();
    if (!reader) return reply;

    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });

      const lines = buffer.split("\n");
      buffer = lines.pop() ?? "";

      for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed || !trimmed.startsWith("data: ")) continue;
        try {
          const ev = JSON.parse(trimmed.slice(6));
          switch (ev.type) {
            case "text":
            case "thinking":
              reply.text += ev.content ?? "";
              break;
            case "tool_call":
              if (ev.tool_call) {
                reply.tool_calls.push(
                  `${ev.tool_call.name ?? ""}(${ev.tool_call.arguments ?? ""})`,
                );
              }
              break;
            case "error":
              throw new APIError(502, ev.error ?? "stream error");
          }
          if (ev.thread_id) reply.thread_id = ev.thread_id;
        } catch (e) {
          if (e instanceof APIError) throw e;
          // ignore non-JSON keepalives
        }
      }
    }

    return reply;
  }

  // --------------------------------------------------------------- Agents

  async listAgents(workspaceId: number): Promise<Agent[]> {
    return this.http.getJSON<Agent[]>(`/workspaces/${workspaceId}/agents`);
  }

  async dispatchAgent(
    workspaceId: number,
    agentId: number,
    task: string,
    issueId?: number,
  ): Promise<AgentActivity> {
    const body: Record<string, unknown> = { task };
    if (issueId) body.issue_id = issueId;
    return this.http.postJSON<AgentActivity>(
      `/workspaces/${workspaceId}/agents/${agentId}/dispatch`,
      body,
      undefined,
      300_000,
    );
  }

  async getAgentTask(workspaceId: number, taskId: number): Promise<AgentTask> {
    return this.http.getJSON<AgentTask>(`/workspaces/${workspaceId}/agent-tasks/${taskId}`);
  }

  async getAgentTaskLogs(workspaceId: number, taskId: number): Promise<TaskLog[]> {
    return this.http.getJSON<TaskLog[]>(
      `/workspaces/${workspaceId}/agent-tasks/${taskId}/logs`,
    );
  }

  // ----------------------------------------------------------------- Chat

  async getIssueChat(issueId: number): Promise<Chat> {
    return this.http.getJSON<Chat>(`/issues/${issueId}/chat`);
  }

  async sendMessage(chatId: number, content: string): Promise<Message> {
    return this.http.postJSON<Message>(`/chats/${chatId}/messages`, { content });
  }
}
