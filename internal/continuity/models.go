// Package continuity provides session continuity engine capabilities for
// generating handoff summaries across Claude Code sessions.
package continuity

import (
	"time"
)

// MaxSummaryLines 交接摘要 Markdown 输出的最大行数
const MaxSummaryLines = 140

// HandoffSummary 会话移交摘要
type HandoffSummary struct {
	Project        string          `json:"project"`        // 项目目录名
	SessionsUsed   int             `json:"sessionsUsed"`   // 参与分析的会话数
	SessionsTotal  int             `json:"sessionsTotal"`  // 项目总会话数
	Summary        string          `json:"summary"`        // 会话核心内容摘要
	CompletedTasks []CompletedTask  `json:"completedTasks"` // 已完成任务
	PendingTasks   []PendingTask    `json:"pendingTasks"`   // 待办任务
	KeyDecisions   []Decision       `json:"keyDecisions"`   // 关键决策
	ModifiedFiles  []FileSummary    `json:"modifiedFiles"`  // 修改的文件概览
	KnownIssues    []string         `json:"knownIssues"`    // 已知问题/陷阱
	GeneratedAt    time.Time        `json:"generatedAt"`    // 生成时间
	Quality        SummaryQuality   `json:"quality"`        // 摘要质量评分
	LLMEnhanced    bool             `json:"llmEnhanced"`    // 是否使用了 LLM 增强
}

// SummaryQuality 摘要质量评分
type SummaryQuality struct {
	Completeness float64 `json:"completeness"` // 完整性评分（0-1）
	Accuracy     float64 `json:"accuracy"`     // 准确性评分（0-1）
	Freshness    float64 `json:"freshness"`    // 时效性评分（0-1）
	OverallScore float64 `json:"overallScore"` // 综合评分（0-1）
}

// CompletedTask 已完成的任务
type CompletedTask struct {
	Description   string   `json:"description"`   // 任务描述
	SessionID     string   `json:"sessionId"`      // 来源会话 ID
	FilesChanged  []string `json:"filesChanged"`   // 相关文件
	VerifiedByGit bool     `json:"verifiedByGit"`  // 是否被 git 记录验证
	Timestamp     time.Time `json:"timestamp"`     // 完成时间
}

// PendingTask 待办任务
type PendingTask struct {
	Description string   `json:"description"` // 任务描述
	Source      string   `json:"source"`      // 来源（用户 prompt 或未完成的 action）
	SessionID   string   `json:"sessionId"`   // 来源会话 ID
	FilesHint   []string `json:"filesHint"`   // 相关文件提示
}

// Decision 关键决策
type Decision struct {
	Description string    `json:"description"` // 决策内容
	Context     string    `json:"context"`     // 决策背景
	Timestamp   time.Time `json:"timestamp"`   // 决策时间
	SessionID   string    `json:"sessionId"`   // 来源会话 ID
}

// FileSummary 文件概览
type FileSummary struct {
	Path         string `json:"path"`         // 文件路径
	ChangeCount  int    `json:"changeCount"`  // 被修改的次数
	ActionCount  int    `json:"actionCount"`  // Agent 操作次数
	LastAction   string `json:"lastAction"`   // 最后一次操作类型
	IsTestFile   bool   `json:"isTestFile"`   // 是否是测试文件
	IsConfigFile bool   `json:"isConfigFile"` // 是否是配置文件
}

// PromptAnalysis 用户 prompt 分析结果
type PromptAnalysis struct {
	Text        string    `json:"text"`        // prompt 原文
	Timestamp   time.Time `json:"timestamp"`   // 时间
	SessionID   string    `json:"sessionId"`   // 所属会话
	TaskType    string    `json:"taskType"`    // 任务类型分类
	HasFilePath bool      `json:"hasFilePath"` // 是否包含文件路径
}

// ============================================
// LLM 增强相关类型
// ============================================

// LLMEnhancedResult 表示 LLM 返回的结构化分析结果。
type LLMEnhancedResult struct {
	NarrativeSummary string            `json:"narrativeSummary"` // 连贯的叙事摘要
	CompletedTasks   []LLMTaskItem     `json:"completedTasks"`   // LLM 识别完成的任务
	PendingTasks     []LLMPendingItem  `json:"pendingTasks"`     // LLM 识别待办
	KeyDecisions     []LLMDecisionItem `json:"keyDecisions"`     // LLM 识别决策
	KnownIssues      []string          `json:"knownIssues"`      // LLM 识别问题
}

// LLMTaskItem LLM 识别的一条已完成任务。
type LLMTaskItem struct {
	Description string   `json:"description"`
	FilesHint   []string `json:"filesHint"`
	Confidence  float64  `json:"confidence"`
}

// LLMPendingItem LLM 识别的一条待办任务。
type LLMPendingItem struct {
	Description string  `json:"description"`
	Priority    string  `json:"priority"` // "high" / "medium" / "low"
	Source      string  `json:"source"`   // "explicit" / "implicit"
	Confidence  float64 `json:"confidence"`
}

// LLMDecisionItem LLM 识别的一条关键决策。
type LLMDecisionItem struct {
	Description string  `json:"description"`
	Rationale   string  `json:"rationale"`
	Confidence  float64 `json:"confidence"`
}
