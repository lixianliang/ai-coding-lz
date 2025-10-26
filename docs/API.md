# ImgAgent API 文档

## 概述

ImgAgent 是一个连环动漫智能体后端服务，使用 Go 语言实现，基于 Gin 框架构建 RESTful API。该服务可根据用户上传的小说文本，自动理解小说内容、提取人物特征和场景信息，生成连环画风格的动漫内容。

### 技术栈

- 编程语言: Go
- HTTP框架: Gin
- 数据库: MySQL 8.4
- ORM: GORM
- 日志: Zap

## 基础信息

- **Base URL**: `http://localhost:8000`
- **API Version**: `/v1`
- **认证方式**: Bearer Token (当前暂未启用)

## 响应格式

所有 API 接口统一使用以下响应格式：

```json
{
  "code": 200,
  "message": "",
  "reqid": "<请求ID>",
  "data": {}
}
```

**字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 业务处理状态码，200 表示成功，非 200 表示失败 |
| message | string | 错误信息，成功时为空 |
| reqid | string | 请求唯一标识，用于追踪和日志记录 |
| data | object | 业务返回数据，具体结构见各接口说明 |

**注意事项**

- HTTP 状态码为 200 表示服务接受并处理了请求
- 具体业务是否成功需查看响应 body 中的 `code` 字段
- `code` 为 200 表示业务处理成功，非 200 表示业务处理失败
- 失败时 `message` 字段包含详细的错误信息
- 每个请求都会在响应头中返回 `x-reqid`，与响应 body 中的 `reqid` 相同

## API 端点

### 文档管理 (Documents)

#### 1. 创建文档

创建一个新的文档并自动进行文本分割处理。

**请求**

```
POST /v1/documents

Content-Type: multipart/form-data
```

**请求体**

```json
--form 'name="名字"'
--form 'file=@example-file'
```

**字段说明**

| 字段   | 类型     | 必填 | 说明                    |
|------|--------|------|-----------------------|
| name | string | 是 | 文档名称，最大长度50字符|
| file | file   | 是 | 本地文件                  |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "id": "文档ID",
    "name": "文档名称",
    "status": "chapterReady",
    "created_at": "2024-10-24 12:00:00",
    "updated_at": "2024-10-24 12:00:00"
  }
}
```

**业务状态码**

- `200`: 创建成功
- `400`: 请求参数错误
- `614`: 文档已存在

**说明**

- 文档状态流转：`chapterReady`（章节准备就绪） → `roleReady`（角色准备就绪） → `sceneReady`（场景准备就绪） → `imgReady`（图片准备就绪）
- 初始状态：文档上传后为 `chapterReady`
- Worker 1 完成角色提取后，状态变为 `roleReady`
- Worker 2 完成场景生成后，状态变为 `sceneReady`
- Worker 3 完成图片生成后，状态变为 `imgReady`

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
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "id": "文档ID",
    "name": "文档名称",
    "status": "sceneReady",
    "created_at": "2024-10-24 12:00:00",
    "updated_at": "2024-10-24 12:00:00"
  }
}
```

**业务状态码**

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
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "id": "文档ID",
    "name": "新的文档名称",
    "status": "sceneReady",
    "created_at": "2024-10-24 12:00:00",
    "updated_at": "2024-10-24 12:30:00"
  }
}
```

**业务状态码**

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
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": null
}
```

