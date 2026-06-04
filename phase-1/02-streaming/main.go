package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const baseURL = "https://api.deepseek.com/v1/chat/completions"

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置环境变量 DEEPSEEK_API_KEY")
		os.Exit(1)
	}

	req := ChatRequest{
		Model:  "deepseek-v4-flash",
		Stream: true, // 开启流式输出
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

	// 逐行读取 SSE 事件流
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("读取流失败:", err)
			break
		}

		line = strings.TrimSpace(line)

		// SSE data 行以 "data: " 开头
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// 流结束标记
		if data == "[DONE]" {
			fmt.Println() // 换行
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			// 实时打印每个 chunk 的内容
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}
}
