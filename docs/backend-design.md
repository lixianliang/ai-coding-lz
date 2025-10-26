# 后端详细设计文档

## 一、数据库设计

### 1.1 表结构设计

#### Document 表（文档表）

```go
type Document struct {
    ID        string    `gorm:"primaryKey;size:32;comment:'主键'"`
    Name      string    `gorm:"uniqueIndex:uk_name;size:128;comment:'文档名称'"`
    FileID    string    `gorm:"size:255;comment:'存储在阿里云百炼的 fileid'"`
    Status    string    `gorm:"size:20;comment:'状态'"`
    CreatedAt time.Time `gorm:"comment:'创建时间'"`
    UpdatedAt time.Time `gorm:"comment:'更新时间'"`
}
```

**字段说明：**
- `ID`: 32位 UUID，主键
- `Name`: 文档名称，唯一索引，最大128字符
- `FileID`: 阿里云百炼返回的文件ID，用于后续调用 qwen-long
- `Status`: 文档状态，取值：
  - `chapterReady`: 章节分割完成，等待角色提取
  - `roleReady`: 角色提取完成，等待场景提取
  - `sceneReady`: 场景提取完成，等待图片生成
  - `imgReady`: 图片生成完成
- `CreatedAt`: 创建时间
- `UpdatedAt`: 更新时间

**索引设计：**
- 主键索引：`id`
- 唯一索引：`uk_name` (name)
- 查询索引：`idx_status` (status) - 用于 worker 查询待处理文档

#### Chapter 表（章节表）

```go
type Chapter struct {
    ID         string    `gorm:"primaryKey;size:32;comment:'主键'"`
    Index      int       `gorm:"uniqueIndex:uk_document_index,priority:2;comment:'章节序号'"`
    DocumentID string    `gorm:"uniqueIndex:uk_document_index,priority:1;size:32;comment:'文档 id'"`
    Title      string    `gorm:"size:100;comment:'标题'"`
    Content    string    `gorm:"size:10000;comment:'章节内容'"`
    SceneIDs   []string  `gorm:"type:json;serializer:json;comment:'场景ID列表'"`
    CreatedAt  time.Time `gorm:"comment:'创建时间'"`
    UpdatedAt  time.Time `gorm:"comment:'更新时间'"`
}
```

**字段说明：**
- `ID`: 32位 UUID，主键
- `Index`: 章节序号，从0开始
- `DocumentID`: 所属文档ID，外键关联
- `Title`: 章节标题，可选，由章节分割时提取
- `Content`: 章节内容，最大10000字符
- `SceneIDs`: 场景ID数组，JSON 类型存储，如：`["id1", "id2", "id3"]`
- `CreatedAt`: 创建时间
- `UpdatedAt`: 更新时间

**索引设计：**
- 主键索引：`id`
- 联合唯一索引：`uk_document_index` (document_id, index) - 确保同一文档的章节序号唯一

#### Scene 表（场景表）

```go
type Scene struct {
    ID        string    `gorm:"primaryKey;size:32;comment:'主键'"`
    ChapterID string    `gorm:"index:idx_chapter_id;size:32;comment:'章节 id'"`
    DocumentID string   `gorm:"index:idx_document_id;size:32;comment:'文档 id'"`
    Index     int       `gorm:"comment:'场景序号'"`
    Content   string    `gorm:"size:1000;comment:'场景描述'"`
    ImageURL  string    `gorm:"size:200;comment:'场景图片url'"`
    VoiceURL  string    `gorm:"size:200;comment:'音频url'"`
    CreatedAt time.Time `gorm:"comment:'创建时间'"`
    UpdatedAt time.Time `gorm:"comment:'更新时间'"`
}
```

**字段说明：**
- `ID`: 32位 UUID，主键
- `ChapterID`: 所属章节ID，外键关联
- `DocumentID`: 所属文档ID，冗余字段，方便按文档查询所有场景
- `Index`: 场景在章节内的序号，从0开始
- `Content`: 场景描述文字，LLM 生成的适合画图的描述
- `ImageURL`: 阿里云百炼生成的图片URL
- `VoiceURL`: TTS 生成的音频URL（预留字段）
- `CreatedAt`: 创建时间
- `UpdatedAt`: 更新时间

**索引设计：**
- 主键索引：`id`
- 普通索引：`idx_chapter_id` (chapter_id)
- 普通索引：`idx_document_id` (document_id)

