# imgagent 设计

## 使用流程
- 用户在界面通过 POST /v1/documents 上传小说，服务端进行章节分割同时将数据写入数据库，此时界面文档处理上传完成（chapterReady）
  - 服务端需要将 file 上传到 阿里云百炼
- 当文档处于 ready 状态（场景抽取已经完成），则用户可以进行动漫
- document_mgr.go 用于异步任务处理
  - 一个 worker 处理 chapterReady 状态的文档
    - 获取文档的 fileid, 调用阿里百炼 qwen-long 抽取人物信息，可以存储到 Role 表
    - 从 db list intited 文档，根据 docID 读取所有章节，每个章节调用阿里云百炼 qwen-long 生成场景，并存储到 Scence 表
    - 对应 doc 状态设置为 scenceReady 状态
  - 另外一个 worker 处理 scenceReady 状态文档
    - 从 db 获取文档所有场景和人物角色，依次调用 阿里云百炼 qwen-image-plus 生成图片，返回的 url 存储到数据库
    - 将 doc 的状态修改为 imgReady

## 阿里云百炼调用 api
```shell
阿里云百炼 file 上传：
curl --location --request POST 'https://dashscope.aliyuncs.com/compatible-mode/v1/files' \
  --header "Authorization: Bearer $DASHSCOPE_API_KEY" \
  --form 'file=@"阿里云百炼系列手机产品介绍.docx"' \
  --form 'purpose="file-extract"'
  
  
qwen-long  使用：
curl --location 'https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions' \
--header "Authorization: Bearer $DASHSCOPE_API_KEY" \
--header "Content-Type: application/json" \
--data '{
    "model": "qwen-long",
    "messages": [
        {"role": "system","content": "You are a helpful assistant."},
        {"role": "system","content": "fileid://file-fe-xxx"},
        {"role": "user","content": "这篇文章讲了什么？"}
    ],
    "stream": true,
    "stream_options": {
        "include_usage": true
    }
}'

qwen-image-plus 使用：
curl --location 'https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $DASHSCOPE_API_KEY" \
--data '{
    "model": "qwen-image-plus",
    "input": {
        "messages": [
            {
                "role": "user",
                "content": [
                    {
                        "text": "一副典雅庄重的对联悬挂于厅堂之中，房间是个安静古典的中式布置，桌子上放着一些青花瓷，对联上左书“义本生知人机同道善思新”，右书“通云赋智乾坤启数高志远”， 横批“智启通义”，字体飘逸，中间挂在一着一副中国风的画作，内容是岳阳楼。"
                    }
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
}'
```


## 数据库
```go
// Document 文档表
type Document struct {
ID        string    `gorm:"primaryKey;size:32;comment:'主键'"`
Name      string    `gorm:"uniqueIndex:uk_name;size:128;comment:'文档名称'"`
FileID    string    `gorm:"size:255;comment:'存储在阿里云百炼的 fileid'"`
Status    string    `gorm:"size:20;comment:'状态 indexing|ready'"`
CreatedAt time.Time `gorm:"comment:'创建时间'"`
UpdatedAt time.Time `gorm:"comment:'更新时间'"`
}

func (Document) TableName() string {
return "documents"
}

// Chapter 章节表
type Chapter struct {
	ID         string    `gorm:"primaryKey;size:32;comment:'主键'"`
	Index      int       `gorm:"uniqueIndex:uk_document_index,priority:2;comment:'章节序号'"`
	DocumentID string    `gorm:"uniqueIndex:uk_document_index,priority:1;size:32;comment:'文档 id'"`
	Title      string    `gorm:"size:100;comment:'标题'"`
	Content    string    `gorm:"size:10000;comment:'章节内容'"`
	ScenceIDs  []string  `gorm:"type:json;serializer:json;comment:'故事场景'"`
	CreatedAt  time.Time `gorm:"comment:'创建时间'"`
	UpdatedAt  time.Time `gorm:"comment:'更新时间'"`
}

func (Chapter) TableName() string {
	return "chapters"
}

// Scence 场景表
type Scence struct {
	ID        string    `gorm:"primaryKey;size:32;comment:'主键'"`
	ChapterID string    `gorm:"index:idx_chapter_index;size:32;comment:'chapter id'"`
	Content   string    `gorm:"size:1000;comment:'章节内容'"`
	ImageURL  string    `gorm:"size:200;comment:'场景图片url'"`
	VoiceURL  string    `gorm:"size:200;comment:'音频url'"`
	CreatedAt time.Time `gorm:"comment:'创建时间'"`
	UpdatedAt time.Time `gorm:"comment:'更新时间'"`
}

func (Scence) TableName() string {
	return "scences"
}

// Role 任务角色标
type Role struct {
	ID         string    `gorm:"primaryKey;size:32;comment:'主键'"`
    DocumentID string    `gorm:"index:idx_document_index;size:32;comment:'doc id'"`
	Name       string    // 名字
	Gender     string    // 性别
	Character  string    // 性格
	Appearance string    // 外貌
	CreatedAt  time.Time `gorm:"comment:'创建时间'"`
	UpdatedAt  time.Time `gorm:"comment:'更新时间'"`
}
```