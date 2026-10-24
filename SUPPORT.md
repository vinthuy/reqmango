# Support

Reqmango is maintained by a small team in its spare time. There is no paid
support and no response-time guarantee, but well-scoped reports usually get an
answer within a few days.

## Before asking

1. [README](README.md) — quickstart, requirements, local development, seed
   accounts.
2. [Docs index](docs/README.md) — architecture, API reference, deployment.
3. [Troubleshooting](docs/deploy/production.md) — secrets, TLS/SSE proxying,
   backups, upgrades.
4. Search existing [issues](https://github.com/vinthuy/reqmango/issues) and
   [discussions](https://github.com/vinthuy/reqmango/discussions).

## Where to ask

| Your question | Where |
|---------------|-------|
| "How do I …?" / "Is this supposed to …?" | [Discussions → Q&A](https://github.com/vinthuy/reqmango/discussions/categories/q-a) |
| Reproducible bug | [New issue](https://github.com/vinthuy/reqmango/issues/new/choose) (bug template) |
| Feature idea | [Discussions → Ideas](https://github.com/vinthuy/reqmango/discussions) first; an issue once the shape is clear |
| Security vulnerability | **Never a public issue** — see [SECURITY.md](SECURITY.md) |
| Code of conduct incident | **vinthuy@qq.com** with `[conduct]` in the subject — see [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |

Chinese-language questions are welcome on the [GitCode mirror](https://gitcode.com/yongfeng9m-/reqmanpy);
code and releases stay in sync with GitHub.

## What to include in a bug report

- the output of `curl -s localhost/version` (version, commit, build date),
- how you run it: Compose (image or local build), bare binary, which reverse
  proxy,
- the exact steps, expected result, actual result,
- relevant backend logs (`docker compose logs backend`) and browser console
  errors, with secrets and personal data removed,
- whether AI features are enabled (an `AI_API_KEY` is set).

Reports without a reproduction usually stall: we cannot inspect your instance.

## What we cannot help with

- debugging a modified fork,
- recovering data without a backup,
- infrastructure outside the app (DNS, TLS certificates, Kubernetes specifics),
- proprietary integrations we have no access to.

## Contributing instead

If you can reproduce a bug in a test, or fix it, that is worth more than a
report. See [CONTRIBUTING.md](CONTRIBUTING.md) and the
[good first issues](docs/dev/good-first-issues.md) list.
