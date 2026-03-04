# IndexAll - API 设计文档

⚠️ **重要**: 本项目已从 tRPC 迁移到 **Kratos v2 + gRPC**。

## 📖 最新 API 设计

**所有 API 接口定义请查看**: [QUERY_API_DESIGN.md](QUERY_API_DESIGN.md)

## 核心端点速查

### Tag 操作
- `POST /v1/tags` - 创建标签
- `PATCH /v1/tags/{id}` - 更新标签
- `DELETE /v1/tags/{id}` - 删除标签
- `GET /v1/tags/search?query=...&tag_scope=...` - 搜索标签

### Resource 查询（统一端点）
- `POST /v1/resources/query` - **统一查询接口**（支持按标签过滤或关键词搜索）
- `GET /v1/resources/{id}` - 获取单个资源
- `GET /v1/resources/by-url?url=...` - 按 URL 查询

### Resource 基本操作
- `POST /v1/resources` - 创建资源
- `PATCH /v1/resources/{id}` - 更新资源
- `DELETE /v1/resources/{id}` - 删除资源
- `POST /v1/resources/{id}/tags` - 添加标签
- `DELETE /v1/resources/{id}/tags/{tag_id}` - 移除标签

---

## 快速开始

### 例1：获取某标签下的所有资源
```bash
POST /v1/resources/query
{
  "tag_query": {
    "tag_id": "tag-learning-python",
    "tag_scope": "WITH_DESCENDANTS"
  }
}
```

### 例2：全文搜索资源
```bash
POST /v1/resources/query
{
  "keyword_query": {
    "keyword": "python",
    "field_scope": "ALL",
    "tag_scope": "WITH_ANCESTORS"
  }
}
```

### 例3：搜索标签
```bash
GET /v1/tags/search?query=python&tag_scope=WITH_ANCESTORS
```

---

## 详细文档

所有完整的 Proto 定义、业务规则、使用示例和数据库设计请见:
- **[QUERY_API_DESIGN.md](QUERY_API_DESIGN.md)** - 完整 API 规范
- **[MODEL.md](MODEL.md)** - 数据模型和查询示例
- **[DESIGN.md](DESIGN.md)** - 产品设计和架构
