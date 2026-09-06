/** Core HTTP client with typed error handling. */

export const DEFAULT_BASE_URL = "http://localhost:8000/api/v1";
const DEFAULT_TIMEOUT_MS = 30_000;
const LONG_TIMEOUT_MS = 300_000; // 5 min for AI/agent ops

export class APIError extends Error {
  constructor(
    public statusCode: number,
    public message: string,
    public body: Record<string, unknown> = {},
  ) {
    super(`api error ${statusCode}: ${message}`);
    this.name = "APIError";
  }
}

export class HTTPClient {
  private baseUrl: string;
  private token: string;

  constructor(baseUrl = "", token = "") {
    this.baseUrl = (baseUrl || DEFAULT_BASE_URL).replace(/\/+$/, "");
    this.token = token;
  }

  getBaseURL(): string {
    return this.baseUrl;
  }

  private async do<T>(
    method: string,
    path: string,
    opts: {
      query?: Record<string, unknown>;
      body?: unknown;
      timeout?: number;
    } = {},
  ): Promise<T> {
    const url = new URL(`${this.baseUrl}${path}`);
    if (opts.query) {
      for (const [k, v] of Object.entries(opts.query)) {
        if (v !== undefined && v !== null && v !== "" && v !== 0) {
          url.searchParams.set(k, String(v));
        }
      }
    }

    const controller = new AbortController();
    const timeout = setTimeout(
      () => controller.abort(),
      opts.timeout ?? DEFAULT_TIMEOUT_MS,
    );

    try {
      const resp = await fetch(url.toString(), {
        method,
        headers: {
          Authorization: `Bearer ${this.token}`,
          ...(opts.body ? { "Content-Type": "application/json" } : {}),
        },
        body: opts.body ? JSON.stringify(opts.body) : undefined,
        signal: controller.signal,
      });

      if (!resp.ok) {
        let body: Record<string, unknown> = {};
        try {
          body = (await resp.json()) as Record<string, unknown>;
        } catch {
          // ignore
        }
        const msg =
          (body.message as string) ?? (await resp.text()).trim();
        throw new APIError(resp.status, msg, body);
      }

      if (resp.status === 204 || resp.headers.get("content-length") === "0") {
        return undefined as T;
      }
      return (await resp.json()) as T;
    } finally {
      clearTimeout(timeout);
    }
  }

  async getJSON<T>(
    path: string,
    query?: Record<string, unknown>,
    timeout?: number,
  ): Promise<T> {
    return this.do<T>("GET", path, { query, timeout });
  }

  async postJSON<T>(
    path: string,
    body?: unknown,
    query?: Record<string, unknown>,
    timeout?: number,
  ): Promise<T> {
    return this.do<T>("POST", path, { body, query, timeout });
  }

  async putJSON<T>(
    path: string,
    body?: unknown,
    query?: Record<string, unknown>,
    timeout?: number,
  ): Promise<T> {
    return this.do<T>("PUT", path, { body, query, timeout });
  }

  async deleteJSON<T>(
    path: string,
    query?: Record<string, unknown>,
    timeout?: number,
  ): Promise<T> {
    return this.do<T>("DELETE", path, { query, timeout });
  }

  async getRaw(
    path: string,
    query?: Record<string, unknown>,
  ): Promise<Response> {
    const url = new URL(`${this.baseUrl}${path}`);
    if (query) {
      for (const [k, v] of Object.entries(query)) {
        if (v !== undefined && v !== null && v !== "" && v !== 0) {
          url.searchParams.set(k, String(v));
        }
      }
    }
    return fetch(url.toString(), {
      headers: { Authorization: `Bearer ${this.token}` },
    });
  }
}
