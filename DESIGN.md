# IndexAll - 统一资源索引平台 方案设计

## 核心思想

**给万物归类**——对所有软件的资源进行统一索引，在本平台中检索和浏览，并能快捷跳转到对应资源的原始展现形式。

## 标签系统

采用网状结构的多标签体系：

- **多标签**：一个资源可以属于多个标签
- **网状层级**：一个标签可以属于多个父标签（DAG 有向无环图），浏览时仍可渲染为层级树形结构
- **标签别名**：一个标签可以有多个别名，扩大搜索召回率

## 资源模型

每个外部资源用统一的抽象模型表示：

```typescript
interface Resource {
  id: string                     // 内部唯一 ID
  source: string                 // 来源系统，如 "github" | "notion" | ...
  externalId: string             // 在来源系统中的稳定 ID
  title: string
  description?: string
  url?: string                   // 可直接打开的 URL
  openWith?: string              // 打开方式（浏览器、App deep link 等）
  tags: TagId[]
  metadata: Record<string, any>  // 来源系统的原始元数据
  syncedAt: Date                 // 最后同步时间
  status: 'active' | 'stale' | 'deleted'
}
```

**`source + externalId` 构成复合主键**，是映射关系的锚点。

## 第三方系统联动

### 采集策略

按实现优先级排列：

| 优先级 | 策略 | 适用场景 | 示例 |
|---|---|---|---|
| P0 | 浏览器扩展 | 网页书签、阅读列表 | Chrome/Firefox Extension，用户主动收藏 |
| P1 | API 集成 | 有开放 API 的平台 | GitHub stars/repos、Notion pages |
| P2 | 导入/同步 | 有导出功能的系统 | 浏览器书签 HTML 导出、Obsidian vault |
| P3 | 系统级 Hook | 本地文件/应用 | macOS Spotlight metadata、文件系统 watcher |
| P3 | 协议处理 | 通用标准 | RSS/Atom feeds、OPML |

### 资源追踪

核心原则：**用稳定 ID 做锚点，用可变信息做辅助，接受不完美并优雅降级。**

#### 追踪依据：外部系统的稳定 ID

大多数有 API 的系统都提供不可变标识符：

| 来源 | 稳定 ID | 说明 |
|---|---|---|
| GitHub | repo id（数字） | rename repo 后 id 不变 |
| Notion | page id（UUID） | 页面移动/改名后不变 |
| 浏览器书签 | 浏览器内部 id | Chrome Bookmarks API 提供 `id` 字段 |

URL、路径、标题等作为**可变元数据**存储，不作为标识依据。

> **前期决策**：只接入有稳定 ID 的来源系统。无稳定 ID 的资源（如纯网页 URL）追踪复杂度高，留到后期实现。

#### 变更检测

| 方式 | 适用场景 |
|---|---|
| Webhook / 推送 | GitHub、Notion 等支持推送的平台 |
| 定期轮询 | 通用方案，调 API 对比差异 |
| 文件系统 watcher | 本地文件（后期） |
| 用户手动触发 | 兜底方案，永远保留 |

#### 变更处理（Reconciliation）

```
外部资源变化
    │
    ├─ 能通过 source + externalId 匹配到
    │     → 更新本地元数据（标题/URL/路径），保留用户标签不动
    │
    ├─ externalId 在来源系统中已不存在
    │     → 标记为 deleted
    │
    └─ 发现新资源（本地没有的）
          → 根据用户策略：自动索引 or 忽略
```

**关键点**：用户打的标签与来源系统的元数据严格分离，同步只更新元数据，不触碰用户标签。

### Connector 抽象

每个第三方系统实现一个 Connector，声明自身能力：

```typescript
interface ConnectorCapabilities {
  hasStableId: boolean          // 有稳定 ID
  supportsWebhook: boolean      // 支持推送通知
  supportsListChanges: boolean  // 能查询变更列表
  supportsBidirectional: boolean // 能反向写回
}
```

UI 层根据能力等级给用户设置不同预期（如"实时同步" vs "需手动刷新"）。

### 同步策略

采用**"索引即快照"**模式——本平台保存资源索引信息的快照，不保证与来源系统实时一致，但提供手动/定时刷新操作。

论据：本平台的定位是索引和分类，资源的权威版本始终在来源系统中，不需要双向同步。这大幅降低实现复杂度。

## 技术栈

全栈 TypeScript monorepo，前后端共享类型定义。

| 层面 | 选型 | 理由 |
|---|---|---|
| Monorepo | pnpm workspaces | 轻量，原生支持 workspace |
| 前端 | React + Vite | 浏览器扩展生态好，Vite 开发体验快 |
| UI | TailwindCSS + shadcn/ui | 组件可定制性强，不引入重运行时 |
| 后端 | Hono (Node.js) | 轻量、TypeScript-first，可平迁到其他运行时 |
| 数据库 | SQLite (better-sqlite3) | 零配置，本地优先的天然选择 |
| ORM | Drizzle | 同时支持 SQLite 和 PostgreSQL，后期可平滑迁移 |
| API 层 | tRPC | 前后端类型直连，省去手写接口定义 |
| 浏览器扩展 | WXT | 现代扩展框架，支持 React + TypeScript + HMR |

### 选型论据

**SQLite 而非 PostgreSQL**：本地优先场景下零配置零依赖；SQLite 3.8.3+ 支持递归 CTE，满足标签 DAG 层级查询；通过 Drizzle 抽象，后期迁移 PostgreSQL 只需改连接配置。

**Hono 而非 Express/Fastify/NestJS**：Express 类型支持弱；NestJS 太重，当前规模不需要；Hono 更轻且多运行时通用，tRPC 有 Hono adapter。

**tRPC**：monorepo 下类型直连是最大优势，改接口前端立刻有类型报错；如后期需开放 API 给第三方，可额外加 REST 路由，两者不冲突。

### 项目结构

```
indexall/
├── packages/
│   ├── web/          # React 前端（Vite）
│   ├── server/       # Hono 后端 + tRPC
│   ├── extension/    # WXT 浏览器扩展
│   └── shared/       # 共享类型、常量
├── pnpm-workspace.yaml
├── package.json
├── tsconfig.base.json
└── DESIGN.md
```

## 实现路径

1. **核心数据模型**：Resource + Tag 网状关系
2. **浏览器扩展 + Web UI**：最小可用产品，覆盖"给网页分类"的首要场景
3. **Connector 抽象层**：定义统一接口
4. **逐步接入更多系统**：按 P0 → P3 优先级推进
