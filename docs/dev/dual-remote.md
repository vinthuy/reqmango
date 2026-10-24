# Dual remotes (maintainers)

**One product, one narrative, two remotes.**

| Role | Host | Audience | Use for |
|------|------|----------|---------|
| **Primary (国内共建主场)** | [GitCode](https://gitcode.com/yongfeng9m-/reqmanpy) (`origin`) | 中文开发者、国内社区、日常 Issue/PR | 讨论、评审、发版说明（中文优先） |
| **Mirror (国际镜像)** | [GitHub](https://github.com/vinthuy/reqmango) (`github`) | 国际可见度、英文检索、海外贡献者 | Topics / Stars / 英文 Issue；保持与 `origin/master` 同步 |

Do **not** run two roadmaps, two READMEs-with-different-stories, or two default feature sets.

## Product sentence (same on both)

- **ZH:** 自建项目管理：新需求先分诊再进待办；可自定义工作项类型 / 工作流 / 自动化；AI 嵌在 Issue / Intake / Cycle，数据在自己的机器上。
- **EN:** Self-hosted project management: triage first; custom issue types, workflows, and automation; AI in Issue / Intake / Cycle — data stays on your machine.

## Sync workflow

After merging to `master` locally:

```bash
git checkout master
git push origin master    # GitCode first
git push github master    # GitHub mirror
```

Optional helper (PowerShell):

```powershell
git push origin master; if ($LASTEXITCODE -eq 0) { git push github master }
```

The `gitee` remote is a read-only convenience mirror for users who cannot reach
GitHub. Push `master` there after the two above; it runs no CI, so a green
GitHub run is the signal, not Gitee.

Tag releases on both remotes with the **same** tag name and changelog (ZH + short EN).

## Where contributors should open PRs

| Contributor | Prefer |
|-------------|--------|
| 国内 / 中文沟通 | GitCode Issue + PR → `origin` |
| International / English | GitHub Issue + PR → mirror; maintainers cherry-pick or merge then sync both |

If the same fix lands on both sides, reconcile on `master` once, then push both remotes — never diverge product scope.

## Branch protection on the GitHub mirror

`master` on GitHub is protected. GitCode/Gitee have no equivalent gate, so the
GitHub mirror is where a red build is actually enforced.

| Rule | Value | Why |
|------|-------|-----|
| Required checks | `Lint`, `Security`, `Test`, `E2E`, `Tenancy audit`, `Migration path`, `Build`, `Docker Build (backend)`, `Docker Build (frontend)` | Must match the job names in `ci.yml` **verbatim** |
| Force push / branch deletion | blocked | History is public and mirrored |
| Admin bypass | allowed | Single maintainer; a hotfix can land directly, PRs still need green checks |
| Unresolved conversations block merge | yes | Review comments should not be silently ignored |
| Require branch up to date | no | Otherwise every PR has to be rebased on each merge |

The two `Docker Build` entries come from the job's matrix, not from the job
name. Renaming `docker:` or editing `matrix.component` silently disables half
the gate — update the rule in the same change:

```bash
# required-checks.json: {"strict": false, "contexts": ["Lint", ...]}
gh api -X PATCH /repos/vinthuy/reqmango/branches/master/protection/required_status_checks \
  --input required-checks.json
```

## GitHub About / Topics

Mirror description and Topics must match the product sentence above. See [github-public-profile.md](./github-public-profile.md). Apply the same blurb on GitCode About when the UI allows.

## Naming note

GitCode path may still show a historical repo slug (`reqmanpy`). Prefer displaying **Reqmango** in titles and README; rename on GitCode when convenient so clone URLs match the product name.
