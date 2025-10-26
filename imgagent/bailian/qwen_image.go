package bailian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"imgagent/pkg/logger"
)

// GenerateCoverImage 根据摘要生成小说封面图片
// 返回图片 URL
func (c *Client) GenerateCoverImage(ctx context.Context, summary string) (string, error) {
	log := logger.FromContext(ctx)
	log.Infof("Generating cover image for summary")

	// 构建封面图 prompt
	prompt := buildCoverImagePrompt(summary)
	log.Infof("Cover image prompt: %s", prompt)

	// 构建请求
	req := ImageGenerationRequest{
		Model: "qwen-image-plus",
		Input: ImageInput{
			Messages: []ImageMessage{
				{
					Role: "user",
					Content: []ImageContent{
						{Text: prompt},
					},
				},
			},
		},
		Parameters: Parameters{
			NegativePrompt: "",
			PromptExtend:   true,
			Watermark:      c.config.ImageWatermark,
			Size:           c.config.ImageSize,
		},
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Failed to marshal request, err: %v", err)
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/api/v1/services/aigc/multimodal-generation/generation", c.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		log.Errorf("Failed to create request, err: %v", err)
		return "", fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Errorf("Failed to send request, err: %v", err)
		return "", fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response, err: %v", err)
		return "", fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Generate cover image failed, status: %d, body: %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("generate cover image failed, status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var imgResp ImageGenerationResponse
	err = json.Unmarshal(respBody, &imgResp)
	if err != nil {
		log.Errorf("Failed to parse response, err: %v, body: %s", err, string(respBody))
		return "", fmt.Errorf("parse response failed: %w", err)
	}

	if len(imgResp.Output.Choices) == 0 {
		log.Errorf("No choices in response, body: %s", string(respBody))
		return "", fmt.Errorf("no choices in response")
	}

	choice := imgResp.Output.Choices[0]
	if len(choice.Message.Content) == 0 {
		log.Errorf("No content in choice, response: %s", string(respBody))
		return "", fmt.Errorf("no content in choice")
	}

	imageURL := choice.Message.Content[0].Image
	log.Infof("Cover image generated successfully, URL: %s", imageURL)
	return imageURL, nil
}

// buildCoverImagePrompt 构建封面图生成 prompt
func buildCoverImagePrompt(summary string) string {
	prompt := `请为这本小说设计一张精美的封面图片。

小说摘要：
` + summary + `

要求：
1. 风格：符合小说整体风格和时代背景
2. 色调：根据小说氛围选择合适的色调（如历史题材用古典色调，悬疑题材用暗色调等）
3. 元素：包含能代表小说主题的关键元素（人物、场景、象征物等）
4. 构图：专业书籍封面构图，突出标题区域，适合作为封面展示
5. 画质：高清、精美、具有视觉冲击力
6. 风格统一：整体风格和谐统一，符合小说类型

请生成一张能够吸引读者的精美封面图。`
	return prompt
}

// GenerateImage 根据场景描述生成图片
// 返回图片 URL
func (c *Client) GenerateImage(ctx context.Context, sceneContent string, summary string, roles []RoleInfo) (string, error) {
	log := logger.FromContext(ctx)
	log.Infof("Generating image for scene, content: %s", sceneContent)

	// 构建完整的提示词
	prompt := buildImagePrompt(sceneContent, summary, roles)
	log.Infof("Full image prompt: %s", prompt)

	// 构建请求
	req := ImageGenerationRequest{
		Model: "qwen-image-plus",
		Input: ImageInput{
			Messages: []ImageMessage{
				{
					Role: "user",
					Content: []ImageContent{
						{Text: prompt},
					},
				},
			},
		},
		Parameters: Parameters{
			NegativePrompt: "",
			PromptExtend:   true,
			Watermark:      c.config.ImageWatermark,
			Size:           c.config.ImageSize,
		},
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Failed to marshal request, err: %v", err)
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/api/v1/services/aigc/multimodal-generation/generation", c.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		log.Errorf("Failed to create request, err: %v", err)
		return "", fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Errorf("Failed to send request, err: %v", err)
		return "", fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response, err: %v", err)
		return "", fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Generate image failed, status: %d, body: %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("generate image failed, status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var imgResp ImageGenerationResponse
	err = json.Unmarshal(respBody, &imgResp)
	if err != nil {
		log.Errorf("Failed to parse response, err: %v, body: %s", err, string(respBody))
		return "", fmt.Errorf("parse response failed: %w", err)
	}

	if len(imgResp.Output.Choices) == 0 {
		log.Errorf("No choices in response, body: %s", string(respBody))
		return "", fmt.Errorf("no choices in response")
	}

	choice := imgResp.Output.Choices[0]
	if len(choice.Message.Content) == 0 {
		log.Errorf("No content in choice, response: %s", string(respBody))
		return "", fmt.Errorf("no content in choice")
	}

	imageURL := choice.Message.Content[0].Image
	if imageURL == "" {
		log.Errorf("Image URL is empty, response: %s", string(respBody))
		return "", fmt.Errorf("image URL is empty")
	}

	log.Infof("Image generated successfully, URL: %s", imageURL)
	return imageURL, nil
}

func buildImagePrompt(sceneContent string, summary string, roles []RoleInfo) string {
	var prompt string

	if summary != "" {
		prompt += fmt.Sprintf("小说概要：%s\n\n", summary)
	}

	if len(roles) > 0 {
		prompt += "主要角色信息：\n"
		for _, role := range roles {
			if role.Appearance != "" {
				prompt += fmt.Sprintf("- %s：性别：%s； 性格特点：%s；外貌特征：%s\n", role.Name, role.Gender, role.Character, role.Appearance)
			}
		}
		prompt += "角色信息使用规则：场景描述中提到的人物需参考对应的角色信息。\n\n"
	}

	prompt += fmt.Sprintf("根据以下场景描述生成一张动漫图片：%s\n", sceneContent)

	return prompt
}
