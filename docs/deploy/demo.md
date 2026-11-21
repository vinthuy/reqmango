# Public demo playground

Tracking: [GitHub #38](https://github.com/vinthuy/reqmango/issues/38).

This is a **shared sandbox**, not production. Prefer a small VPS (~$4–6/mo) running stock Docker Compose + nightly volume wipe.

## Goals

- Visitors open a URL and sign in with published demo credentials
- Data resets every night so the instance stays clean
- No real `AI_API_KEY`; tight rate limits; registration blocked at the proxy

## Host choice (v1)

| Option | Verdict |
|--------|---------|
| **Single VPS + Compose** | **Recommended** — matches `docker-compose.yml` + GHCR images |
| Railway / Render / Fly | More cost/setup for three services + Postgres |

## Server setup

1. Ubuntu 22.04+, 2 GB RAM; open **80/443** only (not 5432/8000).
2. Install Docker Engine + Compose plugin.
3. Clone or copy `docker-compose.yml` + `.env` (from `.env.example`).

### `.env` for the playground

```bash
SEED_DEMO=true
SEED_E2E=false
DEBUG=false
SECRET_KEY=$(openssl rand -hex 32)
DB_PASSWORD=<strong-password>
# leave AI_* empty
RATE_LIMIT_REQUESTS=100
REQMANGO_TAG=0.2.0   # pin a release; avoid floating latest for demos you care about
```

4. Harden ports (see [production.md](./production.md) §3): bind DB/API to loopback; only frontend (or Caddy) public.
5. TLS: Caddy reverse-proxy to `localhost:80`.
6. Optional: block `POST /api/v1/auth/register` at Caddy.

```bash
REQMANGO_TAG=0.2.0 docker compose pull
docker compose up -d
```

Login: `demo@example.com` / `demo1234` (from `SEED_DEMO`).

## Nightly reset

Demo seed is **idempotent** — restart alone does **not** wipe data. Reset = destroy the Postgres volume:

```cron
0 3 * * * cd /opt/reqmango && docker compose down -v && docker compose up -d
```

## README checklist (when URL is live)

- [ ] Add badge + link above the fold in `README.md` / `README-zh.md`
- [ ] Comment on #38 with the public URL and reset schedule
- [ ] Banner in UI or README: “Resets nightly; shared sandbox”

## Maintainer notes

- Do **not** use this checklist for a real customer deployment — turn `SEED_DEMO` **off** there ([production.md](./production.md) §2).
- Oracle Always Free / Hetzner CX22 / Vultr are fine starting points.
