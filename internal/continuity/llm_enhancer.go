package continuity

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"agentscope-desktop/internal/llm"
	"agentscope-desktop/internal/session"
)

// LLMEnhancer 使用 LLM 增强多会话摘要质量。
type LLMEnhancer struct {
	client *llm.Client
}

// NewLLMEnhancer 创建新的 LLM 增强器。
func NewLLMEnhancer(cfg llm.ProviderConfig) *LLMEnhancer {
	return &LLMEnhancer{
		client: llm.NewClient(cfg),
	}
}

// EnhanceSummary 使用 LLM 分析会话数据，返回增强后的摘要结果。
// 如果 LLM 调用失败，返回错误，调用方应回退到关键词结果。
func (e *LLMEnhancer) EnhanceSummary(
	ctx context.Context,
	sessions []*session.Session,
	keywordResult *HandoffSummary,
) (*LLMEnhancedResult, error) {
	messages := e.buildMessages(sessions, keywordResult)

	resp, err := e.client.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM调用失败: %w", err)
	}

	result, err := e.parseResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("LLM响应解析失败: %w", err)
	}

	return result, nil
}

// buildMessages 构造发送给 LLM 的消息列表。
func (e *LLMEnhancer) buildMessages(
	sessions []*session.Session,
	keywordResult *HandoffSummary,
) []llm.ChatMessage {
	return []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: e.buildUserMessage(sessions, keywordResult)},
	}
}

// buildUserMessage 将从会话中提取的结构化数据组装为 LLM 可理解的文本。
func (e *LLMEnhancer) buildUserMessage(
	sessions []*session.Session,
	keywordResult *HandoffSummary,
) string {
	var b strings.Builder

	b.WriteString("请分析以下编程会话的数据，生成一份项目交接摘要。\n\n")
	b.WriteString(fmt.Sprintf("共 %d 个会话。\n\n", len(sessions)))

	// 已有的关键词提取结果（作为参考）
	b.WriteString("=== 关键词规则引擎已经提取的候选数据（供参考） ===\n\n")
	b.WriteString(e.formatKeywordResult(keywordResult))
	b.WriteString("\n")

	// 结构化的会话数据
	b.WriteString("=== 原始会话数据 ===\n\n")
	maxSessions := 10
	if len(sessions) < maxSessions {
		maxSessions = len(sessions)
	}
	for i := 0; i < maxSessions; i++ {
		s := sessions[i]
		b.WriteString(e.formatSessionData(s, i+1))
		b.WriteString("\n")
	}

	if len(sessions) > maxSessions {
		b.WriteString(fmt.Sprintf("... 还有 %d 个早期会话未展示。\n", len(sessions)-maxSessions))
	}

	return b.String()
}

// formatKeywordResult 格式化关键词提取的候选数据。
func (e *LLMEnhancer) formatKeywordResult(summary *HandoffSummary) string {
	var b strings.Builder

	if len(summary.CompletedTasks) > 0 {
		b.WriteString("已完成任务候选:\n")
		for _, t := range summary.CompletedTasks {
			b.WriteString(fmt.Sprintf("  - %s [文件: %s]\n",
				t.Description, strings.Join(t.FilesChanged, ", ")))
		}
	}

	if len(summary.PendingTasks) > 0 {
		b.WriteString("待办候选:\n")
		for _, t := range summary.PendingTasks {
			b.WriteString(fmt.Sprintf("  - %s [来源: %s]\n", t.Description, t.Source))
		}
	}

	if len(summary.KeyDecisions) > 0 {
		b.WriteString("决策候选:\n")
		for _, d := range summary.KeyDecisions {
			b.WriteString(fmt.Sprintf("  - %s\n", d.Description))
		}
	}

	if len(summary.KnownIssues) > 0 {
		b.WriteString("已知问题候选:\n")
		for _, issue := range summary.KnownIssues {
			b.WriteString(fmt.Sprintf("  - %s\n", issue))
		}
	}

	return b.String()
}

