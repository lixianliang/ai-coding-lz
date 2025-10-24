# ImgAgent API 文档

## 基础信息

- **Base URL**: `http://localhost:8000`
- **API Version**: `/v1`
- **认证方式**: Bearer Token (当前暂未启用)

## API 端点

### 文档管理 (Documents)

#### 1. 创建文档

创建一个新的文档并自动进行文本分割处理。

**请求**

```
POST /v1/documents
```

**请求体**

```json
{
  "name": "文档名称",
  "url": "文档URL或存储路径"
}
```

**字段说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 文档名称，最大长度50字符，同一数据集内名称唯一 |
| url | string | 是 | 文档URL或存储路径，最大长度200字符 |

**响应**

```json
{
  "id": "文档ID",
  "name": "文档名称",
  "status": "indexing",
  "created_at": "2024-10-24 12:00:00",
  "updated_at": "2024-10-24 12:00:00"
}
```

**状态码**

- `200`: 创建成功
- `400`: 请求参数错误
- `614`: 文档已存在

**说明**

- 文档创建后会自动进行文本分割，分割参数：
  - chunk_size: 2000
  - chunk_overlap: 100
  - separator: "\n\n"
- 文档初始状态为 `indexing`，处理完成后变为 `ready`

---

#### 2. 获取文档详情

获取指定文档的详细信息。

**请求**

```
GET /v1/documents/:document_id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |

**响应**

```json
{
  "id": "文档ID",
  "name": "文档名称",
  "status": "ready",
  "created_at": "2024-10-24 12:00:00",
  "updated_at": "2024-10-24 12:00:00"
}
```

**状态码**

- `200`: 获取成功
- `400`: 文档ID无效
- `612`: 文档不存在

---

#### 3. 更新文档

更新文档的名称。

**请求**

```
PUT /v1/documents/:document_id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |

**请求体**

```json
{
  "name": "新的文档名称"
}
```

**字段说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 文档名称，最大长度50字符 |

**响应**

```json
{
  "id": "文档ID",
  "name": "新的文档名称",
  "status": "ready",
  "created_at": "2024-10-24 12:00:00",
  "updated_at": "2024-10-24 12:30:00"
}
```

**状态码**

- `200`: 更新成功
- `400`: 请求参数错误
- `612`: 文档不存在

---

#### 4. 删除文档

删除指定文档及其所有章节。

**请求**

```
DELETE /v1/documents/:document_id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |

**响应**

```json
null
```

**状态码**

- `200`: 删除成功
- `400`: 文档ID无效

**说明**

- 删除文档会同时删除该文档下的所有章节

---

#### 5. 获取文档列表

获取所有文档的列表。

**请求**

```
GET /v1/documents
```

**响应**

```json
{
  "documents": [
    {
      "id": "文档ID1",
      "name": "文档名称1",
      "status": "ready",
      "created_at": "2024-10-24 12:00:00",
      "updated_at": "2024-10-24 12:00:00"
    },
    {
      "id": "文档ID2",
      "name": "文档名称2",
      "status": "indexing",
      "created_at": "2024-10-24 13:00:00",
      "updated_at": "2024-10-24 13:00:00"
    }
  ]
}
```

**状态码**

- `200`: 获取成功

**说明**

- 文档列表按更新时间倒序排列

---

### 章节管理 (Chapters)

#### 6. 获取章节详情

获取文档中指定章节的详细信息。

**请求**

```
GET /v1/documents/:document_id/Chapters/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |
| id | string | 章节ID |

**响应**

```json
{
  "id": "章节ID",
  "index": 0,
  "document_id": "文档ID",
  "content": "章节内容",
  "created_at": "2024-10-24 12:00:00",
  "updated_at": "2024-10-24 12:00:00"
}
```

**状态码**

- `200`: 获取成功
- `400`: 参数无效
- `500`: 获取失败

---

#### 7. 更新章节

更新章节的内容。

**请求**

```
PUT /v1/documents/:document_id/Chapters/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |
| id | string | 章节ID |

**请求体**

```json
{
  "content": "新的章节内容"
}
```

**字段说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | 是 | 章节内容，最大长度4000字符 |

**响应**

```json
{
  "id": "章节ID",
  "index": 0,
  "document_id": "文档ID",
  "content": "新的章节内容",
  "created_at": "2024-10-24 12:00:00",
  "updated_at": "2024-10-24 12:30:00"
}
```

**状态码**

- `200`: 更新成功
- `400`: 请求参数错误
- `500`: 更新失败

---

#### 8. 删除章节

删除文档中的指定章节。

**请求**

```
DELETE /v1/documents/:document_id/Chapters/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |
| id | string | 章节ID |

