// Package continuity provides session continuity engine capabilities for
// generating handoff summaries across Claude Code sessions.
package continuity

import (
	"time"
)

// HandoffSummary 会话移交摘要
type HandoffSummary struct {
	Project       string          `json:"project"`       // 项目目录名
	SessionsUsed  int             `json:"sessionsUsed"`  // 参与分析的会话数
	SessionsTotal int             `json:"sessionsTotal"` // 项目总会话数
	CompletedTasks []CompletedTask `json:"completedTasks"` // 已完成任务
	PendingTasks   []PendingTask   `json:"pendingTasks"`   // 待办任务
	KeyDecisions   []Decision      `json:"keyDecisions"`   // 关键决策
	ModifiedFiles  []FileSummary   `json:"modifiedFiles"`  // 修改的文件概览
	KnownIssues    []string        `json:"knownIssues"`    // 已知问题/陷阱
	GeneratedAt    time.Time       `json:"generatedAt"`    // 生成时间
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
