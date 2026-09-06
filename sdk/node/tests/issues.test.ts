import { describe, it, expect } from "vitest";
import { createServer } from "node:http";
import { ReqMangoClient } from "../src/index.js";

function mockServer(handler: (req: any, res: any) => void): Promise<{ url: string; close: () => void }> {
  return new Promise((resolve) => {
    const srv = createServer(handler);
    srv.listen(0, "127.0.0.1", () => {
      const addr = srv.address() as any;
      resolve({
        url: `http://127.0.0.1:${addr.port}/api/v1`,
        close: () => srv.close(),
      });
    });
  });
}

describe("Issues", () => {
  it("listIssues reads X-Total-Count header", async () => {
    const srv = await mockServer((_req, res) => {
      res.writeHead(200, {
        "Content-Type": "application/json",
        "X-Total-Count": "7",
      });
      res.end(JSON.stringify([{ id: 1, name: "bug", sequence_id: 5 }]));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "t" });
      const result = await client.listIssues({ project_id: 1, rql: 'priority = "high"' });
      expect(result.total).toBe(7);
      expect(result.items).toHaveLength(1);
      expect(result.items[0].name).toBe("bug");
    } finally {
      srv.close();
    }
  });

  it("createIssue sends correct path and body", async () => {
    const srv = await mockServer((req, res) => {
      expect(req.url).toContain("project_id=5");
      expect(req.url).toContain("workspace_id=2");
      let body = "";
      req.on("data", (chunk: any) => (body += chunk));
      req.on("end", () => {
        const parsed = JSON.parse(body);
        expect(parsed.name).toBe("Login broken");
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ id: 11, name: "Login broken", sequence_id: 42 }));
      });
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "t" });
      const issue = await client.createIssue(5, 2, { name: "Login broken" });
      expect(issue.id).toBe(11);
      expect(issue.sequence_id).toBe(42);
    } finally {
      srv.close();
    }
  });

  it("searchIssues decodes array", async () => {
    const srv = await mockServer((_req, res) => {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([
        { id: 1, name: "auth bug", sequence_id: 5, project_identifier: "DEMO", project_id: 1 },
      ]));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "t" });
      const results = await client.searchIssues(2, "auth");
      expect(results).toHaveLength(1);
      expect(results[0].project_identifier).toBe("DEMO");
    } finally {
      srv.close();
    }
  });

  it("addComment sends body", async () => {
    const srv = await mockServer((req, res) => {
      let body = "";
      req.on("data", (chunk: any) => (body += chunk));
      req.on("end", () => {
        const parsed = JSON.parse(body);
        expect(parsed.body).toBe("looks good");
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ id: 1, issue_id: 11, body: "looks good", author_id: 1 }));
      });
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "t" });
      const c = await client.addComment(11, "looks good");
      expect(c.id).toBe(1);
      expect(c.body).toBe("looks good");
    } finally {
      srv.close();
    }
  });

  it("listComments decodes wrapped shape", async () => {
    const srv = await mockServer((_req, res) => {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        comments: [{ id: 1, body: "first" }, { id: 2, body: "second" }],
        total: 2,
      }));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "t" });
      const { comments, total } = await client.listComments(11);
      expect(total).toBe(2);
      expect(comments).toHaveLength(2);
      expect(comments[0].body).toBe("first");
    } finally {
      srv.close();
    }
  });
});