**响应**

```json
null
```

**状态码**

- `200`: 删除成功
- `400`: 参数无效
- `500`: 删除失败

---

#### 9. 获取章节列表

获取文档的所有章节列表。

**请求**

```
GET /v1/documents/:document_id/Chapters
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |

**响应**

```json
{
  "Chapters": [
    {
      "id": "章节ID1",
      "index": 0,
      "document_id": "文档ID",
      "content": "章节内容1",
      "created_at": "2024-10-24 12:00:00",
      "updated_at": "2024-10-24 12:00:00"
    },
    {
      "id": "章节ID2",
      "index": 1,
      "document_id": "文档ID",
      "content": "章节内容2",
      "created_at": "2024-10-24 12:00:00",
      "updated_at": "2024-10-24 12:00:00"
    }
  ]
}
```

**状态码**

- `200`: 获取成功
- `400`: 参数无效

**说明**

- 章节列表按 index 升序排列
- 后续需要考虑分页功能

---

## 数据模型

### Document (文档)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 文档唯一标识，32位UUID |
| name | string | 文档名称，最大128字符 |
| status | string | 文档状态：`indexing` (索引中) 或 `ready` (就绪) |
| created_at | string | 创建时间，格式：YYYY-MM-DD HH:MM:SS |
| updated_at | string | 更新时间，格式：YYYY-MM-DD HH:MM:SS |

### Chapter (章节)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 章节唯一标识，32位UUID |
| index | integer | 章节序号，从0开始 |
| document_id | string | 所属文档ID |
| content | string | 章节内容，最大10000字符 |
| created_at | string | 创建时间，格式：YYYY-MM-DD HH:MM:SS |
| updated_at | string | 更新时间，格式：YYYY-MM-DD HH:MM:SS |

---

## 错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权 (当前未启用) |
| 403 | 禁止访问 (当前未启用) |
| 500 | 服务器内部错误 |
| 612 | 文档不存在 |
| 614 | 文档已存在 |

---

## 配置说明

服务配置文件为 `imgagent.json`，主要配置项包括：

```json
{
  "log_conf": {
    "level": "debug",
    "file": "",
    "access_file": ""
  },
  "bind_host": ":8000",
  "api_version": "/v1",
  "temp": "./temp",
  "db": {
    "host": "localhost",
    "port": 3306,
    "user": "root",
    "password": "123456",
    "database": "imgagent"
  },
  "redis": {
    "disable_cluster": true,
    "expire_secs": 60,
    "addrs": ["localhost:6379"]
  },
  "storage": {
    "bucket": "bucket1",
    "domain": "bucket1.com",
    "ak": "xxx",
    "sk": "xxx"
  }
}
```

### 配置项说明

- **log_conf**: 日志配置
  - `level`: 日志级别 (debug/info/warn/error)
  - `file`: 日志文件路径
  - `access_file`: 访问日志文件路径

- **bind_host**: 服务监听地址和端口

- **api_version**: API版本前缀

- **temp**: 临时文件存储目录

- **db**: 数据库配置
  - 支持 MySQL 和 SQLite

- **redis**: Redis配置
  - 用于缓存，过期时间默认120秒

- **storage**: 对象存储配置
  - 用于存储文档文件

---

## 使用示例

### 创建文档

```bash
curl -X POST http://localhost:8000/v1/documents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试文档",
    "url": "https://example.com/document.txt"
  }'
```

### 获取文档列表

```bash
curl http://localhost:8000/v1/documents
```

### 获取文档详情

```bash
curl http://localhost:8000/v1/documents/{document_id}
```

### 更新文档

```bash
curl -X PUT http://localhost:8000/v1/documents/{document_id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "新文档名称"
  }'
```

### 删除文档

```bash
curl -X DELETE http://localhost:8000/v1/documents/{document_id}
```

### 获取章节列表

```bash
curl http://localhost:8000/v1/documents/{document_id}/Chapters
```

### 更新章节

```bash
curl -X PUT http://localhost:8000/v1/documents/{document_id}/Chapters/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "content": "更新后的章节内容"
  }'
```
