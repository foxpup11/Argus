package session

import (
	"time"
)

// SessionMeta 会话元数据（标签、收藏等）
type SessionMeta struct {
	SessionID string    `json:"sessionId"`
	Tags      []string  `json:"tags,omitempty"`      // 用户手动添加的标签
	AutoTags  []string  `json:"autoTags,omitempty"`  // 自动识别的标签（按任务类型）
	Favorited bool      `json:"favorited"`           // 是否收藏
	Note      string    `json:"note,omitempty"`       // 用户备注
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Tag 标签定义
type Tag struct {
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`       // 标签颜色（可选）
	Description string `json:"description,omitempty"` // 标签描述（可选）
}

// AutoTagRule 自动标签规则
type AutoTagRule struct {
	Pattern string   `json:"pattern"` // 匹配模式（正则或关键词）
	Tag     string   `json:"tag"`     // 对应的标签
	Fields  []string `json:"fields"`  // 匹配的字段：prompt, filePath, command
}

// DefaultAutoTags 默认的自动标签规则
var DefaultAutoTags = []AutoTagRule{
	// 按任务类型
	{Pattern: "(?i)(fix|bug|error|修复|错误|bug)", Tag: "bug-fix", Fields: []string{"prompt"}},
	{Pattern: "(?i)(feat|feature|新增|添加|实现)", Tag: "feature", Fields: []string{"prompt"}},
	{Pattern: "(?i)(refactor|重构|优化|improve)", Tag: "refactor", Fields: []string{"prompt"}},
	{Pattern: "(?i)(test|测试|spec)", Tag: "testing", Fields: []string{"prompt"}},
	{Pattern: "(?i)(doc|文档|readme)", Tag: "documentation", Fields: []string{"prompt"}},
	{Pattern: "(?i)(style|css|样式|ui|界面)", Tag: "ui-style", Fields: []string{"prompt", "filePath"}},
	{Pattern: "(?i)(config|配置|setting|setup)", Tag: "configuration", Fields: []string{"prompt", "filePath"}},
	{Pattern: "(?i)(deploy|部署|ci|cd)", Tag: "devops", Fields: []string{"prompt", "command"}},
	{Pattern: "(?i)(performance|性能|优化|speed)", Tag: "performance", Fields: []string{"prompt"}},
	{Pattern: "(?i)(security|安全|vuln)", Tag: "security", Fields: []string{"prompt"}},

	// 按文件类型
	{Pattern: "\\.go$", Tag: "golang", Fields: []string{"filePath"}},
	{Pattern: "\\.js$|\\.ts$|\\.jsx$|\\.tsx$", Tag: "javascript", Fields: []string{"filePath"}},
	{Pattern: "\\.py$", Tag: "python", Fields: []string{"filePath"}},
	{Pattern: "\\.css$|\\.scss$|\\.less$", Tag: "css", Fields: []string{"filePath"}},
	{Pattern: "\\.html$|\\.vue$", Tag: "markup", Fields: []string{"filePath"}},
	{Pattern: "(?i)(docker|k8s|kubernetes)", Tag: "devops", Fields: []string{"prompt", "command"}},
	{Pattern: "(?i)(sql|database|migration)", Tag: "database", Fields: []string{"prompt", "command"}},
}

// SearchQuery 搜索查询参数
type SearchQuery struct {
	Keyword  string   `json:"keyword"`            // 搜索关键词
	Fields   []string `json:"fields,omitempty"`    // 搜索字段：prompt, filePath, command, tags（空则搜索所有）
	Tags     []string `json:"tags,omitempty"`      // 按标签过滤
	Favorited *bool   `json:"favorited,omitempty"` // 按收藏状态过滤
	Projects []string `json:"projects,omitempty"`  // 按项目过滤
}

// SearchResult 搜索结果
type SearchResult struct {
	SessionID string   `json:"sessionId"`
	Matches   []string `json:"matches"` // 匹配的字段和内容
	Score     float64  `json:"score"`   // 匹配分数
}

// BatchOperation 批量操作请求
type BatchOperation struct {
	Action    string   `json:"action"`    // "delete", "export", "tag", "untag", "favorite", "unfavorite"
	SessionIDs []string `json:"sessionIds"`
	Tag       string   `json:"tag,omitempty"`       // 用于 tag/untag 操作
	Format    string   `json:"format,omitempty"`    // 用于 export 操作
	OutputDir string   `json:"outputDir,omitempty"` // 用于 export 操作
}

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}
