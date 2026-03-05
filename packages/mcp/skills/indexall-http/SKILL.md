---
name: indexall-http
description: "Manage and organize your resource collection using IndexAll via direct HTTP calls (no MCP required). Use when: (1) user wants to know what resources they have saved, (2) user needs to index a new resource (URL, article, local file, tool, etc.), (3) organizing tags, searching resources, or maintaining knowledge base — and MCP tools are unavailable. Makes HTTP calls directly to the IndexAll backend and filesystem connector."
---

# IndexAll Assistant (HTTP Mode)

Direct HTTP version of IndexAll Assistant. No MCP required — all operations via HTTP calls.

## Service URLs

- **Backend**: `http://localhost:8080` (or `$INDEXALL_API_URL`)
- **Filesystem Daemon**: `http://localhost:47832` (or `$INDEXALL_DAEMON_URL`)

---

## Quick Start

### Find Resources
1. Ask what you're looking for
2. Search by tag or keyword via `POST /v1/resources/query`
3. Refine as needed

### Index a New Resource
- **Local file**: `POST :47832/index { "path": "/abs/path" }` → then tag it
- **URL/Web resource**: `POST :8080/v1/resources { "title", "url", "source" }` → then tag it

### Organize Tags
1. `GET /v1/tags/tree` to see current structure
2. Create/restructure tags as needed
3. Link resources to tags

---

## Core Workflows

### Workflow 1: Index New Resources

**Goal**: Add new resources and categorize them properly

**Steps**:
1. Check if already indexed: `GET /v1/resources/by-url?url=<encoded>` or search by keyword
2. Create the resource (**choose by type**):
   - **Local file**: `POST http://localhost:47832/index` with `{ "path": "/absolute/path" }`
     → ⚠️ Do NOT use `create_resource` for local files — daemon manages `source`/`external_id` format, bypassing it breaks move/deletion tracking
   - **Web URL / GitHub / Notion / etc.**: `POST /v1/resources` with `{ "title", "url", "source": "manual", "description" }`
3. Assign tags: `POST /v1/resources/{id}/tags` with `{ "tag_id": "<id>" }`
4. Verify: `POST /v1/resources/query` to confirm it's discoverable

**Best for**: Importing bookmarks, indexing files, adding research materials

---

### Workflow 2: Find Resources

**Goal**: Find resources by topic, keyword, or category

**By tag** (get everything under a category):
```
POST /v1/resources/query
{ "tag_query": { "tag_id": "<id>", "tag_scope": "WITH_DESCENDANTS" }, "page": 1, "page_size": 20 }
```

**By keyword** (full-text search):
```
POST /v1/resources/query
{ "keyword_query": { "keyword": "<query>", "field_scope": "ALL", "tag_scope": "WITH_ANCESTORS" } }
```

**Steps**:
1. If you have a topic but not a tag ID: `GET /v1/tags/search?query=<topic>&tag_scope=WITH_ANCESTORS`
2. Use the tag ID for precise filtering, or keyword for broad search
3. Examine details: `GET /v1/resources/{id}`

**`tag_scope` meanings**:
- `WITH_DESCENDANTS` — this tag + all subtopics (use when browsing a category)
- `WITH_ANCESTORS` — this tag + all parent contexts (use when keyword searching)
- `DIRECT` — exact tag only

---

### Workflow 3: Refactor Tag Taxonomy

**Goal**: Improve tag hierarchy (restructure, consolidate, add missing tags)

**Steps**:
1. `GET /v1/tags/tree` — review current structure
2. Identify gaps or issues
3. Execute changes:
   - Create: `POST /v1/tags { "name", "color": "#hex", "parent_ids": [] }`
   - Add alias: `POST /v1/tags/{id}/aliases { "alias": "<name>" }`
   - Add parent: `POST /v1/tags/{child_id}/parents { "parent_id": "<id>" }`
   - Remove parent: `DELETE /v1/tags/{child_id}/parents/{parent_id}`
   - Update: `PATCH /v1/tags/{id} { "name", "color" }`
   - Delete: `DELETE /v1/tags/{id}`
4. Verify: search again to confirm resources still discoverable

---

### Workflow 4: Bulk Retag

**Goal**: Fix miscategorized resources or reorganize a batch

**Steps**:
1. `POST /v1/resources/query` to get the batch
2. For each resource: `POST /v1/resources/{id}/tags` or `DELETE /v1/resources/{id}/tags/{tag_id}`
3. Verify final state with another query

---

## API Reference (Quick)

### Resources
```
POST   /v1/resources/query          unified query (tag or keyword)
POST   /v1/resources                create (URL/web resources only)
GET    /v1/resources/{id}
PATCH  /v1/resources/{id}           { title, description, url }
DELETE /v1/resources/{id}
GET    /v1/resources/by-url?url=<encoded>
POST   /v1/resources/{id}/tags      { "tag_id": "<id>" }
DELETE /v1/resources/{id}/tags/{tag_id}
```

### Tags
```
GET    /v1/tags                     list all
GET    /v1/tags/tree                full DAG tree
GET    /v1/tags/search?query=<q>&tag_scope=WITH_ANCESTORS
POST   /v1/tags                     { name, color, parent_ids }
PATCH  /v1/tags/{id}
DELETE /v1/tags/{id}
POST   /v1/tags/{id}/aliases        { "alias": "<name>" }
DELETE /v1/tags/aliases/{alias_id}
POST   /v1/tags/{child_id}/parents  { "parent_id": "<id>" }
DELETE /v1/tags/{child_id}/parents/{parent_id}
```

### Filesystem Connector (local files)
```
POST   http://localhost:47832/index  { "path": "/absolute/path" }
→ { "id": "<resource-id>", "created": true|false }
```

---

## Tips

- **Use `WITH_DESCENDANTS`** when browsing a category to get all subtopics
- **Use `WITH_ANCESTORS`** when keyword searching to catch broader matches
- **Alias tags** with multiple names so fuzzy search finds them (`add_alias "ReactJS"` for "React")
- **Hierarchy should be meaningful**: parent-child = is-a relationship, not just broad-to-narrow
- **`index_local_file` is idempotent**: safe to call multiple times on the same file

## Customization

Your taxonomy is documented in [`references/taxonomy.md`](references/taxonomy.md). Review and update it to:
- Document your tag structure
- Explain why tags are organized as they are
- Define what each branch of the hierarchy is for
- Store examples of resources in each category

This helps me understand your organization and make better suggestions.
