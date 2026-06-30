package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agentscope-desktop/internal/diff"
	"agentscope-desktop/internal/export"
	"agentscope-desktop/internal/monitor"
	"agentscope-desktop/internal/risk"
	"agentscope-desktop/internal/session"
	"agentscope-desktop/internal/session/claude"
	"agentscope-desktop/internal/settings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	monitor      *monitor.Monitor
	settingsMgr  *settings.Manager
}

// SessionInfo 会话简要信息（用于列表展示）
type SessionInfo struct {
	ID         string    `json:"id"`
	Model      string    `json:"model"`
	Prompt     string    `json:"prompt"`
	Branch     string    `json:"branch"`
	StartedAt  time.Time `json:"startedAt"`
	FileCount  int       `json:"fileCount"`
	ActionCount int      `json:"actionCount"`
}

// FileChangeInfo 文件改动信息（用于表格展示）
type FileChangeInfo struct {
	Path        string `json:"path"`
	ChangeType  string `json:"changeType"`
	Risk        string `json:"risk"`
	RiskReason  string `json:"riskReason"`
	ActionCount int    `json:"actionCount"`
}

// SessionDetail 会话详情
type SessionDetail struct {
	ID          string           `json:"id"`
	Model       string           `json:"model"`
	Prompt      string           `json:"prompt"`
	Branch      string           `json:"branch"`
	StartedAt   time.Time        `json:"startedAt"`
	Duration    time.Duration    `json:"duration"`
	FileChanges []FileChangeInfo `json:"fileChanges"`
	TokenUsage  TokenUsageInfo   `json:"tokenUsage"`
}

// TokenUsageInfo Token 使用信息
type TokenUsageInfo struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}

// DiffMode 对比模式
type DiffMode string

const (
	// DiffModeUncommitted 未提交的改动
	DiffModeUncommitted DiffMode = "uncommitted"
	// DiffModeSession 会话前后对比
	DiffModeSession DiffMode = "session"
)

// DiffInfo diff 详细信息
type DiffInfo struct {
	Mode  DiffMode       `json:"mode"`
	Diffs []DiffFileInfo `json:"diffs"`
}

