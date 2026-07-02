package continuity

import (
	"regexp"
	"strings"
	"time"

	"agentscope-desktop/internal/session"
)

// Extractor 从会话中提取任务信息
type Extractor struct{}

// NewExtractor 创建新的提取器
func NewExtractor() *Extractor {
	return &Extractor{}
}

// filePatterns 匹配文件路径的正则
var filePathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[\w\-/\\.]+\.\w{1,10}`), // 一般文件路径
}

// taskTypeKeywords 任务类型关键词映射
var taskTypeKeywords = map[string][]string{
	"feature":    {"添加", "新增", "实现", "创建", "add", "implement", "create", "build", "develop"},
	"bugfix":     {"修复", "修正", "解决", "fix", "repair", "resolve", "debug", "patch"},
	"refactor":   {"重构", "优化", "清理", "refactor", "optimize", "cleanup", "reorganize"},
	"docs":       {"文档", "注释", "说明", "document", "comment", "readme", "doc"},
	"test":       {"测试", "用例", "验证", "test", "spec", "assert", "verify"},
	"config":     {"配置", "设置", "环境", "config", "setting", "env", "setup"},
	"dependency": {"依赖", "升级", "安装", "depend", "upgrade", "install", "package"},
}

// decisionKeywords 决策相关的关键词
var decisionKeywords = []string{
	"决定", "选择", "方案", "决定用", "最终",
	"decide", "choose", "decision", "approach", "strategy", "方案",
	"convention", "standard", "pattern",
}

// issueKeywords 已知问题/陷阱的关键词
var issueKeywords = []string{
	"注意", "警告", "陷阱", "坑", "避免", "不要", "不能",
	"warn", "caution", "pitfall", "avoid", "don't", "must not", "caveat",
	"issue", "problem", "bug", "workaround",
}

// ExtractTasks 从多个会话中提取任务
func (e *Extractor) ExtractTasks(sessions []*session.Session) []PromptAnalysis {
	var analyses []PromptAnalysis

	for _, sess := range sessions {
		// 提取所有用户 prompt
		prompts := e.extractPrompts(sess)
		analyses = append(analyses, prompts...)
	}

	return analyses
}

// extractPrompts 从单个会话中提取用户 prompt
func (e *Extractor) extractPrompts(sess *session.Session) []PromptAnalysis {
	var analyses []PromptAnalysis
	seen := make(map[string]bool)

	for _, msg := range sess.Messages {
		if msg.Type != session.MessageTypeUser {
			continue
		}

		for _, block := range msg.Content {
			if block.Type != session.ContentTypeText || block.Text == "" {
				continue
			}

			text := strings.TrimSpace(block.Text)
			if text == "" || seen[text] {
				continue
			}
			seen[text] = true

			// 跳过太短的 prompt（通常是确认性回复）
			if len(text) < 5 {
				continue
			}

			// 跳过系统消息模式（以 [ 开头的通常是工具结果回显）
			if strings.HasPrefix(text, "[") {
				continue
			}

			analyses = append(analyses, PromptAnalysis{
				Text:        text,
				Timestamp:   msg.Timestamp,
				SessionID:   sess.ID,
				TaskType:    classifyTaskType(text),
				HasFilePath: containsFilePath(text),
			})
		}
	}

	return analyses
}

// ClassifyTaskType 分类任务类型
func classifyTaskType(text string) string {
	lower := strings.ToLower(text)
	for taskType, keywords := range taskTypeKeywords {
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return taskType
			}
		}
	}
	return "general"
}

// containsFilePath 检查文本是否包含文件路径
func containsFilePath(text string) bool {
	for _, re := range filePathPatterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

// ExtractCompletedTasks 从会话的 actions 中提取已完成的任务
func (e *Extractor) ExtractCompletedTasks(sessions []*session.Session) []CompletedTask {
	var tasks []CompletedTask

	for _, sess := range sessions {
		if len(sess.Actions) == 0 {
			continue
		}

		// 将 actions 按文件路径分组
		fileActions := make(map[string][]session.Action)
		var nonFileActions []session.Action

		for _, action := range sess.Actions {
			if action.FilePath != "" {
				fileActions[action.FilePath] = append(fileActions[action.FilePath], action)
			} else {
				nonFileActions = append(nonFileActions, action)
			}
		}

		// 从用户 prompt 中提取任务描述
		taskDesc := e.inferTaskDescription(sess)

		// 收集所有涉及的文件
		var filesChanged []string
		for filePath := range fileActions {
			filesChanged = append(filesChanged, filePath)
		}

		if taskDesc != "" && len(filesChanged) > 0 {
			tasks = append(tasks, CompletedTask{
				Description:  taskDesc,
				SessionID:    sess.ID,
				FilesChanged: filesChanged,
				Timestamp:    sess.StartedAt,
			})
		}
	}

	return tasks
}

// inferTaskDescription 从会话中推断任务描述
func (e *Extractor) inferTaskDescription(sess *session.Session) string {
	// 优先使用用户 prompt
	if sess.Prompt != "" {
		// 截取第一行作为描述
		lines := strings.SplitN(sess.Prompt, "\n", 2)
		desc := strings.TrimSpace(lines[0])
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		return desc
	}

	// 回退到第一个有内容的用户消息
	for _, msg := range sess.Messages {
		if msg.Type != session.MessageTypeUser {
			continue
		}
		for _, block := range msg.Content {
			if block.Type == session.ContentTypeText && block.Text != "" {
				text := strings.TrimSpace(block.Text)
				if len(text) > 10 {
					lines := strings.SplitN(text, "\n", 2)
					desc := strings.TrimSpace(lines[0])
					if len(desc) > 200 {
						desc = desc[:200] + "..."
					}
					return desc
				}
			}
		}
	}

	return ""
}

// ExtractDecisions 从会话中提取关键决策
func (e *Extractor) ExtractDecisions(sessions []*session.Session) []Decision {
	var decisions []Decision

	for _, sess := range sessions {
		for _, msg := range sess.Messages {
			if msg.Type != session.MessageTypeAssistant {
				continue
			}

			for _, block := range msg.Content {
				if block.Type != session.ContentTypeText {
					continue
				}

				// 在 assistant 消息中查找决策关键词
				if containsAnyKeyword(block.Text, decisionKeywords) {
					sentences := extractRelevantSentences(block.Text, decisionKeywords)
					for _, sentence := range sentences {
						decisions = append(decisions, Decision{
							Description: sentence,
							Context:     truncate(block.Text, 300),
							Timestamp:   msg.Timestamp,
							SessionID:   sess.ID,
						})
					}
				}
			}
		}
	}

	return decisions
}

// ExtractKnownIssues 从会话中提取已知问题
func (e *Extractor) ExtractKnownIssues(sessions []*session.Session) []string {
	issueSet := make(map[string]bool)
	var issues []string

	for _, sess := range sessions {
		for _, msg := range sess.Messages {
			// 检查 assistant 消息
			if msg.Type == session.MessageTypeAssistant {
				for _, block := range msg.Content {
					if block.Type == session.ContentTypeText && containsAnyKeyword(block.Text, issueKeywords) {
						sentences := extractRelevantSentences(block.Text, issueKeywords)
						for _, s := range sentences {
							if !issueSet[s] {
								issueSet[s] = true
								issues = append(issues, s)
							}
						}
					}
				}
			}

			// 也检查用户消息中的注意事项
			if msg.Type == session.MessageTypeUser {
				for _, block := range msg.Content {
					if block.Type == session.ContentTypeText && containsAnyKeyword(block.Text, issueKeywords) {
						sentences := extractRelevantSentences(block.Text, issueKeywords)
						for _, s := range sentences {
							if !issueSet[s] {
								issueSet[s] = true
								issues = append(issues, s)
							}
						}
					}
				}
			}
		}
	}

	return issues
}

// ExtractPendingTasks 从会话中提取可能的待办任务
func (e *Extractor) ExtractPendingTasks(sessions []*session.Session) []PendingTask {
	var tasks []PendingTask
	seen := make(map[string]bool)

	for _, sess := range sessions {
		// 从用户 prompt 中检测 TODO/FIXME/HACK 等模式
		for _, msg := range sess.Messages {
			if msg.Type != session.MessageTypeUser {
				continue
			}
			for _, block := range msg.Content {
				if block.Type != session.ContentTypeText {
					continue
				}
				text := block.Text
				if containsTODOPattern(text) {
					desc := extractTODODescription(text)
					if desc != "" && !seen[desc] {
						seen[desc] = true
						tasks = append(tasks, PendingTask{
							Description: desc,
							Source:      "user_prompt",
							SessionID:   sess.ID,
						})
					}
				}
			}
		}

		// 检测 assistant 消息中的未完成承诺
		for _, msg := range sess.Messages {
			if msg.Type != session.MessageTypeAssistant {
				continue
			}
			for _, block := range msg.Content {
				if block.Type != session.ContentTypeText {
					continue
				}
				if containsUnfinishedPromise(block.Text) {
					desc := extractPromiseDescription(block.Text)
					if desc != "" && !seen[desc] {
						seen[desc] = true
						tasks = append(tasks, PendingTask{
							Description: desc,
							Source:      "assistant_promise",
							SessionID:   sess.ID,
						})
					}
				}
			}
		}
	}

	return tasks
}

// BuildFileSummaries 构建文件概览
func (e *Extractor) BuildFileSummaries(sessions []*session.Session) []FileSummary {
	fileMap := make(map[string]*FileSummary)

	for _, sess := range sessions {
		for _, action := range sess.Actions {
			if action.FilePath == "" {
				continue
			}

			fs, ok := fileMap[action.FilePath]
			if !ok {
				fs = &FileSummary{
					Path:         action.FilePath,
					IsTestFile:   isTestFile(action.FilePath),
					IsConfigFile: isConfigFile(action.FilePath),
				}
				fileMap[action.FilePath] = fs
			}

			fs.ChangeCount++
			fs.ActionCount++
			fs.LastAction = string(action.Type)
		}
	}

	// 转换为切片并按操作次数排序
	summaries := make([]FileSummary, 0, len(fileMap))
	for _, fs := range fileMap {
		summaries = append(summaries, *fs)
	}

	// 简单的冒泡排序，按 ActionCount 降序
	for i := 0; i < len(summaries); i++ {
		for j := i + 1; j < len(summaries); j++ {
			if summaries[j].ActionCount > summaries[i].ActionCount {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}

	return summaries
}

// 辅助函数

func containsAnyKeyword(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func extractRelevantSentences(text string, keywords []string) []string {
	sentences := strings.Split(text, "。")
	sentences = append(sentences, strings.Split(text, ". ")...)

	var relevant []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s == "" || len(s) < 10 {
			continue
		}
		lower := strings.ToLower(s)
		for _, kw := range keywords {
			if strings.Contains(lower, strings.ToLower(kw)) {
				if len(s) > 200 {
					s = s[:200] + "..."
				}
				relevant = append(relevant, s)
				break
			}
		}
	}

	return relevant
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func containsTODOPattern(text string) bool {
	lower := strings.ToLower(text)
	todoPatterns := []string{"todo", "fixme", "hack", "需要完成", "待完成", "后续需要", "还需要", "待处理"}
	for _, p := range todoPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func extractTODODescription(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if strings.Contains(lower, "todo") || strings.Contains(lower, "fixme") || strings.Contains(lower, "hack") ||
			strings.Contains(lower, "待完成") || strings.Contains(lower, "需要完成") || strings.Contains(lower, "还需要") {
			line = strings.TrimSpace(line)
			if len(line) > 200 {
				line = line[:200] + "..."
			}
			return line
		}
	}
	return ""
}

func containsUnfinishedPromise(text string) bool {
	lower := strings.ToLower(text)
	patterns := []string{"接下来", "下一步", "然后我会", "随后", "之后", "next, i will", "then i'll", "will then", "next step"}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func extractPromiseDescription(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if strings.Contains(lower, "下一步") || strings.Contains(lower, "接下来") || strings.Contains(lower, "then i") || strings.Contains(lower, "next step") {
			line = strings.TrimSpace(line)
			if len(line) > 200 {
				line = line[:200] + "..."
			}
			return line
		}
	}
	return ""
}

func isTestFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "_test") ||
		strings.Contains(lower, ".test.") ||
		strings.Contains(lower, "_spec") ||
		strings.Contains(lower, ".spec.") ||
		strings.Contains(lower, "/test/") ||
		strings.Contains(lower, "/tests/") ||
		strings.Contains(lower, "/__tests__/")
}

func isConfigFile(path string) bool {
	lower := strings.ToLower(path)
	configPatterns := []string{
		".env", "config.", "setting.", ".json", ".yaml", ".yml", ".toml",
		"dockerfile", "docker-compose", "makefile", ".gitignore",
		"go.mod", "go.sum", "package.json", "package-lock.json",
		"tsconfig", "webpack", "vite.config", ".eslint", ".prettier",
	}
	for _, p := range configPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// DeduplicateTasks 去重任务列表
func DeduplicateTasks(tasks []CompletedTask) []CompletedTask {
	seen := make(map[string]bool)
	var result []CompletedTask

	for _, t := range tasks {
		key := t.Description
		if !seen[key] {
			seen[key] = true
			result = append(result, t)
		}
	}

	return result
}

// DeduplicateDecisions 去重决策列表
func DeduplicateDecisions(decisions []Decision) []Decision {
	seen := make(map[string]bool)
	var result []Decision

	for _, d := range decisions {
		key := d.Description
		if !seen[key] {
			seen[key] = true
			result = append(result, d)
		}
	}

	return result
}

// DeduplicateIssues 去重已知问题
func DeduplicateIssues(issues []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, issue := range issues {
		if !seen[issue] {
			seen[issue] = true
			result = append(result, issue)
		}
	}

	return result
}

// filterSessionsByProject 按项目过滤会话
func FilterSessionsByProject(sessions []*session.Session, projectDir string) []*session.Session {
	var filtered []*session.Session
	for _, sess := range sessions {
		if strings.Contains(sess.CWD, projectDir) || projectDir == "" {
			filtered = append(filtered, sess)
		}
	}
	return filtered
}

// sortSessionsByTime 按时间排序会话（最新的在前）
func SortSessionsByTime(sessions []*session.Session) {
	timeSortFunc := func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	}
	_ = timeSortFunc // 避免未使用警告

	// 使用标准库排序
	for i := 0; i < len(sessions); i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[j].StartedAt.After(sessions[i].StartedAt) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}
}

// FilterRecentSessions 只保留最近 N 个会话
func FilterRecentSessions(sessions []*session.Session, count int) []*session.Session {
	if count <= 0 || count >= len(sessions) {
		return sessions
	}
	return sessions[:count]
}

// filterByTimeRange 按时间范围过滤会话
func FilterByTimeRange(sessions []*session.Session, since time.Time) []*session.Session {
	var filtered []*session.Session
	for _, sess := range sessions {
		if sess.StartedAt.After(since) || sess.StartedAt.Equal(since) {
			filtered = append(filtered, sess)
		}
	}
	return filtered
}
