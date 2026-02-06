# IndexAll - API 契约设计

基于 tRPC，按领域划分 router。以下定义每个 procedure 的输入输出和行为。

## Router 总览

```
trpc
├── tag
│   ├── create
│   ├── update
│   ├── delete
│   ├── list
│   ├── getTree
│   ├── search
│   ├── addAlias
│   ├── removeAlias
│   ├── addParent
│   └── removeParent
├── resource
│   ├── create
│   ├── update
│   ├── delete
│   ├── get
│   ├── list
│   ├── search
│   ├── getByUrl
│   ├── addTag
│   └── removeTag
└── (后期) connector
    ├── list
    ├── connect
    ├── disconnect
    └── sync
```

---

## Tag Router

### tag.create

创建标签。

```typescript
input: {
  name: string
  color?: string
  aliases?: string[]        // 初始别名
  parentIds?: string[]      // 初始父标签
}
output: {
  id: string
  name: string
  color: string | null
  aliases: string[]
  parentIds: string[]
  createdAt: string
}
```

业务规则：
- `name` 不能与已有标签名或别名重复
- `aliases` 中的每一项不能与已有标签名或别名重复
- `parentIds` 需校验不会产生环

### tag.update

更新标签基本信息。

```typescript
input: {
  id: string
  name?: string
  color?: string
}
output: { success: true }
```

业务规则：
- `name` 修改时同样校验全局唯一

### tag.delete

删除标签。

```typescript
input: { id: string }
output: { success: true }
```

业务规则：
- 级联删除：别名、DAG 关系（作为父或子）、资源关联
- 不删除子标签本身，仅解除关系
- 不删除关联的资源，仅解除标签关联

### tag.list

获取所有标签（扁平列表），用于标签选择器等场景。

```typescript
input: {}
output: Array<{
  id: string
  name: string
  color: string | null
  aliases: string[]
  parentIds: string[]
  resourceCount: number     // 直接关联的资源数
}>
```

### tag.getTree

获取标签树形结构，用于左侧标签树面板。

```typescript
input: {}
output: Array<{
  id: string
  name: string
  color: string | null
  resourceCount: number     // 含子标签递归的资源总数
  children: TreeNode[]      // 递归嵌套
}>
```

说明：
- 没有父标签的标签作为根节点
- 有多个父标签的标签会在多个位置平等出现（DAG 特性）
- `resourceCount` 递归计算，包含所有子标签下的资源

### tag.search

搜索标签（按名称和别名匹配），用于标签选择器的输入即搜。

```typescript
input: {
  query: string
}
output: Array<{
  id: string
  name: string
  color: string | null
  matchedAlias?: string     // 如果是通过别名匹配到的，返回该别名
}>
```

### tag.addAlias

给标签添加别名。

```typescript
input: {
  tagId: string
  alias: string
}
output: { id: string; alias: string }
```

业务规则：
- 别名不能与已有标签名或别名重复

### tag.removeAlias

移除标签别名。

```typescript
input: { aliasId: string }
output: { success: true }
```

### tag.addParent

添加父标签关系。

```typescript
input: {
  childId: string
  parentId: string
}
output: { success: true }
```

业务规则：
- 校验不会产生环（递归 CTE 环检测）
- 禁止自引用

### tag.removeParent

移除父标签关系。

```typescript
input: {
  childId: string
  parentId: string
}
output: { success: true }
```

---

## Resource Router

### resource.create

创建资源。

```typescript
input: {
  url?: string
  title: string
  description?: string
  source?: string             // 默认 'manual'
  externalId?: string
  openWith?: string
  metadata?: Record<string, any>
  tagIds?: string[]           // 初始标签
}
output: {
  id: string
  title: string
  url: string | null
  source: string
  tags: Array<{ id: string; name: string; color: string | null }>
  createdAt: string
}
```

业务规则：
- 如果提供 `source + externalId`，校验不重复
- 如果提供 `url` 且 `title` 为空，后端尝试抓取页面标题

### resource.update

更新资源信息。

```typescript
input: {
  id: string
  title?: string
  description?: string
  url?: string
  openWith?: string
}
output: { success: true }
```

说明：
- 不通过此接口修改标签，标签增删通过 `resource.addTag` / `resource.removeTag`
- 不通过此接口修改 `source` / `externalId`（来源信息不可变）

### resource.delete

删除资源。

```typescript
input: { id: string }
output: { success: true }
```

业务规则：
- 级联删除资源-标签关联
- 同步删除 FTS5 索引（通过触发器自动）

### resource.get

获取单个资源详情。

```typescript
input: { id: string }
output: {
  id: string
  source: string
  externalId: string | null
  title: string
  description: string | null
  url: string | null
  openWith: string | null
  metadata: Record<string, any> | null
  status: 'active' | 'stale' | 'deleted'
  syncedAt: string | null
  createdAt: string
  updatedAt: string
  tags: Array<{ id: string; name: string; color: string | null }>
}
```

### resource.list

分页获取资源列表，支持按标签筛选。

```typescript
input: {
  tagId?: string              // 按标签筛选（含子标签递归）
  status?: 'active' | 'stale' | 'deleted'  // 默认 'active'
  page?: number               // 默认 1
  pageSize?: number           // 默认 20
}
output: {
  items: Array<{
    id: string
    source: string
    title: string
    description: string | null
    url: string | null
    status: string
    createdAt: string
    tags: Array<{ id: string; name: string; color: string | null }>
  }>
  total: number
  page: number
  pageSize: number
}
```

### resource.search

全文搜索资源。

```typescript
input: {
  query: string
  page?: number
  pageSize?: number
}
output: {
  items: Array<{
    id: string
    source: string
    title: string
    description: string | null
    url: string | null
    createdAt: string
    tags: Array<{ id: string; name: string; color: string | null }>
    matchSource: 'title' | 'description' | 'tag' | 'alias'
  }>
  total: number
  page: number
  pageSize: number
}
```

说明：
- 同时搜索标题/描述（FTS5）和标签名/别名（LIKE）
- `matchSource` 告诉前端是通过哪个字段匹配到的，用于高亮

### resource.getByUrl

根据 URL 查找已有资源，用于浏览器扩展判断"当前页面是否已收藏"。

```typescript
input: { url: string }
output: {
  id: string
  title: string
  tags: Array<{ id: string; name: string; color: string | null }>
} | null
```

### resource.addTag

给资源添加标签。

```typescript
input: {
  resourceId: string
  tagId: string
}
output: { success: true }
```

### resource.removeTag

移除资源的标签。

```typescript
input: {
  resourceId: string
  tagId: string
}
output: { success: true }
```

---

## 客户端调用场景

### Web UI 使用的接口

| 页面/组件 | 调用接口 |
|---|---|
| 标签树面板 | `tag.getTree` |
| 资源列表 | `resource.list` (带 tagId 筛选) |
| 搜索 | `resource.search` |
| 添加资源对话框 | `resource.create`, `tag.search`, `tag.create` |
| 编辑资源对话框 | `resource.update`, `resource.addTag`, `resource.removeTag` |
| 编辑标签对话框 | `tag.update`, `tag.addAlias`, `tag.removeAlias`, `tag.addParent`, `tag.removeParent` |
| 标签选择器 | `tag.search`, `tag.create` |

### 浏览器扩展使用的接口

| 场景 | 调用接口 |
|---|---|
| 打开弹窗 | `resource.getByUrl`（判断是否已收藏） |
| 搜索/选择标签 | `tag.search`, `tag.create` |
| 保存 | `resource.create` 或 `resource.addTag`（已收藏时追加标签） |
