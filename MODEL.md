## 数据模型

### ER 关系

```
┌──────────┐       ┌───────────────┐       ┌──────────┐
│   tags   │──1:N──│  tag_aliases  │       │resources │
└──────────┘       └───────────────┘       └──────────┘
     │  ▲                                       │
     │  │  ┌───────────────┐                    │
     │  └──│ tag_relations │                    │
     │     │ (parent→child)│                    │
     └────>│  DAG 邻接表    │                    │
            └───────────────┘                    │
     │                                          │
     │         ┌───────────────┐                │
     └────────>│ resource_tags │<───────────────┘
               │   (M:N 关联)  │
               └───────────────┘
```

### DAG 存储方案：邻接表

标签的 DAG 关系使用邻接表（Adjacency List）存储，通过递归 CTE 查询子树。

**选择邻接表而非闭包表的理由**：
- 标签规模预计在数百到数千级别，递归 CTE 性能完全足够
- 写操作简单（增删父子关系只需一行 INSERT/DELETE）
- 闭包表在 DAG 场景下维护成本高（添加/删除关系需要更新大量传递路径），收益不明显

### 表结构

```sql
-- 标签
CREATE TABLE tags (
  id TEXT PRIMARY KEY,              -- UUID
  name TEXT NOT NULL,
  color TEXT,                       -- 标签颜色，用于 UI 展示
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- 标签别名
CREATE TABLE tag_aliases (
  id TEXT PRIMARY KEY,              -- UUID
  tag_id TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  alias TEXT NOT NULL UNIQUE        -- 别名全局唯一，避免歧义
);
CREATE INDEX idx_tag_aliases_tag_id ON tag_aliases(tag_id);

-- 标签 DAG 关系（邻接表）
CREATE TABLE tag_relations (
  parent_id TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  child_id TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (parent_id, child_id),
  CHECK (parent_id != child_id)     -- 禁止自引用
);
CREATE INDEX idx_tag_relations_child ON tag_relations(child_id);

-- 资源
CREATE TABLE resources (
  id TEXT PRIMARY KEY,              -- UUID
  source TEXT NOT NULL DEFAULT 'manual',  -- 来源系统
  external_id TEXT,                 -- 来源系统的稳定 ID，手动添加时为空
  title TEXT NOT NULL,
  description TEXT,
  url TEXT,
  open_with TEXT,                   -- 打开方式 (deep link scheme)
  metadata TEXT,                    -- JSON，来源系统的原始元数据
  status TEXT NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'stale', 'deleted')),
  synced_at TEXT,                   -- 最后同步时间
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
-- 来源+外部ID 复合唯一（SQLite 允许多个 NULL，所以手动添加的不冲突）
CREATE UNIQUE INDEX idx_resources_source_external
  ON resources(source, external_id) WHERE external_id IS NOT NULL;
CREATE INDEX idx_resources_status ON resources(status);

-- 资源-标签关联（多对多）
CREATE TABLE resource_tags (
  resource_id TEXT NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  tag_id TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (resource_id, tag_id)
);
CREATE INDEX idx_resource_tags_tag ON resource_tags(tag_id);

-- 全文搜索（SQLite FTS5）
CREATE VIRTUAL TABLE resources_fts USING fts5(
  title,
  description,
  content='resources',
  content_rowid='rowid'
);
```

### 关键查询

#### 递归获取某个标签的所有子标签

```sql
WITH RECURSIVE subtags AS (
  -- 起点：目标标签自身
  SELECT id FROM tags WHERE id = :tag_id
  UNION ALL
  -- 递归：沿 DAG 向下展开
  SELECT tr.child_id
  FROM tag_relations tr
  JOIN subtags st ON tr.parent_id = st.id
)
SELECT * FROM subtags;
```

#### 按标签筛选资源（含子标签递归）

```sql
WITH RECURSIVE subtags AS (
  SELECT id FROM tags WHERE id = :tag_id
  UNION ALL
  SELECT tr.child_id
  FROM tag_relations tr
  JOIN subtags st ON tr.parent_id = st.id
)
SELECT DISTINCT r.*
FROM resources r
JOIN resource_tags rt ON r.id = rt.resource_id
WHERE rt.tag_id IN (SELECT id FROM subtags)
  AND r.status = 'active'
ORDER BY r.created_at DESC;
```

#### 搜索资源（标题 + 描述 + 标签名 + 别名）

```sql
-- FTS5 搜索标题和描述
SELECT r.*
FROM resources_fts fts
JOIN resources r ON r.rowid = fts.rowid
WHERE resources_fts MATCH :query AND r.status = 'active'

UNION

-- 标签名/别名匹配，返回关联资源
SELECT DISTINCT r.*
FROM tags t
LEFT JOIN tag_aliases ta ON t.id = ta.tag_id
JOIN resource_tags rt ON t.id = rt.tag_id
JOIN resources r ON rt.resource_id = r.id
WHERE (t.name LIKE :pattern OR ta.alias LIKE :pattern)
  AND r.status = 'active'

ORDER BY created_at DESC;
```

#### DAG 环检测（添加关系前校验）

```sql
-- 添加 parent_id → child_id 之前，检查 child_id 是否是 parent_id 的祖先
WITH RECURSIVE ancestors AS (
  SELECT parent_id AS id FROM tag_relations WHERE child_id = :parent_id
  UNION ALL
  SELECT tr.parent_id
  FROM tag_relations tr
  JOIN ancestors a ON tr.child_id = a.id
)
SELECT EXISTS (SELECT 1 FROM ancestors WHERE id = :child_id) AS has_cycle;
```

### FTS5 同步

资源表的增删改需要同步到 FTS5 索引，通过触发器自动维护：

```sql
-- 插入
CREATE TRIGGER resources_ai AFTER INSERT ON resources BEGIN
  INSERT INTO resources_fts(rowid, title, description)
  VALUES (new.rowid, new.title, new.description);
END;

-- 更新
CREATE TRIGGER resources_au AFTER UPDATE ON resources BEGIN
  INSERT INTO resources_fts(resources_fts, rowid, title, description)
  VALUES ('delete', old.rowid, old.title, old.description);
  INSERT INTO resources_fts(rowid, title, description)
  VALUES (new.rowid, new.title, new.description);
END;

-- 删除
CREATE TRIGGER resources_ad AFTER DELETE ON resources BEGIN
  INSERT INTO resources_fts(resources_fts, rowid, title, description)
  VALUES ('delete', old.rowid, old.title, old.description);
END;
```
