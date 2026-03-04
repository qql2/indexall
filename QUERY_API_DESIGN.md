# IndexAll - 查询 API 设计方案

**版本**: MVP v1.0
**状态**: ✅ 已确认
**更新时间**: 2026-03-04

---

## 概述

本文档定义 IndexAll 的核心查询 API，用于支持资源和标签的灵活搜索。设计目标：
- 🎯 简洁：少而精的端点设计，避免功能重叠
- 🎯 高效：为 AI 调用优化，减少 token 消耗
- 🎯 清晰：语义明确，易于理解和使用

---

## 核心原则

### 1. API 简洁性
- 用**一个统一的查询端点** (`POST /v1/resources/query`) 替代多个搜索端点
- 去掉冗余端点（旧的 `GET /v1/resources` 和 `GET /v1/resources/search`）
- 保留必要的简单端点（`GET /v1/resources/by-url`）

### 2. 搜索的正交维度（3个）

| 维度 | 可选值 | 说明 |
|------|--------|------|
| **搜索对象** | Resource / Tag | 搜什么 |
| **查询方式** | ①按标签 / ②按关键词 | 怎么搜（二选一） |
| **标签范围** | DIRECT / WITH_ANCESTORS / WITH_DESCENDANTS | 仅当①时适用 |

### 3. 默认匹配规则
- **模糊匹配**是系统默认行为，用户无需选择
- 例：搜 `yth` 能匹配到 `python`，搜 `learn` 能匹配到 `learning`
- Exact/Prefix/Regex 等精确匹配规则留到 **Phase 2**

---

## Tag API（保持不变）

### `GET /v1/tags/search`

搜索标签（按名称、别名、描述匹配），默认模糊匹配。支持在标签的不同层级搜索。

```protobuf
message SearchTagsRequest {
  string query = 1;           // 搜索关键词（模糊匹配）

  // 标签范围：在标签的哪些层级搜索
  enum TagScope {
    DIRECT = 0;              // 仅搜自己的信息
    WITH_ANCESTORS = 1;      // 也搜上级标签信息（默认）
    WITH_DESCENDANTS = 2;    // 也搜下级标签信息
  }
  TagScope tag_scope = 2;  // 默认WITH_ANCESTORS

  int32 limit = 3;            // 默认20
  int32 offset = 4;           // 默认0
}

message SearchTagsResponse {
  repeated TagResult tags = 1;
  int32 total = 2;
}

message TagResult {
  string id = 1;
  string name = 2;
  string description = 3;
  repeated string aliases = 4;
  int32 resource_count = 5;   // 该标签直接关联的资源数
}
```

---

## Resource API（重构）

### 精简后的端点列表

| HTTP 方法 | 端点 | 用途 | 优先级 |
|----------|------|------|--------|
| POST | `/v1/resources` | 创建资源 | P0 |
| PATCH | `/v1/resources/{id}` | 更新资源 | P0 |
| DELETE | `/v1/resources/{id}` | 删除资源 | P0 |
| GET | `/v1/resources/{id}` | 获取单个资源详情 | P0 |
| **POST** | **`/v1/resources/query`** | **统一查询接口**（新） | **P0** |
| GET | `/v1/resources/by-url` | 按URL查询（浏览器扩展判重） | P0 |
| POST | `/v1/resources/{id}/tags` | 添加标签 | P1 |
| DELETE | `/v1/resources/{id}/tags/{tag_id}` | 移除标签 | P1 |

**删除的冗余端点**:
- ❌ `GET /v1/resources` (旧的list)
- ❌ `GET /v1/resources/search` (功能被query包含)

---

## 统一查询端点

### `POST /v1/resources/query`

高效、灵活的资源查询接口，支持按标签或关键词查询。

