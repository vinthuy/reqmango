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

describe("Cycles", () => {
  it("listCycles decodes wrapped shape", async () => {
    const srv = await mockServer((_req, res) => {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        items: [{ id: 1, name: "Sprint 1", status: "active", progress: 40.5 }],
        total: 1, limit: 50, offset: 0,
      }));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "t" });
      const result = await client.listCycles(5, { status: "active" });
      expect(result.total).toBe(1);
      expect(result.items).toHaveLength(1);
      expect(result.items[0].progress).toBe(40.5);
    } finally {
      srv.close();
    }
  });

  it("getCycle decodes single object", async () => {
    const srv = await mockServer((_req, res) => {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ id: 3, name: "Sprint 1", status: "active" }));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "t" });
      const cycle = await client.getCycle(3);
      expect(cycle.id).toBe(3);
      expect(cycle.name).toBe("Sprint 1");
    } finally {
      srv.close();
    }
  });

  it("getCycleBurndown decodes daily points", async () => {
    const srv = await mockServer((_req, res) => {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        cycle_id: 3, cycle_name: "S1", total_issues: 10, is_on_track: true,
        daily_points: [{ day_index: 0, actual_remaining: 9.0 }],
      }));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "t" });
      const b = await client.getCycleBurndown(3);
      expect(b.is_on_track).toBe(true);
      expect(b.daily_points).toHaveLength(1);
    } finally {
      srv.close();
    }
  });

  it("addIssueToCycle sends POST", async () => {
    const srv = await mockServer((req, res) => {
      expect(req.method).toBe("POST");
      expect(req.url).toContain("issue_id=11");
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ cycle_id: 3, issue_id: 11, action: "added" }));
    });
    try {
      const client = new ReqMangoClient({ baseUrl: srv.url, token: "t" });
      await client.addIssueToCycle(3, 11); // should not throw
    } finally {
      srv.close();
    }
  });
});