// formatSessionData 格式化为 LLM 友好的会话摘要。
func (e *LLMEnhancer) formatSessionData(sess *session.Session, index int) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("### 会话 %d\n", index))
	b.WriteString(fmt.Sprintf("- 模型: %s\n", sess.Model))
	b.WriteString(fmt.Sprintf("- 时间: %s\n", sess.StartedAt.Format(time.RFC3339)))
	if sess.Prompt != "" {
		desc := truncateUTF8(sess.Prompt, 500)
		b.WriteString(fmt.Sprintf("- 用户需求: %s\n", desc))
	}

	// 文件操作摘要
	if len(sess.Actions) > 0 {
		files := make(map[string]int)
		actionTypes := make(map[string]int)
		for _, a := range sess.Actions {
			if a.FilePath != "" {
				files[a.FilePath]++
			}
			actionTypes[string(a.Type)]++
		}
		if len(files) > 0 {
			// 排序取 top 10
			type fi struct {
				p string
				c int
			}
			var fis []fi
			for p, c := range files {
				fis = append(fis, fi{p, c})
			}
			sort.Slice(fis, func(i, j int) bool { return fis[i].c > fis[j].c })
			b.WriteString("- 操作的文件:\n")
			limit := 10
			if len(fis) < limit {
				limit = len(fis)
			}
			for i := 0; i < limit; i++ {
				b.WriteString(fmt.Sprintf("    - %s (%d次)\n", fis[i].p, fis[i].c))
			}
		}
		// 操作类型分布
		b.WriteString(fmt.Sprintf("- 操作统计: %v\n", actionTypes))
	}

	// Token 用量
	b.WriteString(fmt.Sprintf("- Token: 输入%d 输出%d\n",
		sess.TokenUsage.InputTokens, sess.TokenUsage.OutputTokens))

	// 压缩后的用户消息
	var userMsgs []string
	for _, msg := range sess.Messages {
		if msg.Type == session.MessageTypeUser {
			for _, block := range msg.Content {
				if block.Type == session.ContentTypeText && block.Text != "" {
					text := truncateUTF8(strings.TrimSpace(block.Text), 300)
					if len(text) > 5 {
						userMsgs = append(userMsgs, text)
					}
				}
			}
		}
	}
	if len(userMsgs) > 0 {
		b.WriteString("- 用户消息摘要:\n")
		maxMsgs := 5
		if len(userMsgs) < maxMsgs {
			maxMsgs = len(userMsgs)
		}
		for i := 0; i < maxMsgs; i++ {
			b.WriteString(fmt.Sprintf("    [%d] %s\n", i+1, userMsgs[i]))
		}
	}

	return b.String()
}

// parseResponse 解析 LLM 的 JSON 响应。
func (e *LLMEnhancer) parseResponse(content string) (*LLMEnhancedResult, error) {
	// 尝试提取 JSON 部分（LLM 可能在 JSON 前后添加说明文字）
	content = extractJSONBlock(content)

	var result LLMEnhancedResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w\n原始内容: %s", err, truncateUTF8(content, 500))
	}

	return &result, nil
}

// extractJSONBlock 从文本中提取 JSON 块。
func extractJSONBlock(text string) string {
	// 优先提取 ```json ... ``` 中的内容
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}

// systemPrompt 是发给 LLM 的系统指令。
const systemPrompt = `你是一个专业的软件开发项目分析助手。你的任务是分析多个编程会话记录，生成一份高质量的项目交接摘要。

请严格按以下JSON格式返回（不要包含markdown代码块标记）：
{
  "narrativeSummary": "用200字左右连贯叙述这些会话完成的核心工作",
  "completedTasks": [
    {"description": "已完成任务的简洁描述", "filesHint": ["涉及的关键文件"], "confidence": 0.9}
  ],
  "pendingTasks": [
    {"description": "明确的待办事项或隐式未完成的工作", "priority": "high|medium|low", "source": "explicit|implicit", "confidence": 0.8}
  ],
  "keyDecisions": [
    {"description": "关键架构/技术决策", "rationale": "决策背后考虑的理由", "confidence": 0.85}
  ],
  "knownIssues": ["需要注意或避免的问题"]
}

注意事项：
1. narrativeSummary 应高度凝练，聚焦核心成果和当前状态，而非流水账
2. 区分已完成和待办：已完成 = 代码已提交/功能已实现；待办 = 用户明确提过或从上下文推断仍需继续的工作
3. source: "explicit" 表示用户明确说要做，"implicit" 表示从上下文推断仍需完成
4. confidence: 0.0-1.0，表示你对这条判断的确信程度
5. 去重：相同内容不要重复出现在多个维度
6. 文件路径尽量使用原始格式
7. 请理解中英文混合的会话内容
8. 严格控制输出量：completedTasks 最多 8 条，pendingTasks 最多 5 条，keyDecisions 最多 5 条，knownIssues 最多 5 条
9. narrativeSummary 控制在 150 字以内，聚焦最核心的成果和状态`