#### Role 表（角色表）

```go
type Role struct {
    ID         string    `gorm:"primaryKey;size:32;comment:'主键'"`
    DocumentID string    `gorm:"index:idx_document_id;size:32;comment:'文档 id'"`
    Name       string    `gorm:"size:50;comment:'角色名字'"`
    Gender     string    `gorm:"size:10;comment:'性别'"`
    Character  string    `gorm:"size:500;comment:'性格特点'"`
    Appearance string    `gorm:"size:500;comment:'外貌描述'"`
    CreatedAt  time.Time `gorm:"comment:'创建时间'"`
    UpdatedAt  time.Time `gorm:"comment:'更新时间'"`
}
```

**字段说明：**
- `ID`: 32位 UUID，主键
- `DocumentID`: 所属文档ID，外键关联
- `Name`: 角色名字
- `Gender`: 性别（男/女/未知）
- `Character`: 性格特点描述
- `Appearance`: 外貌特征描述
- `CreatedAt`: 创建时间
- `UpdatedAt`: 更新时间

**索引设计：**
- 主键索引：`id`
- 普通索引：`idx_document_id` (document_id)

### 1.2 ER 关系图（文字描述）

```
Document (1) ----< (N) Chapter
    |
    |----< (N) Scene (冗余关联)
    |
    |----< (N) Role

Chapter (1) ----< (N) Scene
```

**关系说明：**
1. 一个 Document 包含多个 Chapter（一对多）
2. 一个 Document 包含多个 Scene（一对多，通过冗余 document_id 查询）
3. 一个 Document 包含多个 Role（一对多）
4. 一个 Chapter 包含多个 Scene（一对多，0-3个）
5. Chapter.SceneIDs 存储场景ID数组，方便按顺序查询

### 1.3 数据库 DAO 接口

#### Document DAO

```go
// 现有方法
CreateDocument(ctx, docID, args) (*Document, error)
GetDocument(ctx, id) (Document, error)
GetDocumentWithName(ctx, name) (Document, error)
UpdateDocument(ctx, id, args) error
UpdateDocumentStatus(ctx, id, status) error
DeleteDocument(ctx, id) error
ListDocuments(ctx) ([]Document, error)

// 新增方法
UpdateDocumentFileID(ctx, id, fileID) error  // 更新 FileID
ListChapterReadyDocuments(ctx) ([]Document, error)  // 查询 chapterReady 状态
ListSceneReadyDocuments(ctx) ([]Document, error)    // 查询 sceneReady 状态
```

#### Chapter DAO

```go
// 现有方法
CreateChapters(ctx, documentID, texts) error
GetChapter(ctx, id, documentID) (Chapter, error)
UpdateChapter(ctx, id, args) error
DeleteChapter(ctx, id, documentID) error
DeleteAllChapter(ctx, documentID) error
ListChapters(ctx, documentID) ([]Chapter, error)

// 新增方法
UpdateChapterSceneIDs(ctx, chapterID, sceneIDs) error  // 更新场景ID列表
```

#### Scene DAO（新增）

```go
CreateScenes(ctx, scenes []Scene) error
GetScene(ctx, id) (Scene, error)
ListScenesByChapter(ctx, chapterID) ([]Scene, error)
ListScenesByDocument(ctx, documentID) ([]Scene, error)
ListPendingImageScenes(ctx, documentID) ([]Scene, error)  // 查询未生成图片的场景
UpdateSceneImageURL(ctx, sceneID, imageURL) error
DeleteScenesByChapter(ctx, chapterID) error
```

#### Role DAO（新增）

```go
CreateRoles(ctx, roles []Role) error
GetRole(ctx, id) (Role, error)
ListRolesByDocument(ctx, documentID) ([]Role, error)
DeleteRolesByDocument(ctx, documentID) error
```

## 二、阿里云百炼集成设计

### 2.1 包结构

创建 `imgagent/bailian/` 包，封装阿里云百炼 API 调用。

```
imgagent/
└── bailian/
    ├── bailian.go      // 主客户端
    ├── file.go         // 文件上传
    ├── qwen_long.go    // qwen-long 调用
    ├── qwen_image.go   // qwen-image-plus 调用
    └── types.go        // 类型定义
```

### 2.2 配置结构

