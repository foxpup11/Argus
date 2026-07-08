package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"argus-desktop/internal/llm"
	"argus-desktop/internal/session"
)

// LLMAuditor 使用 LLM 进行合规审计。
type LLMAuditor struct {
	client *llm.Client
}

// NewLLMAuditor 创建新的 LLM 审计器。
func NewLLMAuditor(cfg llm.ProviderConfig) *LLMAuditor {
	return &LLMAuditor{
		client: llm.NewClient(cfg),
	}
}

// ExtractRules 使用 LLM 从 CLAUDE.md 内容中提取结构化规则。
func (a *LLMAuditor) ExtractRules(ctx context.Context, claudeMDContent string) ([]ComplianceRule, error) {
	messages := []llm.ChatMessage{
		{Role: "system", Content: extractRulesSystemPrompt},
		{Role: "user", Content: fmt.Sprintf("请从以下 CLAUDE.md 内容中提取所有行为规则：\n\n%s", claudeMDContent)},
	}

	resp, err := a.client.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM规则提取失败: %w", err)
	}

	rules, err := parseRulesResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("规则解析失败: %w", err)
	}

	return rules, nil
}

// AuditSession 使用 LLM 审计单个会话的规则遵守情况。
func (a *LLMAuditor) AuditSession(ctx context.Context, rules []ComplianceRule, sess *session.Session) (*ComplianceScore, error) {
	messages := []llm.ChatMessage{
		{Role: "system", Content: auditSessionSystemPrompt},
		{Role: "user", Content: a.buildAuditUserMessage(rules, sess)},
	}

	resp, err := a.client.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM审计失败: %w", err)
	}

	score, err := parseAuditResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("审计结果解析失败: %w", err)
	}

	return score, nil
}

// buildAuditUserMessage 构建审计请求的用户消息。
func (a *LLMAuditor) buildAuditUserMessage(rules []ComplianceRule, sess *session.Session) string {
	var b strings.Builder

	// 规则列表
	b.WriteString("=== CLAUDE.md 规则 ===\n\n")
	for i, r := range rules {
		b.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, r.Category, r.Rule))
		if r.Description != "" {
			b.WriteString(fmt.Sprintf("   描述: %s\n", r.Description))
		}
	}

	b.WriteString("\n=== 会话数据 ===\n\n")
	b.WriteString(fmt.Sprintf("- 会话ID: %s\n", sess.ID))
	b.WriteString(fmt.Sprintf("- 模型: %s\n", sess.Model))
	b.WriteString(fmt.Sprintf("- 工作目录: %s\n", sess.CWD))

	// 用户请求
	if sess.Prompt != "" {
		prompt := truncateUTF8(sess.Prompt, 500)
		b.WriteString(fmt.Sprintf("- 用户请求: %s\n", prompt))
	}

	// 文件变更
	if len(sess.FileChanges) > 0 {
		b.WriteString(fmt.Sprintf("\n文件变更 (%d 个文件):\n", len(sess.FileChanges)))
		limit := 20
		if len(sess.FileChanges) < limit {
			limit = len(sess.FileChanges)
		}
		for i := 0; i < limit; i++ {
			fc := sess.FileChanges[i]
			b.WriteString(fmt.Sprintf("  - [%s] %s", fc.ChangeType, fc.Path))
			if fc.Risk == session.RiskDanger {
				b.WriteString(" ⚠️高风险")
			}
			b.WriteString("\n")
			// 包含 diff 摘要（截断）
			if fc.Diff != "" {
				diffLines := strings.Split(fc.Diff, "\n")
				if len(diffLines) > 10 {
					diffLines = diffLines[:10]
					diffLines = append(diffLines, fmt.Sprintf("  ... (共 %d 行)", len(strings.Split(fc.Diff, "\n"))))
				}
				for _, line := range diffLines {
					b.WriteString(fmt.Sprintf("    %s\n", line))
				}
			}
		}
		if len(sess.FileChanges) > limit {
			b.WriteString(fmt.Sprintf("  ... 还有 %d 个文件未展示\n", len(sess.FileChanges)-limit))
		}
	}

	// 执行的命令
	var commands []string
	for _, action := range sess.Actions {
		if action.Type == session.ActionBash {
			if cmd, ok := action.Input["command"].(string); ok {
				commands = append(commands, truncateUTF8(cmd, 200))
			}
		}
	}
	if len(commands) > 0 {
		b.WriteString(fmt.Sprintf("\n执行的命令 (%d 个):\n", len(commands)))
		limit := 15
		if len(commands) < limit {
			limit = len(commands)
		}
		for i := 0; i < limit; i++ {
			b.WriteString(fmt.Sprintf("  - %s\n", commands[i]))
		}
		if len(commands) > limit {
			b.WriteString(fmt.Sprintf("  ... 还有 %d 条命令未展示\n", len(commands)-limit))
		}
	}

	return b.String()
}

