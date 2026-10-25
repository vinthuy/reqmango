# Release process

Releases are tag-driven. Pushing a `v*` tag runs
[`.github/workflows/release.yml`](https://github.com/vinthuy/reqmango/blob/master/.github/workflows/release.yml), which
publishes multi-arch images and creates the GitHub Release.

## One-time setup

- The repository must allow GitHub Actions to write packages:
  **Settings → Actions → General → Workflow permissions → Read and write** (the
  workflow asks for `packages: write`, which is not enough on its own if the
  repository default is read-only).
- On a public repository the GHCR packages inherit public visibility, so
  anonymous `docker pull` works after the first release. Verify:

  ```bash
  curl -s "https://ghcr.io/token?scope=repository:vinthuy/reqmango-backend:pull&service=ghcr.io"
  ```

  A token (rather than `DENIED`) means the package is readable without
  credentials. If you ever make the repository or the package private, restore
  it under **Package settings → Change visibility → Public**.

## Cutting a release

1. Update `CHANGELOG.md`:
   - move entries out of `[Unreleased]` into `## [x.y.z] - YYYY-MM-DD`,
   - keep the **Upgrade notes** section honest about configuration or schema
     changes,
   - update the comparison links at the bottom.
2. Make sure `master` is green on CI (all jobs, including both Docker builds).
3. Tag and push. The tag must be semver, `v`-prefixed:

   ```bash
   git tag -a v0.2.0 -m "Reqmango v0.2.0"
   git push github v0.2.0     # triggers the release workflow
   git push origin v0.2.0     # gitcode mirror
   git push gitee v0.2.0      # gitee mirror
   ```

   Pushing the tag to a mirror does not trigger GitHub Actions; only the GitHub
   push does. Mirrors still get the tag, so their archives carry the same
   version.

4. Watch the workflow. It publishes:

   | Image | Tags |
   |-------|------|
   | `ghcr.io/vinthuy/reqmango-backend` | `0.2.0`, `0.2`, `latest` |
   | `ghcr.io/vinthuy/reqmango-frontend` | `0.2.0`, `0.2`, `latest` |

   A pre-release tag behaves differently. `v0.2.0-rc1` publishes **only**
   `0.2.0-rc1` — no `0.2` and no `latest` (`docker/metadata-action` extends just
   `{{version}}` for pre-releases, because a `0.2` tag pointing at an rc is a
   trap for anyone who pinned the minor). The GitHub Release is marked as a
   pre-release too.

5. Verify the published build:

   ```bash
   REQMANGO_TAG=0.2.0 docker compose pull
   REQMANGO_TAG=0.2.0 docker compose up -d
   curl -s localhost/version        # version should be v0.2.0
   ```

   The workflow injects `VERSION` (the tag, e.g. `v0.1.0`), `COMMIT` (short
   SHA) and `BUILD_DATE` into `internal/version`; a local `go run` reports
   `dev`. Image tags drop the `v` (`0.1.0`), because Docker tags are the semver
   numbers themselves.

## If the workflow fails

The tag is only a pointer: nothing has been published, and the GitHub Release
is created by the final job, so a failed run leaves no public trace.

```bash
git tag -d v0.2.0
git push github :refs/tags/v0.2.0
# fix, commit, then re-tag
```

Re-running the workflow from the Actions UI is also safe: image pushes are
idempotent per tag.

## Rehearsing the workflow

To exercise the publish path without spending a version number, push a
throwaway pre-release tag. It creates every artifact the real thing does
(multi-arch images, attestations, the release) and touches nothing anyone
depends on, because a pre-release only publishes its exact tag:

```bash
git tag -a v0.0.0-rc.1 -m "release workflow dry run"
git push github refs/tags/v0.0.0-rc.1
gh run watch                                    # ~4 min for both images
```

Verify, then tear down:

```bash
gh release view v0.0.0-rc.1 --json isPrerelease,body
gh release delete v0.0.0-rc.1 --yes --cleanup-tag   # release + remote tag
git tag -d v0.0.0-rc.1 2>/dev/null || true          # if a local copy is left
```

The image tags need a token with `delete:packages` (`gh auth refresh -s
delete:packages`), or the package UI: **Package settings → Manage versions →
delete `0.0.0-rc.1`** for both packages. The workflow itself never deletes
anything, and `GITHUB_TOKEN` only has to publish.

## Versioning

While the major version is `0`, minor bumps may contain breaking changes and
are called out in the changelog. Patch releases are fixes only. Once `1.0.0`
ships, normal semver applies and `latest` always points at the newest stable
release.

## What a release does *not* do

- It does not push tags to the mirrors; do that explicitly (step 3).
- It does not deploy anything. Self-hosted operators pull the image themselves;
  see [deploy/production.md](../deploy/production.md#5-upgrades).
