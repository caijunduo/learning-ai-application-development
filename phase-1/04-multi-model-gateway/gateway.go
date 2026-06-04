package main

import (
	"fmt"
	"strings"
)

// Message 单条消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 统一返回结构
type ChatResponse struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// LLMProvider 模型提供者接口 —— 这是网关的核心抽象
type LLMProvider interface {
	// Name 返回 provider 名称（用于注册和日志）
	Name() string
	// Chat 发送消息并返回回复
	Chat(model string, messages []Message) (*ChatResponse, error)
}

// providers 注册表：前缀 → Provider 实例
var providers = map[string]LLMProvider{}

// Register 注册一个 Provider
func Register(prefix string, p LLMProvider) {
	providers[prefix] = p
}

// Chat 统一入口：根据 model 前缀路由到对应 Provider
// model 格式: "prefix/model-name"，例如 "deepseek/deepseek-v4-flash"、"zai/glm-4-flash"
func Chat(model string, messages []Message) (*ChatResponse, error) {
	parts := strings.SplitN(model, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的 model 格式: %q，应为 \"前缀/模型名\"，例如 \"deepseek/deepseek-v4-flash\"", model)
	}

	prefix := parts[0]
	actualModel := parts[1]

	p, ok := providers[prefix]
	if !ok {
		keys := make([]string, 0, len(providers))
		for k := range providers {
			keys = append(keys, k)
		}
		return nil, fmt.Errorf("未知的 provider 前缀: %q，已注册的前缀: %v", prefix, keys)
	}

	return p.Chat(actualModel, messages)
}
