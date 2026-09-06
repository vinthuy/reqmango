import { describe, it, expect, beforeAll, afterAll } from "vitest";
import { createServer, type Server } from "node:http";
import { ReqMangoClient, APIError } from "../src/index.js";

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

describe("HTTPClient", () => {
  it("getJSON decodes 200 response", async () => {
    const srv = await mockServer((_req, res) => {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([{ id: 1, name: "Acme", slug: "acme" }]));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "test" });
      const ws = await client.listWorkspaces();
      expect(ws).toHaveLength(1);
      expect(ws[0].name).toBe("Acme");
    } finally {
      srv.close();
    }
  });

  it("throws APIError on 401", async () => {
    const srv = await mockServer((_req, res) => {
      res.writeHead(401, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ message: "token expired" }));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "bad" });
      await expect(client.listWorkspaces()).rejects.toThrow(APIError);
      try {
        await client.listWorkspaces();
      } catch (e: any) {
        expect(e.statusCode).toBe(401);
        expect(e.message).toContain("token expired");
      }
    } finally {
      srv.close();
    }
  });

  it("throws APIError on 409 with body", async () => {
    const srv = await mockServer((_req, res) => {
      res.writeHead(409, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ message: "approval_required", transition_id: 9 }));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "test" });
      try {
        await client.updateIssue(1, { name: "new" });
        expect.fail("should have thrown");
      } catch (e: any) {
        expect(e.statusCode).toBe(409);
        expect(e.body.transition_id).toBe(9);
      }
    } finally {
      srv.close();
    }
  });

  it("postJSON sends body and decodes 201", async () => {
    const srv = await mockServer((req, res) => {
      let body = "";
      req.on("data", (chunk: any) => (body += chunk));
      req.on("end", () => {
        const parsed = JSON.parse(body);
        expect(parsed.name).toBe("Bug");
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ id: 11, name: "Bug", sequence_id: 42 }));
      });
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "test" });
      const issue = await client.createIssue(5, 2, { name: "Bug" });
      expect(issue.id).toBe(11);
      expect(issue.sequence_id).toBe(42);
    } finally {
      srv.close();
    }
  });

  it("deleteJSON sends DELETE", async () => {
    const srv = await mockServer((req, res) => {
      expect(req.method).toBe("DELETE");
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ message: "revoked" }));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "test" });
      await client.revokePAT(3); // should not throw
    } finally {
      srv.close();
    }
  });
});