```protobuf
message ResourceQueryRequest {
  oneof query {
    TagQuery tag_query = 1;           // 方式1：按标签过滤
    KeywordQuery keyword_query = 2;   // 方式2：按关键词搜索
  }

  int32 page = 3;        // 页码，默认1
  int32 page_size = 4;   // 每页数量，默认20
  string sort_by = 5;    // 排序，默认"created_at_desc"
}

message TagQuery {
  string tag_id = 1;

  // 标签范围：查询该标签下的资源
  enum TagScope {
    DIRECT = 0;              // 仅该标签本身
    WITH_ANCESTORS = 1;      // 该标签 + 所有上级标签
    WITH_DESCENDANTS = 2;    // 该标签 + 所有下级标签（默认）
  }
  TagScope tag_scope = 2;
}

message KeywordQuery {
  string keyword = 1;        // 搜索关键词（默认模糊匹配）

  // 字段范围：在哪些字段搜索
  enum FieldScope {
    ALL = 0;                 // 全文搜索（title+description+tags）
    TITLE = 1;               // 仅title字段
    DESCRIPTION = 2;         // 仅description字段
  }
  FieldScope field_scope = 2;  // 默认ALL

  // 标签范围：在标签的哪些层级搜索
  enum TagScope {
    DIRECT = 0;              // 仅搜自己和直接关联的标签
    WITH_ANCESTORS = 1;      // 也搜上级标签信息（默认）
    WITH_DESCENDANTS = 2;    // 也搜下级标签信息
  }
  TagScope tag_scope = 3;  // 默认WITH_ANCESTORS
}

message ResourceQueryResponse {
  repeated ResourceSearchResult items = 1;
  int32 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}

message ResourceSearchResult {
  string id = 1;
  string title = 2;
  string description = 3;
  string url = 4;
  string source = 5;
  int64 created_at = 6;
  int64 updated_at = 7;

  repeated TagInfo tags = 8;

  // 可选：匹配源，用于前端高亮
  // "title" | "description" | "tag_name" | "tag_alias"
  string match_source = 9;
}

message TagInfo {
  string id = 1;
  string name = 2;
  string description = 3;
  repeated string aliases = 4;
}
```

---

## 使用示例

### 例1：按标签过滤资源（获取某标签及其下级的资源）

```json
POST /v1/resources/query
{
  "tag_query": {
    "tag_id": "tag-learning-python",
    "tag_scope": "WITH_DESCENDANTS"
  },
  "page": 1,
  "page_size": 20
}

响应：返回"Learning/Python"及其所有下级标签关联的资源
{
  "items": [
    {
      "id": "res-1",
      "title": "Python Basics",
      "description": "..."
    },
    ...
  ],
  "total": 42
}
```

### 例2：搜索资源（关键词在自己或上级标签信息中）

```json
POST /v1/resources/query
{
  "keyword_query": {
    "keyword": "python",
    "field_scope": "ALL",
    "tag_scope": "WITH_ANCESTORS"
  },
  "page": 1,
  "page_size": 50
}

说明：返回以下资源
- 标题/描述中有"python"
- 或 该资源关联的上级标签中有"python"
```

### 例3：搜索资源标题

```json
POST /v1/resources/query
{
  "keyword_query": {
    "keyword": "pytorch",
    "field_scope": "TITLE",
    "tag_scope": "WITH_DESCENDANTS"
  }
}

说明：搜索标题中的"pytorch"，以及下级标签中有"pytorch"的资源
```

### 例4：搜索标签

```json
GET /v1/tags/search?query=python&tag_scope=WITH_ANCESTORS

说明：搜索名称/别名/描述中有"python"的标签，
      以及上级标签中有"python"的标签
```

### 例5：全文搜索资源（AI查询）

```json
POST /v1/resources/query
{
  "keyword_query": {
    "keyword": "machine learning neural network",
    "field_scope": "ALL",
    "tag_scope": "WITH_ANCESTORS"
  },
  "page": 1,
  "page_size": 50
}

说明：全文搜索，搜索范围包括资源和其上级标签的所有字段
```

---

## 标签范围说明

Scope 定义**搜索范围**，即在哪些信息层级中搜索关键词。

### 示例：标签树结构

```
Learning（学习）
├─ Python
│  ├─ Basics（基础）
│  └─ Advanced（高级）
└─ JavaScript
```

### 三个 Scope 的含义

| Scope | 搜索范围 | 示例说明 |
|-------|---------|---------|
| **DIRECT** | 仅自己的信息 | 搜索'python'时，仅在"Python"标签的name/description中搜索 |
| **WITH_ANCESTORS** | 自己 + 所有上级标签的信息 | 搜索'python'时，在"Python"和"Learning"的信息中都搜索 |
| **WITH_DESCENDANTS** | 自己 + 所有下级标签的信息 | 搜索'python'时，在"Python"、"Basics"、"Advanced"的信息中都搜索 |

### 实际查询示例

**资源查询**：`POST /v1/resources/query`

```json
{
  "keyword_query": {
    "keyword": "python",
    "field_scope": "ALL",
    "tag_scope": "WITH_ANCESTORS"
  }
}
```

返回：标题/描述中有"python" **或** 该资源关联的上级标签中有"python"的所有资源

