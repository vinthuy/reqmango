# awesome-selfhosted PR draft

> Status: **not submitted**. The list is a third-party repo, so the PR has to come from your GitHub account. This file is the exact diff and the PR text.
>
> Checked against `awesome-selfhosted/awesome-selfhosted` `master` on 2026-10-25. Re-read the section before opening the PR; the list moves.

## Where it goes

File: `README.md`  
Section: `### Software Development - Project Management`  
Order is alphabetical. Insert **after Redmine, before Review Board**.

There is no public demo instance, so do not add a `Demo` link. The project URL is the docs site.

## The line

```markdown
- [Reqmango](https://vinthuy.github.io/reqmango/) - Self-hosted project management that triages new requests before they reach the backlog, with custom issue types, workflows and automation. ([Source Code](https://github.com/vinthuy/reqmango)) `MIT` `Docker/Go/Nodejs`
```

Neighbours, so the insertion point is obvious:

```markdown
- [Redmine](https://www.redmine.org/) - Flexible project management web application. ([Source Code](https://svn.redmine.org/redmine/)) `GPL-2.0` `Ruby`
- [Reqmango](https://vinthuy.github.io/reqmango/) - Self-hosted project management that triages new requests before they reach the backlog, with custom issue types, workflows and automation. ([Source Code](https://github.com/vinthuy/reqmango)) `MIT` `Docker/Go/Nodejs`
- [Review Board](https://www.reviewboard.org/) - Extensible and friendly code review tool for projects and companies of all sizes. ([Demo](https://demo.reviewboard.org/), [Source Code](https://github.com/reviewboard/reviewboard)) `MIT` `Python/Docker`
```

## PR

**Title:** Add Reqmango to Software Development - Project Management

**Body:**

```markdown
## Description

Adds [Reqmango](https://github.com/vinthuy/reqmango), a MIT-licensed self-hosted project manager. Intake triages new requests (type, priority, likely duplicates) before they enter the backlog. Issue types, workflows and automation are configurable. Optional AI stays inside Issue / Intake / Cycle; it needs the operator's own API key and is off by default.

- Docs: https://vinthuy.github.io/reqmango/
- Source: https://github.com/vinthuy/reqmango
- License: MIT
- Install: Docker Compose, published images for linux/amd64 and linux/arm64 (`v0.1.0`)

No public demo is linked because there isn't a hosted instance.

## Checklist

- [x] Entry is in the existing "Software Development - Project Management" section
- [x] Alphabetical (between Redmine and Review Board)
- [x] License tag matches the repository (`MIT`)
- [x] Source code link is the canonical GitHub repository
```

## How to send it

```bash
gh repo fork awesome-selfhosted/awesome-selfhosted --clone
cd awesome-selfhosted
git checkout -b add-reqmango
# edit README.md as above
git add README.md
git commit -m "Add Reqmango under Software Development - Project Management"
git push -u origin add-reqmango
gh pr create --repo awesome-selfhosted/awesome-selfhosted --title "Add Reqmango to Software Development - Project Management" --body-file pr-body.md
```

Reviews there are slow and picky about wording. If they ask for a shorter description, cut it to the first sentence and keep the links.
