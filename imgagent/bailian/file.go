package bailian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"imgagent/pkg/logger"
)

// UploadFile 上传文件到阿里云百炼
// 返回 fileID 用于后续 qwen-long 调用
func (c *Client) UploadFile(ctx context.Context, filename string) (string, error) {
	log := logger.FromContext(ctx)
	log.Infof("Uploading file to Bailian, filename: %s", filename)

	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Failed to open file, err: %v", err)
		return "", fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 创建 multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件字段
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		log.Errorf("Failed to create form file, err: %v", err)
		return "", fmt.Errorf("create form file failed: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		log.Errorf("Failed to copy file content, err: %v", err)
		return "", fmt.Errorf("copy file content failed: %w", err)
	}

	// 添加 purpose 字段
	err = writer.WriteField("purpose", "file-extract")
	if err != nil {
		log.Errorf("Failed to write field, err: %v", err)
		return "", fmt.Errorf("write field failed: %w", err)
	}

	err = writer.Close()
	if err != nil {
		log.Errorf("Failed to close writer, err: %v", err)
		return "", fmt.Errorf("close writer failed: %w", err)
	}

	// 创建请求
	url := fmt.Sprintf("%s/compatible-mode/v1/files", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		log.Errorf("Failed to create request, err: %v", err)
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := c.httpClient.Do(req)
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
		log.Errorf("Upload file failed, status: %d, body: %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("upload file failed, status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var uploadResp UploadFileResponse
	err = json.Unmarshal(respBody, &uploadResp)
	if err != nil {
		log.Errorf("Failed to parse response, err: %v, body: %s", err, string(respBody))
		return "", fmt.Errorf("parse response failed: %w", err)
	}

	if uploadResp.ID == "" {
		log.Errorf("File ID is empty, response: %s", string(respBody))
		return "", fmt.Errorf("file ID is empty")
	}

	log.Infof("File uploaded successfully, fileID: %s", uploadResp.ID)
	return uploadResp.ID, nil
}
