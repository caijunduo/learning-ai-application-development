package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const zhipuBaseURL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"

// ZhipuProvider 实现 LLMProvider 接口，对接智谱 GLM API
type ZhipuProvider struct {
	apiKey string
	client *http.Client
}

// NewZhipuProvider 创建智谱 provider
func NewZhipuProvider() (*ZhipuProvider, error) {
	apiKey := os.Getenv("ZHIPU_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("请设置环境变量 ZHIPU_API_KEY")
	}
	return &ZhipuProvider{
		apiKey: apiKey,
		client: http.DefaultClient,
	}, nil
}

func (p *ZhipuProvider) Name() string {
	return "智谱 (Zhipu)"
}

func (p *ZhipuProvider) Chat(model string, messages []Message) (*ChatResponse, error) {
	reqBody := struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}{
		Model:    model,
		Messages: messages,
	}

	body, _ := json.Marshal(reqBody)

	httpReq, _ := http.NewRequest("POST", zhipuBaseURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("智谱 请求失败: %w", err)
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
		return nil, fmt.Errorf("智谱 响应解析失败: %w，原始响应: %s", err, string(respBody))
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("智谱 没有返回结果，原始响应: %s", string(respBody))
	}

	return &ChatResponse{
		Content:          result.Choices[0].Message.Content,
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
		TotalTokens:      result.Usage.TotalTokens,
	}, nil
}
