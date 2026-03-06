---
name: indexall-http
description: "Manage and organize your resource collection using IndexAll via direct HTTP calls (no MCP required). Use when: (1) user wants to know what resources they have saved, (2) user needs to index a new resource (URL, article, local file, tool, etc.), (3) organizing tags, searching resources, or maintaining knowledge base — and MCP tools are unavailable. Makes HTTP calls directly to the IndexAll backend and filesystem connector."
---

# IndexAll Assistant (HTTP Mode)

**Backend**: `http://localhost:8080` (or `$INDEXALL_API_URL`)
**Filesystem Daemon**: `http://localhost:47832` (or `$INDEXALL_DAEMON_URL`)

---

## Workflows

### 1. Index New Resources

1. Check if already indexed: `GET /v1/resources/by-url?url=<encoded>` or keyword search
2. Create (**type matters**):
   - **Local file**: `POST :47832/index { "path": "/absolute/path" }`
     ⚠️ Never use `POST /v1/resources` for local files — the daemon manages `source`/`external_id` for move/deletion tracking
   - **Web/URL**: `POST /v1/resources { "title", "url", "source": "manual", "description" }`
3. Find the right tag — **search before creating**:
   - `GET /v1/tags/search?query=<concept>` — try the topic name and synonyms
   - Found → use its `id`; Not found → create one (see Workflow 3)
4. Assign: `POST /v1/resources/{id}/tags { "tag_id": "<id>" }`
5. Verify: `POST /v1/resources/query`

---

### 2. Find Resources

**By tag** (browse a category and all subtopics):
```
POST /v1/resources/query
{ "tag_query": { "tag_id": "<id>", "tag_scope": "WITH_DESCENDANTS" }, "page": 1, "page_size": 20 }
```

**By keyword** (full-text search):
```
POST /v1/resources/query
{ "keyword_query": { "keyword": "<query>", "field_scope": "ALL", "tag_scope": "WITH_ANCESTORS" } }
```

Don't have a tag ID? `GET /v1/tags/search?query=<topic>&tag_scope=WITH_ANCESTORS`

**`tag_scope`**: `WITH_DESCENDANTS` = tag + all subtopics · `WITH_ANCESTORS` = tag + all parents · `DIRECT` = exact tag only

---

### 3. Create or Update Tags (Alias-First)

**⚠️ Always search before creating a tag.**

`GET /v1/tags/search?query=<concept>` — try multiple synonyms first.

| Result | Action |
|--------|--------|
| Exact match | Use it, no changes needed |
| Same concept, different name | Add alias: `POST /v1/tags/{id}/aliases { "alias": "<name>" }` |
| Broader concept exists | Create child tag under it |
| No match | Create new tag |

**Alias vs new tag**:
- **Alias**: same concept, different spellings/languages ("ReactJS"→"React", "机器学习"→"Machine Learning", "ML"→"Machine Learning")
- **Child tag**: a genuine sub-domain ("PyTorch" under "Deep Learning")
- **New tag**: truly new domain

**API**:
```
POST   /v1/tags                          { name, color: "#hex", parent_ids: [] }
POST   /v1/tags/{id}/aliases             { "alias": "<name>" }
DELETE /v1/tags/aliases/{alias_id}
POST   /v1/tags/{child_id}/parents       { "parent_id": "<id>" }
DELETE /v1/tags/{child_id}/parents/{parent_id}
PATCH  /v1/tags/{id}                     { name, color }
DELETE /v1/tags/{id}
```

---

### 4. Bulk Retag

1. `POST /v1/resources/query` to get the batch
2. For each: `POST /v1/resources/{id}/tags` or `DELETE /v1/resources/{id}/tags/{tag_id}`
3. Verify with another query

---

## API Reference

```
# Resources
POST   /v1/resources/query              unified query (tag or keyword)
POST   /v1/resources                    create (URL/web only)
GET    /v1/resources/{id}
PATCH  /v1/resources/{id}              { title, description, url }
DELETE /v1/resources/{id}
GET    /v1/resources/by-url?url=<encoded>
POST   /v1/resources/{id}/tags         { "tag_id": "<id>" }
DELETE /v1/resources/{id}/tags/{tag_id}

# Tags
GET    /v1/tags                        list all
GET    /v1/tags/tree                   full DAG tree
GET    /v1/tags/search?query=<q>&tag_scope=<scope>
POST   /v1/tags                        { name, color, parent_ids }
PATCH  /v1/tags/{id}
DELETE /v1/tags/{id}
POST   /v1/tags/{id}/aliases           { "alias": "<name>" }
DELETE /v1/tags/aliases/{alias_id}
POST   /v1/tags/{child_id}/parents     { "parent_id": "<id>" }
DELETE /v1/tags/{child_id}/parents/{parent_id}

# Filesystem Daemon
POST   http://localhost:47832/index    { "path": "/absolute/path" }
→ { "id": "<resource-id>", "created": true|false }
```

---

## Customization

Your taxonomy: [`references/taxonomy.md`](references/taxonomy.md) — review to understand existing structure before making changes.
