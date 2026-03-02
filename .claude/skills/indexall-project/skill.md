---
name: indexall-project
description: IndexAll project context - unified resource indexing platform. Use when working on the IndexAll codebase. Provides project positioning, tech stack, design document references, and document maintenance rules. Triggers for architecture, database, API, or UI work on IndexAll.
---

# IndexAll - Unified Resource Indexing Platform

## Overview

Cross-system resource indexer (not a bookmark manager). Uses DAG tag system to index resources from all platforms (webpages, GitHub repos, Notion pages, etc.). Search and browse in this platform, jump to original pages with one click.

**Current Phase: Phase 1 MVP** - Covers basic loop of "classify resources + search + jump".

## Tech Stack

| Layer | Choice |
|---|---|
| Monorepo | pnpm workspaces |
| Frontend | React + Vite + TailwindCSS + shadcn/ui |
| Backend | Hono + tRPC + Drizzle + SQLite (better-sqlite3) |
| Browser Extension | WXT (React + TypeScript) |
| API Layer | tRPC (type-safe frontend-backend connection) |

## Project Structure

```
indexall/
├── packages/
│   ├── web/          # React frontend (Vite)
│   ├── server/       # Hono backend + tRPC
│   ├── extension/    # WXT browser extension
│   └── shared/       # Shared types, constants
```

## Design Documents Reference

**MUST read the corresponding document using Read tool before coding** in these scenarios:

| Trigger Scenario | Required Document |
|---|---|
| Product decisions, feature scope, Phase planning | `REQUIREMENTS.md` |
| Architecture decisions, tech choices, Connector design | `DESIGN.md` |
| Database schema, SQL queries, table structure changes | `MODEL.md` |
| tRPC router, API interface definition and implementation | `API_DESIGN.md` |
| UI components, page layout, interactions, responsive design | `UI_DESIGN.md` |

When uncertain if a feature is in current Phase scope, read `REQUIREMENTS.md` first to confirm.

## Document Maintenance Rules

**MUST synchronously update corresponding document after coding** when code changes involve:

| Code Change Type | Update Document |
|---|---|
| Add/modify/delete database tables or fields | `MODEL.md` |
| Add/modify/delete tRPC procedure | `API_DESIGN.md` |
| Add/modify UI components, page layout, interaction patterns | `UI_DESIGN.md` |
| Architecture changes, tech stack adjustments, project structure changes | `DESIGN.md` |
| Requirement changes, feature additions/removals | `REQUIREMENTS.md` |

Document update principles:
- Keep documents consistent with code, no outdated information
- Maintain original document style and structure when updating
- If change is large and requires updating multiple documents, update one by one and inform user
