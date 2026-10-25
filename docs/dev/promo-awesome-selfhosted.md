# awesome-selfhosted

> Status: **do not submit yet.** Checked 2026-10-25. Two of their rules block a pull request today, and an agent must not send it.

## Why it is not submitted

Submissions no longer go in the README of `awesome-selfhosted/awesome-selfhosted`. Their pull-request template says to use [awesome-selfhosted-data](https://github.com/awesome-selfhosted/awesome-selfhosted-data) and add `software/reqmango.yml`.

Two checks in that repository's template fail right now:

1. **First release age.** "Any software project you are adding was first released more than 4 months ago." The first tag is `v0.1.0` on 2026-10-24. Their usual reply is to close the pull request and ask for a resubmit once that release is four months old, which is **2027-02-24**.
2. **Human author.** The checklist requires "The submission was done by a human, not a machine/LLM." Machine-written contributions are not allowed and can get the account banned. Open the pull request yourself, from the `vinthuy` GitHub account, and write the description in your own words. Do not paste this file in as the pull request.

There is still no public demo, so leave `demo_url` out.

## File to add after 2027-02-24

Path: `software/reqmango.yml`  
Commit message: `add Reqmango`  
Tag name below is the `name` field of `tags/software-development---project-management.yml`. Platform names match `platforms/docker.yml`, `platforms/go.yml` and `platforms/nodejs.yml`. Put Project Management first: the single-page list shows a project only under its first tag.

```yaml
name: Reqmango
website_url: https://vinthuy.github.io/reqmango/
source_code_url: https://github.com/vinthuy/reqmango
description: Project management that triages new requests before they reach the backlog, with custom issue types, workflows and automation.
licenses:
  - MIT
platforms:
  - Docker
  - Go
  - Nodejs
tags:
  - Software Development - Project Management
```

Rewrite `description` before sending. Keep it one sentence, under 250 characters, and do not say "open-source", "free" or "self-hosted".

Their pull-request checklist also asks you to confirm the project is not already on awesome-sysadmin, that installation instructions work, and that you searched existing issues and pull requests.
