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

func (c *Client) GenerateTTS(ctx context.Context, text string) (string, error) {
	log := logger.FromContext(ctx)
	log.Infof("Generating TTS for text, length: %d", len(text))

	req := TTSRequest{
		Model: "qwen3-tts-flash",
		Input: TTSInput{
			Text:         text,
			Voice:        "Cherry",
			LanguageType: "Chinese",
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Failed to marshal request, err: %v", err)
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/services/aigc/multimodal-generation/generation", c.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		log.Errorf("Failed to create request, err: %v", err)
		return "", fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Errorf("Failed to send request, err: %v", err)
		return "", fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response, err: %v", err)
		return "", fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Generate TTS failed, status: %d, body: %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("generate TTS failed, status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var ttsResp TTSResponse
	err = json.Unmarshal(respBody, &ttsResp)
	if err != nil {
		log.Errorf("Failed to parse response, err: %v, body: %s", err, string(respBody))
		return "", fmt.Errorf("parse response failed: %w", err)
	}

	if ttsResp.Output.Audio.URL == "" {
		log.Errorf("Audio URL is empty, response: %s", string(respBody))
		return "", fmt.Errorf("audio URL is empty")
	}

	log.Infof("TTS generated successfully, URL: %s", ttsResp.Output.Audio.URL)
	return ttsResp.Output.Audio.URL, nil
}
