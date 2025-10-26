package bailian

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Config 阿里云百炼配置
type Config struct {
	BaseURL        string `json:"base_url"`        // API 基础 URL
	APIKey         string `json:"api_key"`         // API 密钥
	SummaryPrompt  string `json:"summary_prompt"`  // 摘要提取 Prompt
	RolePrompt     string `json:"role_prompt"`     // 角色提取 Prompt
	ScenePrompt    string `json:"scene_prompt"`    // 场景生成 Prompt
	ImageSize      string `json:"image_size"`      // 图片尺寸
	ImageWatermark bool   `json:"image_watermark"` // 是否添加水印
	RequestTimeout int    `json:"request_timeout"` // 请求超时时间（秒）
	MaxRetries     int    `json:"max_retries"`     // 最大重试次数
}

// Client 阿里云百炼客户端
type Client struct {
	config     Config
	httpClient *http.Client
	logger     *zap.SugaredLogger
}

// NewClient 创建新的百炼客户端
func NewClient(config Config) (*Client, error) {
	// 设置默认值
	if config.BaseURL == "" {
		config.BaseURL = "https://dashscope.aliyuncs.com"
	}
	if config.ImageSize == "" {
		config.ImageSize = "1328*1328"
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 300 // 5分钟
	}

	// 设置默认 Prompt
	if config.SummaryPrompt == "" {
		config.SummaryPrompt = defaultSummaryPrompt
	}
	if config.RolePrompt == "" {
		config.RolePrompt = defaultRolePrompt
	}
	if config.ScenePrompt == "" {
		config.ScenePrompt = defaultScenePrompt
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

// 默认角色提取 Prompt
const defaultRolePrompt = `请仔细分析这篇小说，提取出所有主要人物角色的信息。对每个角色，请提供：
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
]`

// 默认摘要提取 Prompt
const defaultSummaryPrompt = `请为这篇小说生成一段简洁的摘要，用于辅助场景图片生成。

要求：
1. 摘要应包含：故事背景、主要情节线、核心冲突、整体风格/氛围
2. 重点描述小说的视觉风格、时代背景、场景特点等有助于图片生成的信息
3. 控制在 200-300 字以内
4. 使用简洁、客观的语言
5. 直接返回摘要文本，不要有其他说明或格式标记

返回格式示例：
这是一部现代都市悬疑小说，讲述了...`

// 默认场景生成 Prompt
const defaultScenePrompt = `请将以下章节内容拆分为 0-3 个关键场景，用于生成连环漫画。

要求：
1. 每个场景用一句话描述，适合作为画面生成的提示词
2. 场景要能体现章节的关键情节或重要时刻
3. 如果章节内容较少或不适合拆分场景，可以返回空数组
4. 每个场景描述要包含：地点、人物、动作/事件、氛围
5. 场景描述要具体、形象，便于理解和画图
6. 严格返回 JSON 数组格式，每个元素是一个场景描述字符串
7. 最多返回 3 个场景

章节内容：
%s

返回格式示例：
["场景1的描述文字", "场景2的描述文字", "场景3的描述文字"]`
