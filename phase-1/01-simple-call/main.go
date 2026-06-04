package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Deepseek API 兼容 OpenAI Chat Completions 格式
const baseURL = "https://api.deepseek.com/v1/chat/completions"

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置环境变量 DEEPSEEK_API_KEY")
		os.Exit(1)
	}

	req := ChatRequest{
		Model: "deepseek-v4-flash",
		Messages: []Message{
			{Role: "user", Content: "用一句话解释什么是 Go 语言的 goroutine"},
		},
	}

	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", baseURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Println("请求失败:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result ChatResponse
	json.Unmarshal(respBody, &result)

	if len(result.Choices) > 0 {
		fmt.Println(result.Choices[0].Message.Content)
		fmt.Println()
		fmt.Printf("📊 Token 用量: prompt=%d, completion=%d, total=%d\n",
			result.Usage.PromptTokens,
			result.Usage.CompletionTokens,
			result.Usage.TotalTokens)
	} else {
		fmt.Println("没有返回结果，原始响应:", string(respBody))
	}
}
