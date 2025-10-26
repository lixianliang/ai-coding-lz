package bailian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"imgagent/pkg/logger"
)

// ExtractSummary 提取整个小说的摘要
func (c *Client) ExtractSummary(ctx context.Context, fileID string) (string, error) {
	log := logger.FromContext(ctx)
	log.Infof("Extracting summary from document, fileID: %s", fileID)

	req := ChatCompletionRequest{
		Model: "qwen-long",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "system", Content: fmt.Sprintf("fileid://%s", fileID)},
			{Role: "user", Content: c.config.SummaryPrompt},
		},
		Stream: false,
	}

	respBody, err := c.callChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	var chatResp ChatCompletionResponse
	err = json.Unmarshal(respBody, &chatResp)
	if err != nil {
		log.Errorf("Failed to parse chat response, err: %v, body: %s", err, string(respBody))
		return "", fmt.Errorf("parse chat response failed: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		log.Warnf("No choices in response, body: %s", string(respBody))
		return "", fmt.Errorf("no choices in response")
	}

	summary := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	log.Infof("Extracted summary (length: %d): %s", len(summary), summary)

	return summary, nil
}

// ExtractRoles 从文档中提取角色信息
// 使用 qwen-long 分析整个文档
func (c *Client) ExtractRoles(ctx context.Context, fileID string, summary string) ([]RoleInfo, error) {
	log := logger.FromContext(ctx)
	log.Infof("Extracting roles from document, fileID: %s", fileID)

	// 构建请求
	prompt := c.config.RolePrompt
	if summary != "" {
		prompt = fmt.Sprintf("小说摘要：\n%s\n\n%s", summary, c.config.RolePrompt)
	}

	req := ChatCompletionRequest{
		Model: "qwen-long",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "system", Content: fmt.Sprintf("fileid://%s", fileID)},
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	// 调用 API
	respBody, err := c.callChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var chatResp ChatCompletionResponse
	err = json.Unmarshal(respBody, &chatResp)
	if err != nil {
		log.Errorf("Failed to parse chat response, err: %v, body: %s", err, string(respBody))
		return nil, fmt.Errorf("parse chat response failed: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		log.Warnf("No choices in response, body: %s", string(respBody))
		return []RoleInfo{}, nil
	}

	content := chatResp.Choices[0].Message.Content
	log.Infof("Raw role extraction response: %s", content)

	// 提取 JSON 内容
	roles, err := extractRolesFromJSON(content)
	if err != nil {
		log.Errorf("Failed to extract roles from JSON, err: %v, content: %s", err, content)
		return nil, fmt.Errorf("extract roles from JSON failed: %w", err)
	}

	log.Infof("Extracted %d roles", len(roles))
	return roles, nil
}

// GenerateScenes 为章节生成场景描述
// 每章生成 0-3 个场景
func (c *Client) GenerateScenes(ctx context.Context, chapterContent string) ([]string, error) {
	log := logger.FromContext(ctx)
	log.Infof("Generating scenes for chapter, content length: %d", len(chapterContent))

	// 构建 prompt
	prompt := fmt.Sprintf(c.config.ScenePrompt, chapterContent)

	// 构建请求
	req := ChatCompletionRequest{
		Model: "qwen-long",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	// 调用 API
	respBody, err := c.callChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var chatResp ChatCompletionResponse
	err = json.Unmarshal(respBody, &chatResp)
	if err != nil {
		log.Errorf("Failed to parse chat response, err: %v, body: %s", err, string(respBody))
		return nil, fmt.Errorf("parse chat response failed: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		log.Warnf("No choices in response, body: %s", string(respBody))
		return []string{}, nil
	}

	content := chatResp.Choices[0].Message.Content
	log.Infof("Raw scene generation response: %s", content)

	// 提取场景描述
	scenes, err := extractScenesFromJSON(content)
	if err != nil {
		log.Errorf("Failed to extract scenes from JSON, err: %v, content: %s", err, content)
		return nil, fmt.Errorf("extract scenes from JSON failed: %w", err)
	}

	// 限制最多 3 个场景
	if len(scenes) > 3 {
		log.Warnf("Got %d scenes, truncating to 3", len(scenes))
		scenes = scenes[:3]
	}

	log.Infof("Generated %d scenes", len(scenes))
	return scenes, nil
}

// callChatCompletion 调用 chat completion API
func (c *Client) callChatCompletion(ctx context.Context, req ChatCompletionRequest) ([]byte, error) {
	log := logger.FromContext(ctx)

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Failed to marshal request, err: %v", err)
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/compatible-mode/v1/chat/completions", c.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		log.Errorf("Failed to create request, err: %v", err)
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Errorf("Failed to send request, err: %v", err)
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response, err: %v", err)
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("API call failed, status: %d, body: %s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("API call failed, status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// extractRolesFromJSON 从 JSON 字符串中提取角色信息
func extractRolesFromJSON(content string) ([]RoleInfo, error) {
	// 尝试直接解析
	var roles []RoleInfo
	err := json.Unmarshal([]byte(content), &roles)
	if err == nil {
		return roles, nil
	}

	// 尝试提取 JSON 数组（可能包含在代码块或其他文字中）
	jsonPattern := regexp.MustCompile(`\[[\s\S]*?\]`)
	matches := jsonPattern.FindAllString(content, -1)

	for _, match := range matches {
		err = json.Unmarshal([]byte(match), &roles)
		if err == nil && len(roles) > 0 {
			return roles, nil
		}
	}

	// 如果都失败，返回空数组（不报错，因为可能真的没有角色）
	return []RoleInfo{}, nil
}

// extractScenesFromJSON 从 JSON 字符串中提取场景描述
func extractScenesFromJSON(content string) ([]string, error) {
	// 尝试直接解析
	var scenes []string
	err := json.Unmarshal([]byte(content), &scenes)
	if err == nil {
		return scenes, nil
	}

	// 尝试提取 JSON 数组
	jsonPattern := regexp.MustCompile(`\[[\s\S]*?\]`)
	matches := jsonPattern.FindAllString(content, -1)

	for _, match := range matches {
		err = json.Unmarshal([]byte(match), &scenes)
		if err == nil {
			// 过滤空字符串
			filtered := make([]string, 0)
			for _, scene := range scenes {
				scene = strings.TrimSpace(scene)
				if scene != "" {
					filtered = append(filtered, scene)
				}
			}
			return filtered, nil
		}
	}

	// 如果都失败，返回空数组（不报错，因为可能内容不适合生成场景）
	return []string{}, nil
}