```go
type Config struct {
    BaseURL           string `json:"base_url"`
    APIKey            string `json:"api_key"`
    RolePrompt        string `json:"role_prompt"`         // 角色提取 Prompt
    ScenePrompt       string `json:"scene_prompt"`        // 场景生成 Prompt
    ImageSize         string `json:"image_size"`          // 图片尺寸，默认 "1328*1328"
    ImageWatermark    bool   `json:"image_watermark"`     // 是否添加水印，默认 true
    RequestTimeout    int    `json:"request_timeout"`     // 请求超时时间（秒），默认 300
    MaxRetries        int    `json:"max_retries"`         // 最大重试次数，默认 0
}
```

### 2.3 客户端结构

```go
type Client struct {
    config     Config
    httpClient *http.Client
    logger     *zap.SugaredLogger
}

func NewClient(config Config) (*Client, error) {
    // 设置默认值
    if config.BaseURL == "" {
        config.BaseURL = "https://dashscope.aliyuncs.com"
    }
    if config.ImageSize == "" {
        config.ImageSize = "1328*1328"
    }
    if config.RequestTimeout == 0 {
        config.RequestTimeout = 300
    }
    
    // 创建 HTTP 客户端
    httpClient := &http.Client{
        Timeout: time.Duration(config.RequestTimeout) * time.Second,
    }
    
    return &Client{
        config:     config,
        httpClient: httpClient,
        logger:     zap.S().Named("bailian"),
    }, nil
}
```

### 2.4 核心方法设计

#### 2.4.1 文件上传

```go
// UploadFile 上传文件到阿里云百炼
// 返回 fileID 用于后续 qwen-long 调用
func (c *Client) UploadFile(ctx context.Context, filename string) (string, error)
```

**实现要点：**
- URL: `POST /compatible-mode/v1/files`
- Headers: `Authorization: Bearer {APIKey}`
- Form Data: 
  - `file`: 文件内容
  - `purpose`: "file-extract"
- 返回格式：
```json
{
    "id": "file-fe-xxx",
    "object": "file",
    "created_at": 1234567890
}
```

**错误处理：**
- 文件不存在：返回错误
- 上传失败：记录日志并返回错误
- 解析响应失败：返回错误

#### 2.4.2 角色提取

```go
// ExtractRoles 从文档中提取角色信息
// 使用 qwen-long 分析整个文档
func (c *Client) ExtractRoles(ctx context.Context, fileID string) ([]RoleInfo, error)

type RoleInfo struct {
    Name       string `json:"name"`
    Gender     string `json:"gender"`
    Character  string `json:"character"`
    Appearance string `json:"appearance"`
}
```

**默认 Prompt：**
```
请仔细分析这篇小说，提取出所有主要人物角色的信息。对每个角色，请提供：
1. 姓名（name）
2. 性别（gender）：男/女/未知
3. 性格特点（character）：简要描述角色的性格特征
4. 外貌描述（appearance）：描述角色的外貌特征，用于生成角色画像

要求：
- 只提取主要角色（出场次数较多或对情节有重要影响的角色）
- 每个角色的描述要简洁准确
- 如果信息不明确，可以标注为"未知"或省略
- 严格按照 JSON 数组格式返回，不要有其他文字说明

返回格式示例：
[
    {
        "name": "张三",
        "gender": "男",
        "character": "勇敢、正直、善良",
        "appearance": "身材魁梧，浓眉大眼，面容刚毅"
    }
]
```

**实现要点：**
- URL: `POST /compatible-mode/v1/chat/completions`
- Request Body:
```json
{
    "model": "qwen-long",
    "messages": [
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "system", "content": "fileid://{fileID}"},
        {"role": "user", "content": "{RolePrompt}"}
    ],
    "stream": false
}
```
- 解析响应中的 JSON 内容
- 如果 LLM 返回非 JSON 格式，尝试提取 JSON 部分

**错误处理：**
- API 调用失败：返回错误，不重试（由 worker 重试）
- JSON 解析失败：记录原始响应，返回错误
- 返回空数组：正常情况，返回空列表

#### 2.4.3 场景生成

```go
// GenerateScenes 为章节生成场景描述
// 每章生成 0-3 个场景
func (c *Client) GenerateScenes(ctx context.Context, chapterContent string) ([]string, error)
```

