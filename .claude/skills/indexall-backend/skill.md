---
name: indexall-backend
description: IndexAll backend development standards. Use when editing files in packages/server/** for database operations, API development, tRPC routers, Drizzle ORM, SQLite queries, FTS5 search, DAG operations, or resource/tag management.
---

# IndexAll Backend Development Standards

## Pre-Development Checklist

- Database operations → Read `MODEL.md` first (table structure, indexes, key queries, FTS5 triggers)
- API interfaces → Read `API_DESIGN.md` first (input/output definitions, business rules)

## Database Conventions

- **ORM**: Drizzle (SQLite dialect), schema defined in server package
- **Tag DAG relations**: Adjacency list storage, query subtree via recursive CTE
- **Full-text search**: SQLite FTS5, auto-sync via AFTER INSERT/UPDATE/DELETE triggers
- **DAG operations MUST do cycle detection** (recursive CTE to check ancestor chain)
- **`source + externalId`** is the composite unique identifier for resources (manually added resources have externalId as NULL)
- Tag names and aliases are **globally unique** (deduplicate across tag names and alias tables)
- Time fields stored as ISO 8601 TEXT

## API Conventions

- tRPC on Hono, routers divided by domain: `tag.*`, `resource.*`
- Tag additions/removals via `resource.addTag` / `resource.removeTag`, **NOT through** `resource.update`
- Resource's `source` / `externalId` are **immutable** after creation
- When deleting tag: cascade delete aliases, DAG relations, resource associations; **do NOT delete** child tags or associated resources themselves
- When deleting resource: cascade delete resource-tag associations, FTS5 auto-cleaned by triggers

## Post-Development Maintenance

- Add/modify table structure or indexes → Update `MODEL.md`
- Add/modify tRPC procedure → Update `API_DESIGN.md` (input/output + business rules)
