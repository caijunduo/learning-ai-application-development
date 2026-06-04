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

func callAPI(messages []Message, apiKey string) (string, int, int, int, error) {
	req := ChatRequest{
		Model:    "deepseek-v4-flash",
		Messages: messages,
	}

	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", baseURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", 0, 0, 0, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result ChatResponse
	json.Unmarshal(respBody, &result)

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content,
			result.Usage.PromptTokens,
			result.Usage.CompletionTokens,
			result.Usage.TotalTokens,
			nil
	}
	return "", 0, 0, 0, fmt.Errorf("没有返回结果: %s", string(respBody))
}

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置环境变量 DEEPSEEK_API_KEY")
		os.Exit(1)
	}

	// 维护一个 messages 数组，这就是模型的"记忆"
	messages := []Message{
		{Role: "system", Content: "你是一个简洁的助手，回答尽量简短。"},
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== 多轮对话（共 3 轮）===")
	fmt.Println()

	for turn := 1; turn <= 3; turn++ {
		fmt.Printf("--- 第 %d 轮 ---\n", turn)
		fmt.Print("你: ")

		if !scanner.Scan() {
			break
		}
		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			turn-- // 空输入不计轮次
			continue
		}

		// 把用户消息加入历史
		messages = append(messages, Message{Role: "user", Content: userInput})

		// 把完整历史发给模型
		reply, promptTokens, completionTokens, totalTokens, err := callAPI(messages, apiKey)
		if err != nil {
			fmt.Println("调用失败:", err)
			continue
		}

		// 把模型回复也加入历史
		messages = append(messages, Message{Role: "assistant", Content: reply})

		fmt.Printf("AI: %s\n", reply)
		fmt.Printf("📊 Token 用量: prompt=%d, completion=%d, total=%d\n\n",
			promptTokens, completionTokens, totalTokens)
	}

	fmt.Println("=== 对话结束 ===")
	fmt.Printf("\n最终 messages 数组长度: %d 条\n", len(messages))
}