**默认 Prompt：**
```
请将以下章节内容拆分为 0-3 个关键场景，用于生成连环漫画。

要求：
1. 每个场景用一句话描述，适合作为画面生成的提示词
2. 场景要能体现章节的关键情节或重要时刻
3. 如果章节内容较少或不适合拆分场景，可以返回空数组
4. 每个场景描述要包含：地点、人物、动作/事件、氛围
5. 场景描述要具体、形象，便于理解和画图
6. 严格返回 JSON 数组格式，每个元素是一个场景描述字符串
7. 最多返回 3 个场景

章节内容：
{chapterContent}

返回格式示例：
["场景1的描述文字", "场景2的描述文字", "场景3的描述文字"]
```

**实现要点：**
- URL: `POST /compatible-mode/v1/chat/completions`
- Request Body:
```json
{
    "model": "qwen-long",
    "messages": [
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "{ScenePrompt with chapterContent}"}
    ],
    "stream": false
}
```
- 解析返回的 JSON 数组
- 限制最多返回 3 个场景

**错误处理：**
- API 调用失败：返回错误
- JSON 解析失败：记录原始响应，返回空数组
- 返回超过 3 个场景：只取前 3 个

#### 2.4.4 图片生成

```go
// GenerateImage 根据场景描述生成图片
// 返回图片 URL
func (c *Client) GenerateImage(ctx context.Context, sceneContent string) (string, error)
```

**实现要点：**
- URL: `POST /api/v1/services/aigc/multimodal-generation/generation`
- Headers: 
  - `Content-Type: application/json`
  - `Authorization: Bearer {APIKey}`
- Request Body:
```json
{
    "model": "qwen-image-plus",
    "input": {
        "messages": [
            {
                "role": "user",
                "content": [
                    {"text": "{sceneContent}"}
                ]
            }
        ]
    },
    "parameters": {
        "negative_prompt": "",
        "prompt_extend": true,
        "watermark": true,
        "size": "1328*1328"
    }
}
```
- 响应格式：
```json
{
    "output": {
        "results": [
            {
                "url": "https://..."
            }
        ]
    }
}
```

**错误处理：**
- API 调用失败：返回错误
- 响应格式错误：返回错误
- URL 为空：返回错误

### 2.5 错误处理和重试策略

**原则：不在客户端层重试，由上层 Worker 控制重试**

- API 调用失败：直接返回错误，记录详细日志
- 超时：返回超时错误
- 响应解析失败：返回解析错误，保留原始响应内容
- 业务错误：返回业务错误信息

**日志记录：**
- 请求开始：记录 API、参数摘要
- 请求成功：记录耗时、响应摘要
- 请求失败：记录错误详情、请求参数（脱敏）

## 三、异步任务管理器设计

### 3.1 配置结构

```go
type DocumentConfig struct {
    Enable                       bool `json:"enable"`                          // 是否启用异步任务
    HandleSceneIntervalSecs      int  `json:"handle_scene_interval_secs"`      // 场景提取轮询间隔（秒）
    HandleImageGenIntervalSecs   int  `json:"handle_image_gen_interval_secs"`  // 图片生成轮询间隔（秒）
}
```

### 3.2 管理器结构

```go
type DocumentMgr struct {
    config        DocumentConfig
    db            db.IDataBase
    bailianClient *bailian.Client
    close         chan bool
}

func NewDocumentMgr(config DocumentConfig, db db.IDataBase, bailianClient *bailian.Client) (*DocumentMgr, error) {
    if config.HandleSceneIntervalSecs == 0 {
        config.HandleSceneIntervalSecs = 30
    }
    if config.HandleImageGenIntervalSecs == 0 {
        config.HandleImageGenIntervalSecs = 30
    }
    
    return &DocumentMgr{
        config:        config,
        db:            db,
        bailianClient: bailianClient,
        close:         make(chan bool),
    }, nil
}
```

### 3.3 Worker 架构

#### Worker 1: 场景提取 Worker

**启动方式：**
```go
func (m *DocumentMgr) Run() {
    if !m.config.Enable {
        return
    }
    go m.loopHandleSceneTasks()
    go m.loopHandleImageGenTasks()
}

func (m *DocumentMgr) loopHandleSceneTasks() {
    ticker := time.NewTicker(time.Duration(m.config.HandleSceneIntervalSecs) * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ctx := logger.NewContext(fmt.Sprintf("HandleSceneTasks-%d", time.Now().Unix()))
            m.HandleSceneTasks(ctx)
        case <-m.close:
            return
        }
    }
}
```

