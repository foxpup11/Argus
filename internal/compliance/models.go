// Package compliance provides LLM-powered CLAUDE.md rule compliance auditing.
package compliance

// Severity 违规严重程度
type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

// ComplianceRule 从 CLAUDE.md 提取的合规规则（由 LLM 生成）
type ComplianceRule struct {
	Rule        string `json:"rule"`        // 规则原文
	Category    string `json:"category"`    // 类别: file, command, format, naming, convention, other
	Description string `json:"description"` // 规则描述
}

// AuditResult 单条规则的审计结果
type AuditResult struct {
	Rule        string   `json:"rule"`        // 规则描述
	Compliant   bool     `json:"compliant"`   // 是否合规
	Severity    Severity `json:"severity"`    // 违规严重程度
	Evidence    string   `json:"evidence"`    // 证据（违规/合规说明）
}

// ComplianceScore 合规评分
type ComplianceScore struct {
	Overall       float64       `json:"overall"`       // 总体分数 0-100
	TotalRules    int           `json:"totalRules"`    // 规则总数
	CompliedCount int           `json:"compliedCount"` // 合规规则数
	Violations    []AuditResult `json:"violations"`    // 违规详情
}

// ComplianceOverview 合规概览（所有会话聚合）
type ComplianceOverview struct {
	AverageScore    float64              `json:"averageScore"`    // 平均合规分数
	TotalSessions   int                  `json:"totalSessions"`   // 总会话数
	AuditedSessions int                  `json:"auditedSessions"` // 已审计会话数
	Violations      []ViolationStat      `json:"violations"`      // 最常见违规
	Sessions        []SessionAuditResult `json:"sessions"`        // 每个会话的审计结果
}

// SessionAuditResult 单个会话的审计结果
type SessionAuditResult struct {
	SessionID string           `json:"sessionId"`
	Score     *ComplianceScore `json:"score"`
}

// ViolationStat 违规统计
type ViolationStat struct {
	Rule        string   `json:"rule"`        // 规则描述
	Count       int      `json:"count"`       // 违规次数
	Severity    Severity `json:"severity"`    // 严重程度
}

// AuditRequest 前端发起审计的请求
type AuditRequest struct {
	ClaudeMDPath string `json:"claudeMDPath"`
	SessionID    string `json:"sessionId,omitempty"` // 可选，指定单个会话
}
