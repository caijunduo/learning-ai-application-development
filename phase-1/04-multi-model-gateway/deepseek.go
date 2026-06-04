package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const deepseekBaseURL = "https://api.deepseek.com/v1/chat/completions"

// DeepSeekProvider 实现 LLMProvider 接口，对接 DeepSeek API
type DeepSeekProvider struct {
	apiKey string
	client *http.Client
}

// NewDeepSeekProvider 创建 DeepSeek provider
func NewDeepSeekProvider() (*DeepSeekProvider, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("请设置环境变量 DEEPSEEK_API_KEY")
	}
	return &DeepSeekProvider{
		apiKey: apiKey,
		client: http.DefaultClient,
	}, nil
}

func (p *DeepSeekProvider) Name() string {
	return "DeepSeek"
}

func (p *DeepSeekProvider) Chat(model string, messages []Message) (*ChatResponse, error) {
	reqBody := struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}{
		Model:    model,
		Messages: messages,
	}

	body, _ := json.Marshal(reqBody)

	httpReq, _ := http.NewRequest("POST", deepseekBaseURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("DeepSeek 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("DeepSeek 响应解析失败: %w，原始响应: %s", err, string(respBody))
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("DeepSeek 没有返回结果，原始响应: %s", string(respBody))
	}

	return &ChatResponse{
		Content:          result.Choices[0].Message.Content,
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
		TotalTokens:      result.Usage.TotalTokens,
	}, nil
}
