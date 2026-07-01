// Package analytics provides Token usage analytics for Claude Code sessions.
package analytics

// TokenOverview holds the high-level dashboard summary data.
type TokenOverview struct {
	TotalInputTokens  int            `json:"totalInputTokens"`
	TotalOutputTokens int            `json:"totalOutputTokens"`
	TotalTokens       int            `json:"totalTokens"`
	TotalSessions     int            `json:"totalSessions"`
	TodayTokens       int            `json:"todayTokens"`
	ThisMonthTokens   int            `json:"thisMonthTokens"`
	LastMonthTokens   int            `json:"lastMonthTokens"`
	MonthTokenChange  float64        `json:"monthTokenChange"`
	ProjectBreakdown  []ProjectStats `json:"projectBreakdown"`
	ModelBreakdown    []ModelStats   `json:"modelBreakdown"`
	DailyTrend        []DailyUsage   `json:"dailyTrend"`
}

// ProjectStats holds per-project token usage.
type ProjectStats struct {
	ProjectDir   string `json:"projectDir"`
	ProjectName  string `json:"projectName"`
	SessionCount int    `json:"sessionCount"`
	TotalTokens  int    `json:"totalTokens"`
	InputTokens  int    `json:"inputTokens"`
	OutputTokens int    `json:"outputTokens"`
}

// ModelStats holds per-model token usage.
type ModelStats struct {
	Model        string `json:"model"`
	SessionCount int    `json:"sessionCount"`
	TotalTokens  int    `json:"totalTokens"`
	InputTokens  int    `json:"inputTokens"`
	OutputTokens int    `json:"outputTokens"`
}

// DailyUsage holds per-day aggregated usage.
type DailyUsage struct {
	Date         string `json:"date"`
	Tokens       int    `json:"tokens"`
	InputTokens  int    `json:"inputTokens"`
	OutputTokens int    `json:"outputTokens"`
	SessionCount int    `json:"sessionCount"`
}