// parseRulesResponse 解析 LLM 返回的规则提取结果。
func parseRulesResponse(content string) ([]ComplianceRule, error) {
	content = extractJSONBlock(content)

	var result struct {
		Rules []ComplianceRule `json:"rules"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w\n原始内容: %s", err, truncateUTF8(content, 500))
	}

	return result.Rules, nil
}

// parseAuditResponse 解析 LLM 返回的审计结果。
func parseAuditResponse(content string) (*ComplianceScore, error) {
	content = extractJSONBlock(content)

	var score ComplianceScore
	if err := json.Unmarshal([]byte(content), &score); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w\n原始内容: %s", err, truncateUTF8(content, 500))
	}

	return &score, nil
}

// extractJSONBlock 从文本中提取 JSON 块。
func extractJSONBlock(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}

// truncateUTF8 截断字符串到指定长度。
func truncateUTF8(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// extractRulesSystemPrompt 是提取规则的系统提示词。
const extractRulesSystemPrompt = `你是一个 CLAUDE.md 规则提取专家。你的任务是从 CLAUDE.md 文件内容中提取所有行为规则。

请严格按以下JSON格式返回（不要包含markdown代码块标记）：
{
  "rules": [
    {
      "rule": "规则原文或精简表述",
      "category": "类别",
      "description": "规则的具体含义和目的"
    }
  ]
}

category 必须是以下之一：
- "file": 文件操作相关规则（不要修改xxx、禁止删除xxx等）
- "command": 命令执行相关规则（不要运行xxx、禁止使用sudo等）
- "format": 代码格式相关规则（缩进、空格、换行等）
- "naming": 命名规范相关规则（变量名、函数名、文件名等）
- "convention": 编码规范和工作流程（提交规范、分支策略、测试要求等）
- "other": 其他类型的行为规则

注意事项：
1. 提取所有明确的行为规则，包括"禁止"、"不要"、"必须"、"应该"等表述
2. 忽略项目概述、架构描述、技术栈说明等非规则内容
3. 将复杂的多行规则简化为清晰的单条规则
4. 忽略过于笼统的规则（如"写好代码"），只提取可执行的具体规则
5. 严格控制输出量：最多提取 30 条规则`

// auditSessionSystemPrompt 是审计会话的系统提示词。
const auditSessionSystemPrompt = `你是一个 CLAUDE.md 规则合规审计专家。你的任务是根据 CLAUDE.md 中的规则，检查一个 Claude Code 会话是否遵守了这些规则。

请严格按以下JSON格式返回（不要包含markdown代码块标记）：
{
  "overall": 85.0,
  "totalRules": 10,
  "compliedCount": 8,
  "violations": [
    {
      "rule": "违反的规则描述",
      "compliant": false,
      "severity": "high",
      "evidence": "具体的违规证据和说明"
    }
  ]
}

评分规则：
- overall = compliedCount / totalRules * 100
- 对于无法判断的规则（会话中没有涉及该规则相关操作），视为"未违反"（compliant=true）
- violations 只列出确实违规的规则

severity 级别：
- "high": 严重违规（如删除关键文件、执行危险命令、泄露敏感信息）
- "medium": 中度违规（如格式不符、命名不规范）
- "low": 轻度违规（如建议未遵循、最佳实践未采用）

注意事项：
1. 仔细检查文件变更（FileChanges）和执行的命令（Commands）是否违反规则
2. 对于"禁止修改xxx"类规则，检查是否有修改该文件的操作
3. 对于"不要运行xxx"类规则，检查是否有执行该命令的操作
4. 对于格式和命名规则，检查文件变更中的diff内容
5. 证据(evidence)要具体说明违规的文件路径、命令内容等
6. 如果某个规则在会话中没有被触及（既没违反也没遵守的证据），标记为compliant=true
7. 严格控制输出量：violations 最多列出 15 条`