**处理流程：**
```go
func (m *DocumentMgr) HandleSceneTasks(ctx context.Context) {
    log := logger.FromContext(ctx)
    
    // 1. 查询 chapterReady 状态的文档
    docs, err := m.db.ListChapterReadyDocuments(ctx)
    if err != nil {
        log.Errorf("Failed to list chapterReady documents, err: %v", err)
        return
    }
    
    // 2. 逐个处理文档
    for _, doc := range docs {
        if err := m.HandleDocumentScene(ctx, doc); err != nil {
            log.Errorf("Failed to handle document scene, doc: %v, err: %v", doc.ID, err)
            continue  // 失败保持状态，下次继续处理
        }
        
        // 3. 更新文档状态为 sceneReady
        if err := m.db.UpdateDocumentStatus(ctx, doc.ID, db.DocumentStatusSceneReady); err != nil {
            log.Errorf("Failed to update document status, doc: %v, err: %v", doc.ID, err)
        }
    }
}

func (m *DocumentMgr) HandleDocumentScene(ctx context.Context, doc db.Document) error {
    log := logger.FromContext(ctx)
    
    // 1. 如果 FileID 为空，上传文件到阿里云百炼
    if doc.FileID == "" {
        // 需要从 temp 目录找到对应文件（根据 doc.ID）
        filename := fmt.Sprintf("./temp/%s.txt", doc.ID)  // 需要调整
        fileID, err := m.bailianClient.UploadFile(ctx, filename)
        if err != nil {
            log.Errorf("Failed to upload file, doc: %v, err: %v", doc.ID, err)
            return err
        }
        
        // 更新 FileID
        if err := m.db.UpdateDocumentFileID(ctx, doc.ID, fileID); err != nil {
            log.Errorf("Failed to update fileID, doc: %v, err: %v", doc.ID, err)
            return err
        }
        doc.FileID = fileID
    }
    
    // 2. 提取角色信息
    roles, err := m.bailianClient.ExtractRoles(ctx, doc.FileID)
    if err != nil {
        log.Errorf("Failed to extract roles, doc: %v, err: %v", doc.ID, err)
        return err
    }
    
    // 保存角色到数据库
    if len(roles) > 0 {
        dbRoles := make([]db.Role, 0, len(roles))
        now := time.Now()
        for _, r := range roles {
            dbRoles = append(dbRoles, db.Role{
                ID:         db.MakeUUID(),
                DocumentID: doc.ID,
                Name:       r.Name,
                Gender:     r.Gender,
                Character:  r.Character,
                Appearance: r.Appearance,
                CreatedAt:  now,
                UpdatedAt:  now,
            })
        }
        if err := m.db.CreateRoles(ctx, dbRoles); err != nil {
            log.Errorf("Failed to create roles, doc: %v, err: %v", doc.ID, err)
            return err
        }
    }
    
    // 3. 获取所有章节
    chapters, err := m.db.ListChapters(ctx, doc.ID)
    if err != nil {
        log.Errorf("Failed to list chapters, doc: %v, err: %v", doc.ID, err)
        return err
    }
    
    // 4. 为每个章节生成场景
    for _, chapter := range chapters {
        scenes, err := m.bailianClient.GenerateScenes(ctx, chapter.Content)
        if err != nil {
            log.Errorf("Failed to generate scenes, chapter: %v, err: %v", chapter.ID, err)
            return err
        }
        
        // 保存场景到数据库
        if len(scenes) > 0 {
            dbScenes := make([]db.Scene, 0, len(scenes))
            sceneIDs := make([]string, 0, len(scenes))
            now := time.Now()
            
            for i, sceneContent := range scenes {
                sceneID := db.MakeUUID()
                sceneIDs = append(sceneIDs, sceneID)
                dbScenes = append(dbScenes, db.Scene{
                    ID:         sceneID,
                    ChapterID:  chapter.ID,
                    DocumentID: doc.ID,
                    Index:      i,
                    Content:    sceneContent,
                    CreatedAt:  now,
                    UpdatedAt:  now,
                })
            }
            
            if err := m.db.CreateScenes(ctx, dbScenes); err != nil {
                log.Errorf("Failed to create scenes, chapter: %v, err: %v", chapter.ID, err)
                return err
            }
            
            // 更新 Chapter 的 SceneIDs
            if err := m.db.UpdateChapterSceneIDs(ctx, chapter.ID, sceneIDs); err != nil {
                log.Errorf("Failed to update chapter sceneIDs, chapter: %v, err: %v", chapter.ID, err)
                return err
            }
        }
    }
    
    return nil
}
```

