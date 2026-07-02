package continuity

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agentscope-desktop/internal/session"
	"agentscope-desktop/internal/session/claude"
)

// Engine 会话连续性引擎
type Engine struct {
	extractor     *Extractor
	validator     *Validator
	handoffGen    *HandoffGenerator
	homeDir       string
}

// NewEngine 创建新的连续性引擎
func NewEngine() (*Engine, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户目录失败: %w", err)
	}

	validator, err := NewValidator()
	if err != nil {
		return nil, fmt.Errorf("创建验证器失败: %w", err)
	}

	handoffGen, err := NewHandoffGenerator()
	if err != nil {
		return nil, fmt.Errorf("创建手交生成器失败: %w", err)
	}

	return &Engine{
		extractor:  NewExtractor(),
		validator:  validator,
		handoffGen: handoffGen,
		homeDir:    homeDir,
	}, nil
}

// GenerateHandoff 生成会话交接摘要
func (e *Engine) GenerateHandoff(projectDir string, sessionCount int) (*HandoffSummary, error) {
	// 加载项目的所有会话
	allSessions, err := e.loadProjectSessions(projectDir)
	if err != nil {
		return nil, fmt.Errorf("加载会话失败: %w", err)
	}

	if len(allSessions) == 0 {
		return nil, fmt.Errorf("项目 %s 没有会话数据", projectDir)
	}

	// 按时间排序（最新在前）
	SortSessionsByTime(allSessions)

	// 限制会话数量
	totalSessions := len(allSessions)
	if sessionCount > 0 && sessionCount < totalSessions {
		allSessions = FilterRecentSessions(allSessions, sessionCount)
	}

	// 提取各项信息
	completedTasks := e.extractor.ExtractCompletedTasks(allSessions)
	completedTasks = DeduplicateTasks(completedTasks)

	pendingTasks := e.extractor.ExtractPendingTasks(allSessions)

	decisions := e.extractor.ExtractDecisions(allSessions)
	decisions = DeduplicateDecisions(decisions)

	fileSummaries := e.extractor.BuildFileSummaries(allSessions)

	issues := e.extractor.ExtractKnownIssues(allSessions)
	issues = DeduplicateIssues(issues)

	// Git 交叉验证
	if len(allSessions) > 0 {
		cwd := allSessions[0].CWD
		completedTasks = e.validator.ValidateTasks(completedTasks, cwd)
		completedTasks = e.validator.ValidateAgainstSessions(completedTasks, allSessions)
	}

	// 构建摘要
	summary := &HandoffSummary{
		Project:        projectDir,
		SessionsUsed:   len(allSessions),
		SessionsTotal:  totalSessions,
		CompletedTasks: completedTasks,
		PendingTasks:   pendingTasks,
		KeyDecisions:   decisions,
		ModifiedFiles:  fileSummaries,
		KnownIssues:    issues,
		GeneratedAt:    time.Now(),
	}

	return summary, nil
}

// GenerateHandoffMarkdown 生成 Markdown 格式的交接摘要
func (e *Engine) GenerateHandoffMarkdown(projectDir string, sessionCount int) (string, *HandoffSummary, error) {
	summary, err := e.GenerateHandoff(projectDir, sessionCount)
	if err != nil {
		return "", nil, err
	}

	markdown := e.handoffGen.GenerateMarkdown(summary)
	return markdown, summary, nil
}

// GenerateHandoffPrompt 生成可粘贴的 prompt 片段
func (e *Engine) GenerateHandoffPrompt(projectDir string, sessionCount int) (string, *HandoffSummary, error) {
	summary, err := e.GenerateHandoff(projectDir, sessionCount)
	if err != nil {
		return "", nil, err
	}

	prompt := e.handoffGen.GeneratePrompt(summary)
	return prompt, summary, nil
}

// ExportToMemory 导出交接摘要到 memory 目录
func (e *Engine) ExportToMemory(projectDir string, sessionCount int) (string, error) {
	summary, err := e.GenerateHandoff(projectDir, sessionCount)
	if err != nil {
		return "", err
	}

	markdown := e.handoffGen.GenerateMarkdown(summary)
	return e.handoffGen.SaveToMemory(summary, markdown)
}

// GetAvailableProjects 获取所有有会话的项目列表
func (e *Engine) GetAvailableProjects() ([]ProjectInfo, error) {
	claudeDir := filepath.Join(e.homeDir, ".claude", "projects")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return []ProjectInfo{}, nil
	}

	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return nil, fmt.Errorf("读取 Claude 项目目录失败: %w", err)
	}

	var projects []ProjectInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectDir := filepath.Join(claudeDir, entry.Name())
		jsonlFiles, _ := filepath.Glob(filepath.Join(projectDir, "*.jsonl"))
		if len(jsonlFiles) == 0 {
			continue
		}

		// 获取最近的会话时间
		var lastActivity time.Time
		for _, f := range jsonlFiles {
			info, err := os.Stat(f)
			if err == nil && info.ModTime().After(lastActivity) {
				lastActivity = info.ModTime()
			}
		}

		projects = append(projects, ProjectInfo{
			Name:         formatProjectDirName(entry.Name()),
			DirName:      entry.Name(),
			SessionCount: len(jsonlFiles),
			LastActivity: lastActivity,
		})
	}

	// 按最后活动时间排序
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastActivity.After(projects[j].LastActivity)
	})

	return projects, nil
}

// loadProjectSessions 加载项目的所有会话
func (e *Engine) loadProjectSessions(projectDir string) ([]*session.Session, error) {
	claudeDir := filepath.Join(e.homeDir, ".claude", "projects", projectDir)
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("项目目录不存在: %s", projectDir)
	}

	jsonlFiles, err := filepath.Glob(filepath.Join(claudeDir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("查找会话文件失败: %w", err)
	}

	var sessions []*session.Session
	reader := claude.NewReader()

	for _, jsonlPath := range jsonlFiles {
		sess, err := reader.Read(jsonlPath)
		if err != nil {
			continue // 跳过解析失败的会话
		}
		sessions = append(sessions, sess)
	}

	return sessions, nil
}

// formatProjectDirName 将项目目录名转换为可读的名称
func formatProjectDirName(dirName string) string {
	name := strings.TrimPrefix(dirName, "-")
	parts := strings.Split(name, "-")
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) >= 2 {
		return filtered[len(filtered)-2] + "-" + filtered[len(filtered)-1]
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return name
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	Name         string    `json:"name"`         // 显示名称
	DirName      string    `json:"dirName"`      // 目录名
	SessionCount int       `json:"sessionCount"` // 会话数
	LastActivity time.Time `json:"lastActivity"` // 最后活动时间
}
