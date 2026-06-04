package main

import (
	"fmt"
	"os"
)

func main() {
	// 1. 注册所有 Provider（在 init 里也可以，但显式调用更清晰）
	deepseek, err := NewDeepSeekProvider()
	if err != nil {
		fmt.Println("DeepSeek provider 初始化失败:", err)
	} else {
		Register("deepseek", deepseek)
		fmt.Println("✅ DeepSeek provider 已注册")
	}

	zhipu, err := NewZhipuProvider()
	if err != nil {
		fmt.Println("智谱 provider 初始化失败:", err)
	} else {
		Register("zai", zhipu)
		fmt.Println("✅ 智谱 (Zhipu) provider 已注册")
	}

	if len(providers) == 0 {
		fmt.Println("\n⚠️  没有可用的 provider，请设置 API Key 后重试")
		fmt.Println("   export DEEPSEEK_API_KEY=your-key")
		fmt.Println("   export ZHIPU_API_KEY=your-key")
		os.Exit(1)
	}

	fmt.Println()

	// 2. 测试：用统一入口调用不同模型
	messages := []Message{
		{Role: "user", Content: "用一句话解释什么是 Go 语言的接口（interface）"},
	}

	// 测试 DeepSeek
	if _, ok := providers["deepseek"]; ok {
		fmt.Println("=== 调用 DeepSeek ===")
		resp, err := Chat("deepseek/deepseek-v4-flash", messages)
		if err != nil {
			fmt.Println("❌ 调用失败:", err)
		} else {
			fmt.Println(resp.Content)
			fmt.Printf("\n📊 Token 用量: prompt=%d, completion=%d, total=%d\n\n",
				resp.PromptTokens, resp.CompletionTokens, resp.TotalTokens)
		}
	}

	// 测试智谱
	if _, ok := providers["zai"]; ok {
		fmt.Println("=== 调用智谱 (Zhipu) ===")
		resp, err := Chat("zai/glm-4-flash", messages)
		if err != nil {
			fmt.Println("❌ 调用失败:", err)
		} else {
			fmt.Println(resp.Content)
			fmt.Printf("\n📊 Token 用量: prompt=%d, completion=%d, total=%d\n\n",
				resp.PromptTokens, resp.CompletionTokens, resp.TotalTokens)
		}
	}

	// 3. 测试错误路由
	fmt.Println("=== 测试未知前缀 ===")
	_, err = Chat("unknown/model", messages)
	fmt.Println("预期错误:", err)
}