**业务状态码**

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
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "documents": [
      {
        "id": "文档ID1",
        "name": "文档名称1",
        "status": "imgReady",
        "created_at": "2024-10-24 12:00:00",
        "updated_at": "2024-10-24 12:00:00"
      },
      {
        "id": "文档ID2",
        "name": "文档名称2",
        "status": "chapterReady",
        "created_at": "2024-10-24 13:00:00",
        "updated_at": "2024-10-24 13:00:00"
      }
    ]
  }
}
```

**业务状态码**

- `200`: 获取成功

**说明**

- 文档列表按更新时间倒序排列

---

### 章节管理 (chapters)

#### 6. 获取章节详情

获取文档中指定章节的详细信息。

**请求**

```
GET /v1/documents/:document_id/chapters/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |
| id | string | 章节ID |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "id": "章节ID",
    "index": 0,
    "document_id": "文档ID",
    "title": "章节标题",
    "content": "章节内容",
    "scene_ids": ["场景ID1", "场景ID2"],
    "created_at": "2024-10-24 12:00:00",
    "updated_at": "2024-10-24 12:00:00"
  }
}
```

**业务状态码**

- `200`: 获取成功
- `400`: 参数无效
- `500`: 获取失败

---

#### 7. 更新章节

更新章节的内容。

**请求**

```
PUT /v1/documents/:document_id/chapters/:id
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
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "id": "章节ID",
    "index": 0,
    "document_id": "文档ID",
    "title": "章节标题",
    "content": "新的章节内容",
    "scene_ids": ["场景ID1", "场景ID2"],
    "created_at": "2024-10-24 12:00:00",
    "updated_at": "2024-10-24 12:30:00"
  }
}
```

**业务状态码**

- `200`: 更新成功
- `400`: 请求参数错误
- `500`: 更新失败

---

#### 8. 删除章节

删除文档中的指定章节。

**请求**

```
DELETE /v1/documents/:document_id/chapters/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |
| id | string | 章节ID |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": null
}
```

**业务状态码**

- `200`: 删除成功
- `400`: 参数无效
- `500`: 删除失败

**说明**

- 删除章节时会先删除该章节下的所有场景，然后再删除章节本身
- 这是级联删除操作，确保数据一致性

---

#### 9. 获取章节列表

获取文档的所有章节列表。

**请求**

```
GET /v1/documents/:document_id/chapters
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "chapters": [
      {
        "id": "章节ID1",
        "index": 0,
        "document_id": "文档ID",
        "title": "第一章",
        "content": "章节内容1",
        "scene_ids": ["场景ID1", "场景ID2"],
        "created_at": "2024-10-24 12:00:00",
        "updated_at": "2024-10-24 12:00:00"
      },
      {
        "id": "章节ID2",
        "index": 1,
        "document_id": "文档ID",
        "title": "第二章",
        "content": "章节内容2",
        "scene_ids": ["场景ID3"],
        "created_at": "2024-10-24 12:00:00",
        "updated_at": "2024-10-24 12:00:00"
      }
    ]
  }
}
```

**业务状态码**

- `200`: 获取成功
- `400`: 参数无效

**说明**

- 章节列表按 index 升序排列
- 后续需要考虑分页功能

---

### 角色管理 (Roles)

#### 10. 获取文档角色列表

获取文档中的所有角色信息。

**请求**

```
GET /v1/documents/:document_id/roles
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "roles": [
      {
        "id": "角色ID1",
        "document_id": "文档ID",
        "name": "张三",
        "gender": "男",
        "character": "勇敢、正直",
        "appearance": "身材高大，浓眉大眼",
        "created_at": "2024-10-24 12:00:00",
        "updated_at": "2024-10-24 12:00:00"
      },
      {
        "id": "角色ID2",
        "document_id": "文档ID",
        "name": "李四",
        "gender": "女",
        "character": "聪明、机智",
        "appearance": "身材苗条，长发飘飘",
        "created_at": "2024-10-24 12:00:00",
        "updated_at": "2024-10-24 12:00:00"
      }
    ]
  }
}
```

**业务状态码**

- `200`: 获取成功
- `400`: 参数无效
- `500`: 获取失败

---

#### 11. 更新角色信息

更新指定角色的信息。

**请求**

```
PUT /v1/roles/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 角色ID |

**请求体**

```json
{
  "name": "张三",
  "gender": "男",
  "character": "勇敢、正直、善良",
  "appearance": "身材魁梧，浓眉大眼，面容刚毅"
}
```

**字段说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 角色名称，最大50字符 |
| gender | string | 是 | 性别，最大10字符 |
| character | string | 是 | 性格特点，最大500字符 |
| appearance | string | 是 | 外貌描述，最大500字符 |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "id": "角色ID",
    "document_id": "文档ID",
    "name": "张三",
    "gender": "男",
    "character": "勇敢、正直、善良",
    "appearance": "身材魁梧，浓眉大眼，面容刚毅",
    "created_at": "2024-10-24 12:00:00",
    "updated_at": "2024-10-24 12:30:00"
  }
}
```

**业务状态码**

- `200`: 更新成功
- `400`: 请求参数错误
- `404`: 角色不存在
- `500`: 更新失败

---

### 场景管理 (Scenes)

#### 12. 获取文档的所有场景

获取文档中的所有场景列表。

**请求**

```
GET /v1/documents/:document_id/scenes
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| document_id | string | 文档ID |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "scenes": [
      {
        "id": "场景ID1",
        "chapter_id": "章节ID1",
        "document_id": "文档ID",
        "index": 0,
        "content": "场景描述内容1",
        "image_url": "https://example.com/image1.jpg",
        "voice_url": "https://example.com/voice1.mp3",
        "created_at": "2024-10-24 12:00:00",
        "updated_at": "2024-10-24 12:00:00"
      },
      {
        "id": "场景ID2",
        "chapter_id": "章节ID1",
        "document_id": "文档ID",
        "index": 1,
        "content": "场景描述内容2",
        "image_url": "https://example.com/image2.jpg",
        "voice_url": "",
        "created_at": "2024-10-24 12:00:00",
        "updated_at": "2024-10-24 12:00:00"
      }
    ]
  }
}
```

**业务状态码**

- `200`: 获取成功
- `400`: 参数无效
- `500`: 获取失败

**说明**

- 场景列表按章节ID和场景索引升序排列
- 确保同一章节的场景按顺序展示

---

#### 13. 获取章节的场景列表

获取指定章节的所有场景。

**请求**

```
GET /v1/chapters/:chapter_id/scenes
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| chapter_id | string | 章节ID |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "scenes": [
      {
        "id": "场景ID1",
        "chapter_id": "章节ID",
        "document_id": "文档ID",
        "index": 0,
        "content": "场景描述内容",
        "image_url": "https://example.com/image.jpg",
        "voice_url": "https://example.com/voice.mp3",
        "created_at": "2024-10-24 12:00:00",
        "updated_at": "2024-10-24 12:00:00"
      }
    ]
  }
}
```

**业务状态码**

- `200`: 获取成功
- `400`: 参数无效
- `500`: 获取失败

---

#### 14. 更新场景内容

更新场景的描述内容，并立即重新生成图片和语音。

**请求**

```
PUT /v1/scenes/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 场景ID |

