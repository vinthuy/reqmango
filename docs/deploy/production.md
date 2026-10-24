# Production deployment

This is the checklist for running Reqmango on a server that other people can
reach. It assumes Docker Compose, but the settings apply to a bare binary as
well. For the local demo, the README quickstart is enough — this page is about
not getting owned.

Related: [security & operations](../../docs/security.md),
[security policy](../../SECURITY.md).

---

## 0. Pre-flight checklist

- [ ] `SEED_DEMO=false` (no published demo accounts)
- [ ] `SECRET_KEY` is a fresh 64-character random value, stored outside git
- [ ] `DB_PASSWORD` is not `reqmango`
- [ ] `ADMIN_PASSWORD` is set, or the generated one was captured from the log
- [ ] PostgreSQL and the backend are **not** published to the internet
- [ ] TLS terminates in front of the app
- [ ] Backups are scheduled **and** a restore has been tested once

---

## 1. Secrets

`SECRET_KEY` signs access tokens. If it is missing, the backend generates a
random key per process and prints a warning; every restart then invalidates all
sessions. Generate it once and keep it:

```bash
openssl rand -hex 32
```

Store it in `.env` (git-ignored), a Docker secret, or your orchestrator's secret
store. Rotating it logs everybody out, which is the correct behaviour after a
suspected leak.

Also change `DB_PASSWORD`. The Compose default (`reqmango`) is only acceptable
on a loopback-only database.

---

## 2. Turn the demo data off

`SEED_DEMO=true` creates sample workspaces plus the accounts advertised in the
README (`admin@reqmango.com`, `demo@example.com`, password `demo1234`). Those
passwords are in the public repository, so an instance with this flag on is an
instance anyone can administer.

```dotenv
SEED_DEMO=false
SEED_E2E=false
```

With the demo off, the first boot on an empty database creates a single
superuser:

| Setting | Effect |
|---------|--------|
| `ADMIN_EMAIL` (default `admin@reqmango.com`) | Login address for the bootstrap admin |
| `ADMIN_PASSWORD` set | That exact password is used |
| `ADMIN_PASSWORD` empty | A random 24-character password is generated and printed **once** to the backend log |

The bootstrap only runs when no superuser exists, so restarts never recreate or
modify it. It also refuses to promote an already-registered account: if a user
with `ADMIN_EMAIL` already exists and no superuser does, the server logs an
instruction to pick another address instead of silently granting admin to
whoever signed up first.

Capture the generated password:

```bash
docker compose logs backend | grep -A6 "Bootstrap superuser created"
```

Change it after the first sign-in (account settings → password).

> Already deployed an older version? An instance that has been running with the
> old defaults still contains `admin@reqmango.com` / `demo1234` and
> `qa_tester@reqmango.com` / `Test@12345`. Changing `SEED_DEMO` does not delete
> existing rows: sign in and change those passwords, or delete the accounts if
> they are not needed.

---

## 3. Network exposure

The stock `docker-compose.yml` publishes PostgreSQL (`5432`) and the API
(`8000`) on the host. That is convenient for development and wrong for a public
server. Either bind them to loopback:

```yaml
ports:
  - "127.0.0.1:5432:5432"   # db
  - "127.0.0.1:8000:8000"   # backend
```

…or drop the backend port mapping entirely and let the frontend container (or
your proxy) reach it over the Compose network. Only the HTTP(S) entry point
should be reachable from outside.

### TLS

Terminate TLS in a reverse proxy. Caddy gets a certificate automatically:

```caddyfile
reqmango.example.com {
    encode zstd gzip
    reverse_proxy frontend:80
}
```

With nginx, forward the usual headers and **disable buffering for the event
stream** (`/api/v1/sse/*`), otherwise notifications arrive in batches:

```nginx
location / {
    proxy_pass http://frontend:80;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 3600s;
}

location /api/v1/sse/ {
    proxy_pass http://frontend:80;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_buffering off;
    proxy_read_timeout 3600s;
}
```

Set a generous read timeout: SSE connections are long-lived.

### Open registration

`POST /api/v1/auth/register` is public. If the instance is not meant to accept
sign-ups from the internet, block that path at the proxy or keep the whole app
behind your SSO/VPN.

---

## 4. Backups

Everything lives in PostgreSQL. Take a logical dump on a schedule:

```bash
docker compose exec -T db pg_dump -U "$DB_USER" "$DB_NAME" \
  | gzip > "reqmango-$(date +%F).sql.gz"
```

Restore into a fresh database and start the backend against it:

```bash
gunzip -c reqmango-2026-10-24.sql.gz | \
  docker compose exec -T db psql -U "$DB_USER" -d "$DB_NAME"
```

Keep dumps off the host (object storage or another machine), and restore one
into a scratch database periodically — an untested backup is a guess.

---

## 5. Upgrades

1. Back up first (see above).
2. Pick a release tag — **pin it**, don't float on `latest` in production:

   ```bash
   REQMANGO_TAG=0.2.0 docker compose pull
   REQMANGO_TAG=0.2.0 docker compose up -d
   ```

   Building from a checkout (`git pull && docker compose up -d --build`) is
   equivalent, just slower.
3. Confirm the running build, then watch the logs:

   ```bash
   curl -s localhost/version
   docker compose logs -f backend
   ```

On boot the server applies `AutoMigrate` and the idempotent SQL baseline. If you
prefer an explicit migration step (`cmd/migrate`), see the README — the two
paths converge on the same schema, and `cmd/migrate up` is safe to re-run.

Read `CHANGELOG.md` (or the GitHub release notes) before a minor bump for
configuration changes.

---

## 6. Health, metrics, logs

| Endpoint | Use |
|----------|-----|
| `GET /health` | liveness; container healthcheck |
| `GET /readyz` | readiness; load balancer / rollout gate (pings the database) |
| `GET /version` | build identity — version, commit, build date, Go version |
| `GET /api/v1/metrics` | in-process request metrics; **superuser only** |

`DB_LOG_LEVEL=warn` keeps SQL parameters out of the logs. Requests slower than
`SLOW_REQUEST_MS` (default 1000) are logged as `[SLOW]` with the route template,
not the concrete URL.

---

## 7. Hardening checklist (optional but cheap)

- [ ] `RATE_LIMIT_REQUESTS` tuned to your traffic (default 500 / 60 s per IP)
- [ ] Security headers at the proxy (`Strict-Transport-Security`,
      `X-Content-Type-Options`, `Referrer-Policy`, a CSP you have tested)
- [ ] Restrict outbound network access if you do not use AI providers, webhooks
      or MCP runtimes
- [ ] Rotate `SECRET_KEY`, `DB_PASSWORD` and AI provider keys on a schedule
- [ ] Watch for releases: security fixes land on `master` and in release notes