#### Worker 2: 图片生成 Worker

**启动方式：**
```go
func (m *DocumentMgr) loopHandleImageGenTasks() {
    ticker := time.NewTicker(time.Duration(m.config.HandleImageGenIntervalSecs) * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ctx := logger.NewContext(fmt.Sprintf("HandleImageGenTasks-%d", time.Now().Unix()))
            m.HandleImageGenTasks(ctx)
        case <-m.close:
            return
        }
    }
}
```

**处理流程：**
```go
func (m *DocumentMgr) HandleImageGenTasks(ctx context.Context) {
    log := logger.FromContext(ctx)
    
    // 1. 查询 sceneReady 状态的文档
    docs, err := m.db.ListSceneReadyDocuments(ctx)
    if err != nil {
        log.Errorf("Failed to list sceneReady documents, err: %v", err)
        return
    }
    
    // 2. 逐个处理文档
    for _, doc := range docs {
        if err := m.HandleDocumentImageGen(ctx, doc); err != nil {
            log.Errorf("Failed to handle document image gen, doc: %v, err: %v", doc.ID, err)
            continue  // 失败保持状态，下次继续处理
        }
        
        // 3. 更新文档状态为 imgReady
        if err := m.db.UpdateDocumentStatus(ctx, doc.ID, db.DocumentStatusImgReady); err != nil {
            log.Errorf("Failed to update document status, doc: %v, err: %v", doc.ID, err)
        }
    }
}

func (m *DocumentMgr) HandleDocumentImageGen(ctx context.Context, doc db.Document) error {
    log := logger.FromContext(ctx)
    
    // 1. 获取所有未生成图片的场景
    scenes, err := m.db.ListPendingImageScenes(ctx, doc.ID)
    if err != nil {
        log.Errorf("Failed to list pending image scenes, doc: %v, err: %v", doc.ID, err)
        return err
    }
    
    if len(scenes) == 0 {
        // 所有场景都已生成图片
        return nil
    }
    
    // 2. 为每个场景生成图片
    for _, scene := range scenes {
        imageURL, err := m.bailianClient.GenerateImage(ctx, scene.Content)
        if err != nil {
            log.Errorf("Failed to generate image, scene: %v, err: %v", scene.ID, err)
            return err  // 失败则整个文档重试
        }
        
        // 更新场景图片 URL
        if err := m.db.UpdateSceneImageURL(ctx, scene.ID, imageURL); err != nil {
            log.Errorf("Failed to update scene imageURL, scene: %v, err: %v", scene.ID, err)
            return err
        }
        
        log.Infof("Generated image for scene, doc: %v, scene: %v, url: %v", doc.ID, scene.ID, imageURL)
    }
    
    return nil
}
```

### 3.4 状态流转图

```
文档上传
    ↓
[chapterReady]  ← 章节分割完成
    ↓
    ↓ (Worker 1: 角色提取)
    ↓ - 提取角色信息
    ↓ - 保存到数据库
    ↓
[roleReady]  ← 角色提取完成
    ↓
    ↓ (Worker 2: 场景生成)
    ↓ - 为每章节生成场景
    ↓ - 保存场景到数据库
    ↓
[sceneReady]  ← 场景提取完成
    ↓
    ↓ (Worker 3: 图片生成)
    ↓ - 为每个场景生成图片
    ↓ - 更新场景图片URL
    ↓
[imgReady]  ← 图片生成完成（最终状态）
```

### 3.5 错误处理机制

**原则：失败保持当前状态，下次轮询继续处理**

1. **任务级别失败**：
   - 某个文档处理失败，不影响其他文档
   - 记录错误日志，继续处理下一个文档
   - 失败的文档保持当前状态，下次轮询继续尝试

2. **步骤级别失败**：
   - 文档处理的某个步骤失败，整个文档处理流程中断
   - 已完成的步骤数据保留（如已保存的角色、场景）
   - 下次轮询时会重新执行（需要处理幂等性）

3. **幂等性处理**：
   - 角色提取：删除已有角色，重新提取
   - 场景生成：检查 Chapter.SceneIDs，如果已有则跳过
   - 图片生成：只处理 ImageURL 为空的场景

4. **日志记录**：
   - 每个文档处理开始/结束记录
   - 每个步骤的成功/失败记录
   - 错误详情记录（包含文档ID、章节ID等上下文）