// DiffFileInfo 单个文件的 diff 信息
type DiffFileInfo struct {
	FilePath     string `json:"filePath"`
	Patch        string `json:"patch"`
	ChangeType   string `json:"changeType"`
	AddedLines   int    `json:"addedLines"`
	RemovedLines int    `json:"removedLines"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化设置管理器
	mgr, err := settings.NewManager()
	if err == nil {
		a.settingsMgr = mgr
	}
}

// GetSessions 获取所有会话列表
func (a *App) GetSessions() ([]SessionInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户目录失败: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude", "projects")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return []SessionInfo{}, nil
	}

	var sessions []SessionInfo

	// 遍历所有项目目录
	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return nil, fmt.Errorf("读取 Claude 项目目录失败: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectDir := filepath.Join(claudeDir, entry.Name())
		jsonlFiles, _ := filepath.Glob(filepath.Join(projectDir, "*.jsonl"))

		for _, jsonlPath := range jsonlFiles {
			reader := claude.NewReader()
			sess, err := reader.Read(jsonlPath)
			if err != nil {
				continue
			}

			sessions = append(sessions, SessionInfo{
				ID:          sess.ID,
				Model:       sess.Model,
				Prompt:      sess.Prompt,
				Branch:      sess.GitBranch,
				StartedAt:   sess.StartedAt,
				FileCount:   len(sess.FileChanges),
				ActionCount: len(sess.Actions),
			})
		}
	}

	// 按时间倒序排列（最新的在前）
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	return sessions, nil
}

// GetSession 获取单个会话详情
func (a *App) GetSession(id string) (*SessionDetail, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户目录失败: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude", "projects")

	// 查找会话文件
	var sessionPath string
	entries, _ := os.ReadDir(claudeDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectDir := filepath.Join(claudeDir, entry.Name())
		jsonlFiles, _ := filepath.Glob(filepath.Join(projectDir, "*.jsonl"))
		for _, jsonlPath := range jsonlFiles {
			if filepath.Base(jsonlPath) == id+".jsonl" || filepath.Base(jsonlPath) == id {
				sessionPath = jsonlPath
				break
			}
		}
		if sessionPath != "" {
			break
		}
	}

	if sessionPath == "" {
		return nil, fmt.Errorf("未找到会话: %s", id)
	}

	// 读取会话
	reader := claude.NewReader()
	sess, err := reader.Read(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("读取会话失败: %w", err)
	}

	// 从 actions 中提取文件改动
	fileChangesMap := make(map[string]*session.FileChange)
	for _, action := range sess.Actions {
		if action.FilePath == "" {
			continue
		}

		fc, exists := fileChangesMap[action.FilePath]
		if !exists {
			// 确定变更类型
			changeType := session.ChangeModified
			// 简单判断：如果 action 是 Write 且是第一个，可能是新建
			if action.Type == session.ActionWrite {
				changeType = session.ChangeCreated
			}

			fc = &session.FileChange{
				Path:       action.FilePath,
				ChangeType: changeType,
				Actions:    []session.Action{action},
			}
			fileChangesMap[action.FilePath] = fc
		} else {
			fc.Actions = append(fc.Actions, action)
		}
	}

	// 尝试获取 Git Diff（如果会话目录存在）
	workDir := sess.CWD
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	gitRoot, err := diff.FindGitRoot(workDir)
	if err == nil {
		diffEngine := diff.NewEngine(gitRoot)
		diffs, err := diffEngine.GetUncommittedDiff()
		if err == nil {
			// 将 git diff 合并到 fileChangesMap
			for _, d := range diffs {
				fc, exists := fileChangesMap[d.FilePath]
				if exists {
					fc.Diff = d.Patch
					fc.ChangeType = d.ChangeType
				} else {
					fileChangesMap[d.FilePath] = &session.FileChange{
						Path:       d.FilePath,
						ChangeType: d.ChangeType,
						Diff:       d.Patch,
						Actions:    []session.Action{},
					}
				}
			}
		}
	}

	// 转换为切片
	fileChanges := make([]session.FileChange, 0, len(fileChangesMap))
	for _, fc := range fileChangesMap {
		// 检查文件是否实际存在，如果不存在则标记为删除
		fullPath := fc.Path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(workDir, fc.Path)
		}
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fc.ChangeType = session.ChangeDeleted
		}
		fileChanges = append(fileChanges, *fc)
	}

	// 风险评估
	riskEngine := risk.NewEngine()
	fileChanges = riskEngine.EvaluateAll(fileChanges)

	// 转换为前端格式
	fileChangesInfo := make([]FileChangeInfo, 0, len(fileChanges))
	for _, fc := range fileChanges {
		// 计算操作次数
		actionCount := len(fc.Actions)
		fc.ActionCount = actionCount

		fileChangesInfo = append(fileChangesInfo, FileChangeInfo{
			Path:        fc.Path,
			ChangeType:  string(fc.ChangeType),
			Risk:        string(fc.Risk),
			RiskReason:  fc.RiskReason,
			ActionCount: actionCount,
		})
	}

	return &SessionDetail{
		ID:          sess.ID,
		Model:       sess.Model,
		Prompt:      sess.Prompt,
		Branch:      sess.GitBranch,
		StartedAt:   sess.StartedAt,
		Duration:    sess.Duration,
		FileChanges: fileChangesInfo,
		TokenUsage: TokenUsageInfo{
			InputTokens:  sess.TokenUsage.InputTokens,
			OutputTokens: sess.TokenUsage.OutputTokens,
		},
	}, nil
}

// GetDiff 获取指定文件的 diff
func (a *App) GetDiff(sessionID, filePath string) (string, error) {
	// 如果文件路径是绝对路径，使用文件所在目录查找 Git 仓库
	if filepath.IsAbs(filePath) {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return "", fmt.Errorf("文件不存在: %s (文件可能已被移动或删除)", filepath.Base(filePath))
		}
		dir := filepath.Dir(filePath)
		gitRoot, err := diff.FindGitRoot(dir)
		if err == nil {
			diffEngine := diff.NewEngine(gitRoot)
			// 使用相对路径
			relPath, _ := filepath.Rel(gitRoot, filePath)
			patch, err := diffEngine.GetFilePatch(relPath)
			if err == nil {
				return patch, nil
			}
		}
	}

	// 回退：从会话中获取工作目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude", "projects")

	// 查找会话文件
	var sessionPath string
	entries, _ := os.ReadDir(claudeDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectDir := filepath.Join(claudeDir, entry.Name())
		jsonlFiles, _ := filepath.Glob(filepath.Join(projectDir, "*.jsonl"))
		for _, jsonlPath := range jsonlFiles {
			if filepath.Base(jsonlPath) == sessionID+".jsonl" || filepath.Base(jsonlPath) == sessionID {
				sessionPath = jsonlPath
				break
			}
		}
		if sessionPath != "" {
			break
		}
	}

	if sessionPath == "" {
		return "", fmt.Errorf("未找到会话: %s", sessionID)
	}

	// 读取会话
	reader := claude.NewReader()
	sess, err := reader.Read(sessionPath)
	if err != nil {
		return "", fmt.Errorf("读取会话失败: %w", err)
	}

	// 获取文件 diff
	workDir := sess.CWD
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	gitRoot, err := diff.FindGitRoot(workDir)
	if err != nil {
		return "", fmt.Errorf("未找到 Git 仓库: %s (文件可能不在 Git 仓库中)", workDir)
	}

	diffEngine := diff.NewEngine(gitRoot)

	// 根据路径类型处理
	var relPath string
	if filepath.IsAbs(filePath) {
		// 绝对路径：转换为相对于 git 根目录的路径
		relPath, err = filepath.Rel(gitRoot, filePath)
		if err != nil {
			relPath = filepath.Base(filePath)
		}
	} else {
		// 相对路径：直接使用（已经是相对于工作目录的路径）
		relPath = filePath
	}

	patch, err := diffEngine.GetFilePatch(relPath)
	if err != nil {
		return "", fmt.Errorf("获取 diff 失败 (文件: %s): %w", relPath, err)
	}

	return patch, nil
}

// GetSessionDiff 获取会话的 diff（支持多种对比模式）
// mode: "uncommitted" 获取未提交的改动, "session" 获取会话前后对比
func (a *App) GetSessionDiff(sessionID string, mode DiffMode) (*DiffInfo, error) {
	// 获取会话数据
	sess, err := a.getSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// 确定工作目录
	workDir := sess.CWD
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	// 查找 Git 仓库
	gitRoot, err := diff.FindGitRoot(workDir)
	if err != nil {
		return &DiffInfo{
			Mode:  mode,
			Diffs: []DiffFileInfo{},
		}, nil
	}

	diffEngine := diff.NewEngine(gitRoot)
	var diffs []diff.DiffResult

	// 根据模式获取 diff
	switch mode {
	case DiffModeSession:
		// 会话前后对比
		diffs, err = diffEngine.GetDiffBetweenSession(sess)
	default:
		// 默认：未提交的改动
		diffs, err = diffEngine.GetUncommittedDiff()
	}

	if err != nil {
		return nil, fmt.Errorf("获取 diff 失败: %w", err)
	}

	// 为每个 diff 获取完整 patch
	diffInfos := make([]DiffFileInfo, 0, len(diffs))
	for _, d := range diffs {
		patch := d.Patch
		// 如果没有 patch，尝试获取
		if patch == "" {
			// 确保文件路径是相对于 git 根目录的
			relPath, err := filepath.Rel(gitRoot, d.FilePath)
			if err != nil {
				relPath = d.FilePath
			}
			patch, _ = diffEngine.GetFilePatch(relPath)
		}

		diffInfos = append(diffInfos, DiffFileInfo{
			FilePath:     d.FilePath,
			Patch:        patch,
			ChangeType:   string(d.ChangeType),
			AddedLines:   d.AddedLines,
			RemovedLines: d.RemovedLines,
		})
	}

	return &DiffInfo{
		Mode:  mode,
		Diffs: diffInfos,
	}, nil
}

// SettingsInfo 设置信息
type SettingsInfo struct {
	Theme       string             `json:"theme"`
	CustomRules []CustomRuleInfo   `json:"customRules"`
}

// CustomRuleInfo 自定义规则信息
type CustomRuleInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       string `json:"level"`
	Pattern     string `json:"pattern"`
	Enabled     bool   `json:"enabled"`
}

// GetSettings 获取应用设置
func (a *App) GetSettings() (*SettingsInfo, error) {
	if a.settingsMgr == nil {
		return &SettingsInfo{
			Theme:       "auto",
			CustomRules: []CustomRuleInfo{},
		}, nil
	}

	s := a.settingsMgr.Get()
	rules := make([]CustomRuleInfo, len(s.CustomRules))
	for i, r := range s.CustomRules {
		rules[i] = CustomRuleInfo{
			Name:        r.Name,
			Description: r.Description,
			Level:       string(r.Level),
			Pattern:     r.Pattern,
			Enabled:     r.Enabled,
		}
	}

	return &SettingsInfo{
		Theme:       string(s.Theme),
		CustomRules: rules,
	}, nil
}

// UpdateTheme 更新主题设置
func (a *App) UpdateTheme(theme string) error {
	if a.settingsMgr == nil {
		return fmt.Errorf("设置管理器未初始化")
	}
	return a.settingsMgr.UpdateTheme(settings.Theme(theme))
}

// AddCustomRule 添加自定义规则
func (a *App) AddCustomRule(name, description, level, pattern string) error {
	if a.settingsMgr == nil {
		return fmt.Errorf("设置管理器未初始化")
	}
	return a.settingsMgr.AddCustomRule(settings.CustomRule{
		Name:        name,
		Description: description,
		Level:       session.RiskLevel(level),
		Pattern:     pattern,
		Enabled:     true,
	})
}

// RemoveCustomRule 删除自定义规则
func (a *App) RemoveCustomRule(name string) error {
	if a.settingsMgr == nil {
		return fmt.Errorf("设置管理器未初始化")
	}
	return a.settingsMgr.RemoveCustomRule(name)
}

// UpdateCustomRule 更新自定义规则
func (a *App) UpdateCustomRule(name, description, level, pattern string, enabled bool) error {
	if a.settingsMgr == nil {
		return fmt.Errorf("设置管理器未初始化")
	}
	return a.settingsMgr.UpdateCustomRule(name, settings.CustomRule{
		Name:        name,
		Description: description,
		Level:       session.RiskLevel(level),
		Pattern:     pattern,
		Enabled:     enabled,
	})
}

// StartMonitoring starts watching the Claude sessions directory for changes.
// Returns true if monitoring started successfully, false if already running.
func (a *App) StartMonitoring() (bool, error) {
	if a.monitor != nil && a.monitor.IsRunning() {
		return false, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("获取用户目录失败: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude", "projects")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return false, fmt.Errorf("Claude 项目目录不存在: %s", claudeDir)
	}

	// Create callback that emits event to frontend
	callback := func() {
		// Emit event to frontend to refresh session list
		runtime.EventsEmit(a.ctx, "session-updated", nil)
	}

	m, err := monitor.New(claudeDir, callback)
	if err != nil {
		return false, fmt.Errorf("创建监控器失败: %w", err)
	}

	if err := m.Start(a.ctx); err != nil {
		return false, fmt.Errorf("启动监控器失败: %w", err)
	}

	a.monitor = m
	return true, nil
}

// StopMonitoring stops the file system monitor.
func (a *App) StopMonitoring() {
	if a.monitor != nil {
		a.monitor.Stop()
		a.monitor = nil
	}
}

// IsMonitoring returns whether the monitor is currently active.
func (a *App) IsMonitoring() bool {
	if a.monitor == nil {
		return false
	}
	return a.monitor.IsRunning()
}

// SelectDirectory opens a directory selection dialog and returns the selected path.
func (a *App) SelectDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择导出目录",
	})
	if err != nil {
		return "", fmt.Errorf("打开目录选择对话框失败: %w", err)
	}

	// 用户取消选择
	if dir == "" {
		return "", nil
	}

	return dir, nil
}

// ExportResult represents the result of a session export operation.
type ExportResult struct {
	FilePath string `json:"filePath"`
	Format   string `json:"format"`
	FileSize int64  `json:"fileSize"`
}

// ExportSession exports the session to an HTML or Markdown file.
// Returns the file path of the exported report.
// format: "html" or "markdown"
// outputDir: optional custom output directory
func (a *App) ExportSession(sessionID string, format string, outputDir string) (*ExportResult, error) {
	// Get session data
	sess, err := a.getSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// Get diff data
	var diffContent string
	workDir := sess.CWD
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	gitRoot, err := diff.FindGitRoot(workDir)
	if err == nil {
		diffEngine := diff.NewEngine(gitRoot)
		diffs, err := diffEngine.GetUncommittedDiff()
		if err == nil && len(diffs) > 0 {
			// Combine all diffs
			var diffParts []string
			for _, d := range diffs {
				if d.Patch != "" {
					diffParts = append(diffParts, d.Patch)
				}
			}
			diffContent = strings.Join(diffParts, "\n\n")
		}
	}

	// Determine export format
	var exportFormat export.ExportFormat
	switch format {
	case "markdown", "md":
		exportFormat = export.FormatMarkdown
	case "html":
		exportFormat = export.FormatHTML
	default:
		exportFormat = export.FormatHTML
	}

	// Export session with optional custom path
	result, err := export.ExportSession(sess, diffContent, export.ExportOptions{
		Format:    exportFormat,
		SessionID: sessionID,
		OutputDir: outputDir,
	})
	if err != nil {
		return nil, fmt.Errorf("导出会话失败: %w", err)
	}

	return &ExportResult{
		FilePath: result.FilePath,
		Format:   string(result.Format),
		FileSize: result.FileSize,
	}, nil
}

// getSessionByID retrieves session data by ID.
func (a *App) getSessionByID(id string) (*session.Session, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户目录失败: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude", "projects")

	// 遍历所有项目目录，查找匹配的会话 ID
	entries, _ := os.ReadDir(claudeDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectDir := filepath.Join(claudeDir, entry.Name())
		jsonlFiles, _ := filepath.Glob(filepath.Join(projectDir, "*.jsonl"))
		for _, jsonlPath := range jsonlFiles {
			// 优化：先通过文件名快速匹配，避免读取所有文件
			baseName := filepath.Base(jsonlPath)
			if baseName != id+".jsonl" && baseName != id {
				continue
			}

			// 文件名匹配后才读取文件内容
			reader := claude.NewReader()
			sess, err := reader.Read(jsonlPath)
			if err != nil {
				continue
			}
			// 检查会话 ID 是否匹配（双重验证）
			if sess.ID == id {
				return sess, nil
			}
		}
	}

	return nil, fmt.Errorf("未找到会话: %s", id)
}
