package bailian

// RoleInfo 角色信息
type RoleInfo struct {
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	Character  string `json:"character"`
	Appearance string `json:"appearance"`
}

// UploadFileResponse 文件上传响应
type UploadFileResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
}

// ChatCompletionRequest qwen-long 请求
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse qwen-long 响应
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ImageGenerationRequest 图片生成请求
type ImageGenerationRequest struct {
	Model      string     `json:"model"`
	Input      ImageInput `json:"input"`
	Parameters Parameters `json:"parameters"`
}

// ImageInput 图片输入
type ImageInput struct {
	Messages []ImageMessage `json:"messages"`
}

// ImageMessage 图片消息
type ImageMessage struct {
	Role    string         `json:"role"`
	Content []ImageContent `json:"content"`
}

// ImageContent 图片内容
type ImageContent struct {
	Text string `json:"text"`
}

// Parameters 参数
type Parameters struct {
	NegativePrompt string `json:"negative_prompt"`
	PromptExtend   bool   `json:"prompt_extend"`
	Watermark      bool   `json:"watermark"`
	Size           string `json:"size"`
}

// ImageGenerationResponse 图片生成响应
type ImageGenerationResponse struct {
	Output ImageOutput `json:"output"`
	Usage  ImageUsage  `json:"usage"`
}

// ImageOutput 输出
type ImageOutput struct {
	Choices []ImageChoice `json:"choices"`
}

// ImageChoice 选择
type ImageChoice struct {
	FinishReason string           `json:"finish_reason"`
	Message      ImageResponseMsg `json:"message"`
}

// ImageResponseMsg 消息
type ImageResponseMsg struct {
	Role    string              `json:"role"`
	Content []ImageResponseItem `json:"content"`
}

// ImageResponseItem 内容项
type ImageResponseItem struct {
	Image string `json:"image"`
}

// ImageUsage 使用情况
type ImageUsage struct {
	Width       int `json:"width"`
	Height      int `json:"height"`
	ImageCount  int `json:"image_count"`
}