## 四、API 接口设计

### 4.1 修改现有接口

#### POST /v1/documents

**修改点：**
1. 文件上传后直接创建 Document 记录，状态设为 `chapterReady`
2. 同步进行章节分割并保存 Chapter
3. 不再上传到阿里云（由 Worker 1 异步处理）
4. 保存原始文件副本到 temp 目录（命名：`{docID}.{ext}`）

**响应：**
```json
{
    "code": 200,
    "data": {
        "id": "xxx",
        "name": "金庸-天龙八部",
        "status": "chapterReady",
        "created_at": "2024-01-01 12:00:00",
        "updated_at": "2024-01-01 12:00:00"
    }
}
```

### 4.2 新增接口

#### GET /v1/documents/:document_id/roles

获取文档的角色列表。

**请求：**
```
GET /v1/documents/{document_id}/roles
```

**响应：**
```json
{
    "code": 200,
    "data": {
        "roles": [
            {
                "id": "xxx",
                "document_id": "xxx",
                "name": "令狐冲",
                "gender": "男",
                "character": "洒脱不羁，重情重义",
                "appearance": "身材修长，相貌英俊",
                "created_at": "2024-01-01 12:00:00",
                "updated_at": "2024-01-01 12:00:00"
            }
        ]
    }
}
```

**错误码：**
- 400: 文档ID无效
- 612: 文档不存在

#### GET /v1/documents/:document_id/scenes

获取文档的所有场景列表。

**请求：**
```
GET /v1/documents/{document_id}/scenes
```

**响应：**
```json
{
    "code": 200,
    "data": {
        "scenes": [
            {
                "id": "xxx",
                "chapter_id": "xxx",
                "document_id": "xxx",
                "index": 0,
                "content": "华山之巅，剑气纵横，令狐冲手持长剑面对强敌",
                "image_url": "https://...",
                "voice_url": "",
                "created_at": "2024-01-01 12:00:00",
                "updated_at": "2024-01-01 12:00:00"
            }
        ]
    }
}
```

**错误码：**
- 400: 文档ID无效
- 612: 文档不存在

#### GET /v1/chapters/:chapter_id/scenes

获取章节的场景列表。

**请求：**
```
GET /v1/chapters/{chapter_id}/scenes
```

**查询参数：**
- `document_id` (必填): 文档ID，用于验证权限

**响应：**
```json
{
    "code": 200,
    "data": {
        "scenes": [
            {
                "id": "xxx",
                "chapter_id": "xxx",
                "document_id": "xxx",
                "index": 0,
                "content": "华山之巅，剑气纵横",
                "image_url": "https://...",
                "voice_url": "",
                "created_at": "2024-01-01 12:00:00",
                "updated_at": "2024-01-01 12:00:00"
            }
        ]
    }
}
```

**错误码：**
- 400: 参数无效
- 612: 章节不存在

### 4.3 API 类型定义

在 `imgagent/api/document.go` 中添加：

```go
// Role 角色信息
type Role struct {
    ID         string `json:"id"`
    DocumentID string `json:"document_id"`
    Name       string `json:"name"`
    Gender     string `json:"gender"`
    Character  string `json:"character"`
    Appearance string `json:"appearance"`
    CreatedAt  string `json:"created_at"`
    UpdatedAt  string `json:"updated_at"`
}

// Scene 场景信息
type Scene struct {
    ID         string `json:"id"`
    ChapterID  string `json:"chapter_id"`
    DocumentID string `json:"document_id"`
    Index      int    `json:"index"`
    Content    string `json:"content"`
    ImageURL   string `json:"image_url"`
    VoiceURL   string `json:"voice_url"`
    CreatedAt  string `json:"created_at"`
    UpdatedAt  string `json:"updated_at"`
}

// ListRolesResult 角色列表响应
type ListRolesResult struct {
    Roles []Role `json:"roles"`
}

// ListScenesResult 场景列表响应
type ListScenesResult struct {
    Scenes []Scene `json:"scenes"`
}
```

### 4.4 路由注册

在 `imgagent/svr/svr.go` 的 `RegisterRouter` 方法中添加：

```go
// Role
authGroup.GET("/documents/:document_id/roles", s.HandleGetRoles)

// Scene
authGroup.GET("/documents/:document_id/scenes", s.HandleListScenesByDocument)
authGroup.GET("/chapters/:chapter_id/scenes", s.HandleListScenesByChapter)
```

