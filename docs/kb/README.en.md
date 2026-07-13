# Knowledge Base

The KB is the **single source of truth** for the reqmango system, always describing the current actual state.

**Last Updated**: 2026-07-13

## Maintenance Principles

- **Timeliness**: When a feature is implemented, the corresponding KB document must be updated synchronously
- **Accuracy**: KB describes what the code actually does, not what was intended during design
- **Conciseness**: KB is a reference document, not a development log. Keep core information, remove procedural content

## KB Update Triggers

| Event | Update Scope |
|-------|--------------|
| New feature completed | Update `architecture/README.en.md` module table + `data-model.en.md` table list + `frontend.en.md` |
| API changes | Update `api-conventions.en.md` |
| Data model changes | Update `data-model.en.md` |
| Tech stack upgrade | Update `architecture/README.en.md` |

## KB Document List

### Product Layer
- [PRD.en.md](PRD.en.md) — Product Requirements Document

### Architecture Layer
- [architecture/README.en.md](architecture/README.en.md) — Architecture overview + module status
- [architecture/project-layout.en.md](architecture/project-layout.en.md) — Project directory structure
- [architecture/backend-go.en.md](architecture/backend-go.en.md) — Go backend architecture
- [architecture/frontend.en.md](architecture/frontend.en.md) — Frontend architecture
- [architecture/data-model.en.md](architecture/data-model.en.md) — Data model
- [architecture/api-conventions.en.md](architecture/api-conventions.en.md) — API design conventions
- [architecture/tech-stack.en.md](architecture/tech-stack.en.md) — Tech stack details

### Feature Design Documents
- [architecture/saved-views-design.en.md](architecture/saved-views-design.en.md) — Saved Views
- [architecture/pages-design.en.md](architecture/pages-design.en.md) — Pages/Wiki
- [architecture/notification-design.en.md](architecture/notification-design.en.md) — Notification system
- [architecture/type-hierarchy-template-design.en.md](architecture/type-hierarchy-template-design.en.md) — Type hierarchy & templates
- [architecture/relation-system-design.en.md](architecture/relation-system-design.en.md) — Relation type system

### Legacy
- [architecture/backend-python.en.md](architecture/backend-python.en.md) — Python backend (deprecated)

### Changelog
- [changelog/README.en.md](changelog/README.en.md) — KB update history

---

## 🌐 Language

- **English** (this document)
- [中文文档](README.md)