**请求体**

```json
{
  "content": "新的场景描述内容"
}
```

**字段说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | 是 | 场景描述内容，最大1000字符 |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": {
    "id": "场景ID",
    "chapter_id": "章节ID",
    "document_id": "文档ID",
    "index": 0,
    "content": "新的场景描述内容",
    "image_url": "https://example.com/new-image.jpg",
    "voice_url": "https://example.com/new-voice.mp3",
    "created_at": "2024-10-24 12:00:00",
    "updated_at": "2024-10-24 12:30:00"
  }
}
```

**业务状态码**

- `200`: 更新成功
- `400`: 请求参数错误
- `404`: 场景不存在
- `500`: 更新失败或图片/语音生成失败

**说明**

- 更新场景内容后，系统会立即调用阿里云百炼 API 重新生成图片和语音
- 整个过程可能需要几秒钟时间
- 如果图片或语音生成失败，整个更新操作会回滚
- 响应中返回的是更新后的场景信息，包含新生成的图片和语音 URL

---

#### 15. 删除场景

删除指定的场景。

**请求**

```
DELETE /v1/scenes/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 场景ID |

**响应**

```json
{
  "code": 200,
  "message": "",
  "reqid": "abc123-def456-ghi789",
  "data": null
}
```

**业务状态码**

- `200`: 删除成功
- `400`: 场景ID无效
- `404`: 场景不存在
- `500`: 删除失败

---

## 数据模型

### Document (文档)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 文档唯一标识，32位UUID |
| name | string | 文档名称，最大50字符 |
| status | string | 文档状态：`chapterReady` (章节就绪)、`roleReady` (角色就绪)、`sceneReady` (场景就绪)、`imgReady` (图片就绪) |
| created_at | string | 创建时间，格式：YYYY-MM-DD HH:MM:SS |
| updated_at | string | 更新时间，格式：YYYY-MM-DD HH:MM:SS |

### Chapter (章节)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 章节唯一标识，32位UUID |
| index | integer | 章节序号，从0开始 |
| document_id | string | 所属文档ID |
| title | string | 章节标题，最大100字符 |
| content | string | 章节内容，最大10000字符 |
| scene_ids | array | 场景ID列表 |
| created_at | string | 创建时间，格式：YYYY-MM-DD HH:MM:SS |
| updated_at | string | 更新时间，格式：YYYY-MM-DD HH:MM:SS |

### Scene (场景)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 场景唯一标识，32位UUID |
| chapter_id | string | 所属章节ID |
| document_id | string | 所属文档ID |
| index | integer | 场景序号，从0开始 |
| content | string | 场景描述内容，最大1000字符 |
| image_url | string | 场景图片URL，最大500字符 |
| voice_url | string | 音频URL，最大500字符 |
| created_at | string | 创建时间，格式：YYYY-MM-DD HH:MM:SS |
| updated_at | string | 更新时间，格式：YYYY-MM-DD HH:MM:SS |

### Role (角色)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 角色唯一标识，32位UUID |
| document_id | string | 所属文档ID |
| name | string | 角色名字，最大50字符 |
| gender | string | 性别，最大10字符 |
| character | string | 性格特点，最大500字符 |
| appearance | string | 外貌描述，最大500字符 |
| created_at | string | 创建时间，格式：YYYY-MM-DD HH:MM:SS |
| updated_at | string | 更新时间，格式：YYYY-MM-DD HH:MM:SS |

---

## 业务状态码

系统使用业务状态码（响应 body 中的 `code` 字段）来表示具体的业务处理结果：

| 状态码 | 说明 |
|--------|------|
| 200 | 业务处理成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 (当前未启用) |
| 403 | 禁止访问 (当前未启用) |
| 500 | 服务器内部错误 |
| 599 | 服务器内部错误 (默认错误码) |
| 612 | 文档不存在 |
| 614 | 文档已存在 |

**注意**

- HTTP 状态码始终为 200（表示请求被服务器接受）
- 业务处理是否成功通过响应 body 中的 `code` 字段判断
- `code` 为 200 表示业务成功，非 200 表示业务失败
- 失败时 `message` 字段包含详细错误信息
- 每个响应都包含 `reqid` 字段，用于请求追踪和日志关联

---