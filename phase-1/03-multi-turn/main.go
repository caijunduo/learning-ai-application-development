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

	systemPrompt := `角色：Go代码审查官
约束：只需要审查Go代码即可，其他无关的事情直接回复无法处理
背景：只审查Go1.x.x版本的代码
流程：
1. 先逐行阅读代码，站在全局视角看代码，标注所有涉及错误处理或逻辑错误的位置
2. 对照Go官方错误处理规范，判断每次是否合规
3. 汇总发现，区分好的写法和问题写法
4. 按照输出格式生成报告
任务：审查Go代码的安全性、错误处理、代码规范等，重点关注错误处理是否规范
输出：按顺序输出【审查报告】【好的处理】【坏的处理】`

	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}

	fmt.Println("=== Go 代码审查工具 ===")
	fmt.Println("请输入 Go 代码（输入空行结束，输入 'quit' 退出）:")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(">>> 请输入 Go 代码:\n")

		// 读取多行代码，空行结束
		var codeLines []string
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				break
			}
			if strings.TrimSpace(line) == "quit" {
				fmt.Println("再见！")
				return
			}
			codeLines = append(codeLines, line)
		}

		code := strings.TrimSpace(strings.Join(codeLines, "\n"))
		if code == "" {
			continue
		}

		// 快速检测：如果不是 Go 代码，直接跳过
		if !looksLikeGo(code) {
			fmt.Println("⚠️  这看起来不是 Go 代码，请输入 Go 代码。")
			fmt.Println()
			continue
		}

		// 构建审查请求
		reviewMessages := append(messages, Message{
			Role:    "user",
			Content: fmt.Sprintf("请审查以下 Go 代码：\n\n```go\n%s\n```", code),
		})

		reply, promptTokens, completionTokens, totalTokens, err := callAPI(reviewMessages, apiKey)
		if err != nil {
			fmt.Println("调用失败:", err)
			continue
		}

		fmt.Println()
		fmt.Println(reply)
		fmt.Println()
		fmt.Printf("📊 Token 用量: prompt=%d, completion=%d, total=%d\n",
			promptTokens, completionTokens, totalTokens)
		fmt.Println()

		// 询问是否继续
		fmt.Print(">>> 继续审查？(直接回车继续 / 输入 quit 退出): ")
		if !scanner.Scan() {
			break
		}
		choice := strings.TrimSpace(scanner.Text())
		if choice == "quit" {
			fmt.Println("再见！")
			return
		}
		fmt.Println()
	}
}

// looksLikeGo 简单判断输入是否像 Go 代码
func looksLikeGo(code string) bool {
	code = strings.TrimSpace(code)
	// 检查是否包含 Go 关键字特征
	goKeywords := []string{"package ", "func ", "import ", "type ", "var ", "const "}
	for _, kw := range goKeywords {
		if strings.Contains(code, kw) {
			return true
		}
	}
	return false
}
