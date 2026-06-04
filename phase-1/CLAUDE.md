# Phase 1 · 模块一：LLM API 调用

## 概念地图

**这个东西是什么**

LLM API 就是一个 HTTP 接口。你发一段文字进去，它返回一段文字出来。本质上和你平时调用任何第三方 API 没有区别。

但它有几个地方和普通 API 不一样，这些差异决定了你后续怎么设计系统。

---

**三个必须理解的核心概念**

**① Messages 结构**

不是简单的 `input → output`，而是一个对话数组：

```json
[
  {"role": "system", "content": "你是一个助手"},
  {"role": "user", "content": "你好"},
  {"role": "assistant", "content": "你好，有什么可以帮你？"},
  {"role": "user", "content": "帮我写段代码"}
]
```

关键问题：**模型没有记忆**。每次调用都要把完整的对话历史带上去，模型才知道「上下文」是什么。这一点会深刻影响你后面怎么设计多轮对话系统。

---

**② Streaming 流式输出**

模型不是算完再返回，而是一个 token 一个 token 地吐出来。

为什么重要：
- 用户体验（不用等10秒才看到结果）
- 你的服务必须支持 SSE 或 chunked response
- 处理流式数据的方式和普通 JSON 响应完全不同

这是工程上第一个容易卡住的地方。

---

**③ Token 和计费**

模型处理的不是字符，是 token。

```
"Hello world" ≈ 2 tokens
"你好世界"    ≈ 4-8 tokens（中文更贵）
```

为什么重要：
- `max_tokens` 控制输出长度
- 输入 + 输出都计费
- Context Window 有上限，超了就报错

你在做多模型网关的时候，token 的统计和透传是一个真实的工程问题。

---

**一个参数你要现在就搞清楚：Temperature**

```
Temperature = 0    → 确定性输出，每次结果一样
Temperature = 1    → 正常创意
Temperature = 2    → 很随机，通常用不到
```

做应用的时候：代码生成用低 temperature，创意写作用高 temperature。这不是玄学，是概率分布的问题，后面我会追问你。

---

## 你现在要做的事

**第一步：读文档（30分钟）**

只读这两个，不要发散：
- [Anthropic Messages API](https://docs.anthropic.com/en/api/messages)
- [OpenAI Chat Completions](https://platform.openai.com/docs/api-reference/chat)

重点看：请求结构、响应结构、Streaming 怎么用。

---

**第二步：用 Go 写三个东西（用 Claude Code 做）**

```
1. 最简单的单次调用
   发一个 user message，打印 assistant 的回复

2. 流式输出
   同样的调用，改成 streaming，实时打印每个 chunk

3. 多轮对话
   维护一个 messages 数组，支持连续对话 3 轮
```

不需要封装，不需要优雅，先跑通、先感受。

---

**第三步：完成后来找我**

我会问你几个问题，不是考你 API 文档，是考你「真的理解了什么」。

比如：
- 如果用户发了一条消息，模型返回到一半断了，你怎么处理？
- 为什么 system message 要单独一个 role，直接放在 user message 里不行吗？
- token 计费是在哪一端算的？