**标签查询**：`GET /v1/tags/search?query=python&tag_scope=WITH_ANCESTORS`

返回：名称/别名/描述中有"python" **或** 上级标签中有"python"的所有标签

---

## 其他端点

### `GET /v1/resources/{id}`

获取单个资源详情。

```protobuf
message GetResourceRequest {
  string id = 1;
}

message GetResourceResponse {
  ResourceSearchResult resource = 1;
}
```

### `GET /v1/resources/by-url`

根据 URL 查找资源，用于浏览器扩展判断"当前页面是否已收藏"。

```protobuf
message GetResourceByUrlRequest {
  string url = 1;
}

message GetResourceByUrlResponse {
  ResourceSearchResult resource = 1;  // 如果存在
  // 如果不存在，返回 404 或 empty
}
```

### `POST /v1/resources`

创建资源。

```protobuf
message CreateResourceRequest {
  string title = 1;
  string description = 2;
  string url = 3;
  string source = 4;              // 默认 'manual'
  string external_id = 5;         // 可选，外部系统ID
  repeated string tag_ids = 6;    // 初始标签
}

message CreateResourceResponse {
  string id = 1;
  string title = 2;
  string url = 3;
  int64 created_at = 4;
}
```

### `PATCH /v1/resources/{id}`

更新资源信息。

```protobuf
message UpdateResourceRequest {
  string id = 1;
  optional string title = 2;
  optional string description = 3;
  optional string url = 4;
}

message UpdateResourceResponse {
  bool success = 1;
}
```

### `DELETE /v1/resources/{id}`

删除资源。

```protobuf
message DeleteResourceRequest {
  string id = 1;
}

message DeleteResourceResponse {
  bool success = 1;
}
```

---

## 实现路径

### Phase 1（MVP）：基础版本

- ✅ 实现 `POST /v1/resources/query`
  - 支持 TagQuery（按标签查询，3个范围）
  - 支持 KeywordQuery（全文搜索 + 字段范围）
  - 默认模糊匹配，系统内部用 FTS5 + 编辑距离实现
- ✅ 保留 `GET /v1/resources/by-url`
- ✅ 为 tags 表创建 FTS5 虚拟表
- ✅ 删除旧的 list/search 端点

### Phase 2（增强）：精确匹配

- 支持 Exact/Prefix/Regex 匹配规则
- 支持复杂条件组合（AND/OR）
- 缓存优化

### Phase 3（优化）：性能调优

- 添加查询缓存
- 性能分析和优化
- 支持更复杂的查询表达式

---

## AI 调用优化

### 响应体精简

ResourceSearchResult 中已去除不必要的字段：
- ❌ 不返回 metadata（外部系统原始数据）
- ❌ 不返回 synced_at（仅内部使用）
- ✅ 返回 title + description + tags（AI 最关心的）

### 批量查询支持

通过单次 `/v1/resources/query` 可以：
- 获取某标签下的所有资源（避免多次调用）
- 分页控制，防止一次加载过多数据

### 模糊匹配的优势

- AI 可以输入 "pyt" 就能匹配 "python"
- 系统内部处理拼写变体和近似匹配
- 无需用户指定匹配规则

---

## 数据库支持

### FTS5 虚拟表

```sql
-- 资源全文索引（已有）
CREATE VIRTUAL TABLE resources_fts USING fts5(
  title,
  description,
  content='resources',
  content_rowid='id'
);

-- 标签全文索引（新增）
CREATE VIRTUAL TABLE tags_fts USING fts5(
  name,
  description,
  alias,
  content='tags',
  content_rowid='id'
);
```

### 标签递归查询

```sql
-- 获取标签的所有下级标签
WITH RECURSIVE descendants AS (
  SELECT id FROM tags WHERE id = :tag_id
  UNION ALL
  SELECT tr.child_id
  FROM tag_relations tr
  JOIN descendants d ON tr.parent_id = d.id
)
SELECT * FROM descendants;

-- 获取标签的所有上级标签
WITH RECURSIVE ancestors AS (
  SELECT id FROM tags WHERE id = :tag_id
  UNION ALL
  SELECT tr.parent_id
  FROM tag_relations tr
  JOIN ancestors a ON tr.child_id = a.id
)
SELECT * FROM ancestors;
```

---

## 参考

- DESIGN.md - 产品和架构设计
- MODEL.md - 数据模型和关键查询
- REQUIREMENTS.md - 功能需求