## 五、配置文件设计

### 5.1 imgagent.json 完整配置

```json
{
    "log_conf": {
        "level": "debug",
        "file": "",
        "access_file": "",
        "rotation": {
            "max_size": 100,
            "max_backups": 10,
            "max_age": 30
        }
    },
    "bind_host": ":8000",
    "api_version": "/v1",
    "temp": "./temp",
    "db": {
        "host": "localhost",
        "port": 3306,
        "user": "root",
        "password": "123456",
        "database": "imgagent",
        "enable_log": true
    },
    "storage": {
        "bucket": "bucket1",
        "domain": "bucket1.com",
        "ak": "xxx",
        "sk": "xxx"
    },
    "bailian": {
        "base_url": "https://dashscope.aliyuncs.com",
        "api_key": "sk-xxxxxxxxxxxx",
        "role_prompt": "",
        "scene_prompt": "",
        "image_size": "1328*1328",
        "image_watermark": true,
        "request_timeout": 300,
        "max_retries": 0
    },
    "document_mgr": {
        "enable": true,
        "handle_scene_interval_secs": 30,
        "handle_image_gen_interval_secs": 30
    }
}
```

### 5.2 配置项说明

#### bailian 配置段

- `base_url`: 阿里云百炼 API 基础 URL
- `api_key`: API 密钥
- `role_prompt`: 角色提取 Prompt（可选，为空则使用默认）
- `scene_prompt`: 场景生成 Prompt（可选，为空则使用默认）
- `image_size`: 生成图片尺寸，默认 "1328*1328"
- `image_watermark`: 是否添加水印，默认 true
- `request_timeout`: 请求超时时间（秒），默认 300
- `max_retries`: 客户端最大重试次数，默认 0（不重试，由 Worker 控制）

#### document_mgr 配置段

- `enable`: 是否启用异步任务管理器
- `handle_scene_interval_secs`: 场景提取轮询间隔（秒）
- `handle_image_gen_interval_secs`: 图片生成轮询间隔（秒）

### 5.3 配置加载

修改 `imgagent/main.go` 的 Config 结构：

```go
type Config struct {
    LogConf      logger.Config      `json:"log_conf"`
    BindHost     string             `json:"bind_host"`
    BailianConf  bailian.Config     `json:"bailian"`
    DocumentMgr  svr.DocumentConfig `json:"document_mgr"`
    
    svr.Config
}
```

修改 `imgagent/svr/svr.go` 的 Config 和 Service：

```go
type Config struct {
    APIVersion     string                `json:"api_version"`
    Temp           string                `json:"temp"`
    Storage        storage.Config        `json:"storage"`
    DB             dbutil.Config         `json:"db"`
    BailianConfig  bailian.Config        `json:"-"`  // 从外部传入
    DocumentConfig DocumentConfig        `json:"-"`  // 从外部传入
}

type Service struct {
    conf          Config
    db            db.IDataBase
    stg           *storage.Storage
    bailianClient *bailian.Client
    documentMgr   *DocumentMgr
}
```

## 六、实施注意事项

### 6.1 数据库迁移

- 新增 Scene 和 Role 表
- Document 表添加 FileID 字段
- Chapter 表修改 SceneIDs 字段类型为 JSON
- 添加必要的索引

### 6.2 文件管理

- 上传的原始文件保存到 temp 目录
- 文件命名：`{docID}.{ext}`
- Worker 处理完成后可选择是否删除（建议保留）

### 6.3 并发控制

**说明：** 服务端部署为单节点，无并发冲突问题

- Worker 轮询使用串行处理（同一时刻只处理一个文档）
- 数据库操作使用事务（必要时）
- 无需乐观锁机制

### 6.4 性能优化

- 批量插入场景和角色
- 缓存文档状态（可选）
- 图片生成可考虑并发（后期优化）

### 6.5 监控和告警

- 记录处理耗时
- 统计成功/失败率
- 监控队列长度（待处理文档数）
- 设置告警阈值

## 七、测试要点

### 7.1 单元测试

- 阿里云百炼客户端各方法
- DAO 层各方法
- Worker 处理逻辑（使用 mock）

### 7.2 集成测试

- 完整的文档处理流程
- 状态流转正确性
- 错误恢复机制

### 7.3 边界测试

- 空章节内容
- LLM 返回空结果
- API 调用失败
- 超大文件处理

