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
		prompt += "角色信息使用规则：\n场景描述中提到的人物需参考对应的角色信息\n"
	}

	prompt += fmt.Sprintf("根据以下场景描述生成一张动漫图片：%s\n", sceneContent)

	return prompt
}
