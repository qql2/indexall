# 文件系统 Connector 设计

## 核心设计

**稳定 ID 策略**：
```
externalId = "{machineId}:{absolutePath}"
例：externalId = "macbook-pro-123:/Users/construct/Documents/article.pdf"
```

**原理**：
- 文件的唯一标识就是其所在机器 + 绝对路径
- 当文件路径变化（移动/改名）时，externalId 自动更新
- 当文件被复制时，新副本有新路径 → 新 externalId → 自动创建新索引

---

## 变更追踪

使用 FSWatcher（文件系统监听器）实时捕获变更：

| 事件 | 处理逻辑 | 结果 |
|------|--------|------|
| **file_modified** | 更新 fileSize、modifiedAt | 索引保持同步 |
| **file_moved/renamed** | externalId 变化 → 更新资源路径信息 | 自动追踪新位置 |
| **file_deleted** | watcher 检测路径不存在 | 标记资源为 `deleted` |
| **file_added** | 新路径 → 新 externalId | 根据配置决定是否自动索引 |
| **batch changes** | 500ms 防抖后统一处理 | 避免重复更新 |

---

## Connector 能力声明

```typescript
{
  hasStableId: true,           // ✓ 机器 + 路径组合是稳定的
  supportsWebhook: false,      // ✗ 文件系统无 push 通知
  supportsListChanges: false,  // ✗ 无 API 查询变更列表
  supportsBidirectional: false // ✗ 不支持反向写回
}
```

---

## 配置与生命周期

```typescript
interface FileSystemConnectorConfig {
  monitoredDirs: {
    path: string
    recursive: boolean
    ignorePatterns?: string[]  // 忽略规则（.gitignore 格式）
  }[]
  autoIndexNewFiles: boolean   // 新文件是否自动索引
  maxFileSize?: number         // 忽略超大文件
}

// 生命周期
1. authenticate() → 无需认证（本地文件系统）
2. start()        → 扫描初始目录 + 启动 FSWatcher
3. stop()         → 关闭 watcher（保留数据）
4. teardown()     → 删除所有索引（可选）
```

---

## 方案演变（已评估的方案）

### 方案 A：xattr（扩展文件属性）
在文件元数据中存储索引 ID。**已弃用**：xattr 易被网盘同步工具破坏，跨平台兼容性差。

### 方案 B：inode + 平台索引
用文件系统的 inode 追踪。**已弃用**：inode 跨文件系统时会变化，实现复杂度高。

### 方案 C：文件名中写入 ID
如 `article[idx_abc123].pdf`。**已弃用**：文件名侵入性强，用户体验差。

### 方案 D：机器 + 路径作为 externalId（最终选定）
**优势**：最简单、跨平台、符合直觉、实现代码量少。

---

## 实现路径

**Phase 1**：
- 后端 API：创建/查询文件系统 Connector 配置
- 手动扫描目录接口

**Phase 2**：
- 开发本地守护进程（Rust）
- 实现实时 FSWatcher（跨平台）
- 变更推送到后端 API

**Phase 3+**：
- 跨平台优化（macOS FSEvents → Linux inotify → Windows WatchDir）
- 网络文件系统支持（SMB/NFS）
- 智能冲突处理（并发修改、批量操作）
