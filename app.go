package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agentscope-desktop/internal/analytics"
	"agentscope-desktop/internal/continuity"
	"agentscope-desktop/internal/diff"
	"agentscope-desktop/internal/export"
	"agentscope-desktop/internal/knowledge"
	"agentscope-desktop/internal/monitor"
	"agentscope-desktop/internal/risk"
	"agentscope-desktop/internal/session"
	"agentscope-desktop/internal/session/claude"
	"agentscope-desktop/internal/settings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx         context.Context
	monitor     *monitor.Monitor
	settingsMgr *settings.Manager
	analytics   *analytics.Engine
	metaStore   *session.MetaStore
	knowledge   *knowledge.Engine
	continuity  *continuity.Engine
}

// SessionInfo 会话简要信息（用于列表展示）
type SessionInfo struct {
	ID          string    `json:"id"`
	Model       string    `json:"model"`
	Prompt      string    `json:"prompt"`
	Branch      string    `json:"branch"`
	StartedAt   time.Time `json:"startedAt"`
	FileCount   int       `json:"fileCount"`
	ActionCount int       `json:"actionCount"`
	ProjectDir  string    `json:"projectDir"`  // 项目目录名（用于分组）
	ProjectName string    `json:"projectName"` // 项目显示名称
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
	if err != nil {
		log.Printf("WARN: 设置管理器初始化失败: %v", err)
	} else {
		a.settingsMgr = mgr
	}

	// 初始化 Token 分析引擎
	engine, err := analytics.NewEngine()
	if err != nil {
		log.Printf("WARN: Token分析引擎初始化失败: %v", err)
	} else {
		a.analytics = engine
	}

	// 初始化会话元数据存储
	metaStore, err := session.NewMetaStore()
	if err != nil {
		log.Printf("WARN: 会话元数据存储初始化失败: %v", err)
	} else {
		a.metaStore = metaStore
	}

	// 初始化知识管理引擎
	knowledgeEngine, err := knowledge.NewEngine()
	if err != nil {
		log.Printf("WARN: 知识管理引擎初始化失败: %v", err)
	} else {
		a.knowledge = knowledgeEngine
	}

	// 初始化会话连续性引擎
	continuityEngine, err := continuity.NewEngine()
	if err != nil {
		log.Printf("WARN: 会话连续性引擎初始化失败: %v", err)
	} else {
		a.continuity = continuityEngine
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
				ProjectDir:  entry.Name(),
				ProjectName: formatProjectName(entry.Name()),
			})
		}
	}

	// 按时间倒序排列（最新的在前）
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	return sessions, nil
}

// formatProjectName 将项目目录名转换为可读的项目名称
// 例如: "-g-ltch-git-learn-agentscope-desktop" -> "agentscope-desktop"
func formatProjectName(dirName string) string {
	// 去掉开头的连字符
	name := strings.TrimPrefix(dirName, "-")

	// 过滤空字符串并取最后两个段
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

// GetAllProjectDirs 获取所有有会话的项目目录名（共享逻辑，与会话列表保持一致）
func GetAllProjectDirs() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	claudeDir := filepath.Join(homeDir, ".claude", "projects")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return nil, err
	}

	var projects []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 只返回有会话文件的项目（与 GetSessions 保持一致）
		projectDir := filepath.Join(claudeDir, entry.Name())
		jsonlFiles, _ := filepath.Glob(filepath.Join(projectDir, "*.jsonl"))
		if len(jsonlFiles) > 0 {
			projects = append(projects, entry.Name())
		}
	}

	return projects, nil
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
	// 跟踪每个文件是否有 Write action
	fileHasWrite := make(map[string]bool)
	for _, action := range sess.Actions {
		if action.FilePath == "" {
			continue
		}

		// 记录是否有 Write action
		if action.Type == session.ActionWrite {
			fileHasWrite[action.FilePath] = true
		}

		fc, exists := fileChangesMap[action.FilePath]
		if !exists {
			fc = &session.FileChange{
				Path:       action.FilePath,
				ChangeType: session.ChangeModified,
				Actions:    []session.Action{action},
			}
			fileChangesMap[action.FilePath] = fc
		} else {
			fc.Actions = append(fc.Actions, action)
		}
	}

	// 根据是否有 Write action 确定 ChangeType
	for filePath, fc := range fileChangesMap {
		if fileHasWrite[filePath] {
			fc.ChangeType = session.ChangeCreated
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

// GetSessionMessages 获取会话的完整消息历史
func (a *App) GetSessionMessages(sessionID string) ([]session.Message, error) {
	sess, err := a.getSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}
	return sess.Messages, nil
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

// ============================================
// Token Analytics API
// ============================================

// GetTokenOverview 获取 Token 使用概览数据（仪表盘首页）
func (a *App) GetTokenOverview() (*analytics.TokenOverview, error) {
	if a.analytics == nil {
		return nil, fmt.Errorf("Token 分析引擎未初始化")
	}
	return a.analytics.Refresh()
}

// GetTokenTrend 获取 Token 使用趋势（最近 N 天）
func (a *App) GetTokenTrend(days int) ([]analytics.DailyUsage, error) {
	if a.analytics == nil {
		return nil, fmt.Errorf("Token 分析引擎未初始化")
	}
	return a.analytics.GetTrend(days)
}

// GetTokenByProject 获取按项目分组的 Token 使用统计
func (a *App) GetTokenByProject() ([]analytics.ProjectStats, error) {
	if a.analytics == nil {
		return nil, fmt.Errorf("Token 分析引擎未初始化")
	}
	return a.analytics.GetProjectBreakdown()
}

// GetTokenByModel 获取按模型分组的 Token 使用统计
func (a *App) GetTokenByModel() ([]analytics.ModelStats, error) {
	if a.analytics == nil {
		return nil, fmt.Errorf("Token 分析引擎未初始化")
	}
	return a.analytics.GetModelBreakdown()
}

// ============================================
// Session Management Enhancement APIs
// ============================================

// SearchSessions 全文搜索会话
func (a *App) SearchSessions(keyword string, fields []string, tags []string, favorited *bool) ([]session.SearchResult, error) {
	if a.metaStore == nil {
		return nil, fmt.Errorf("元数据存储未初始化")
	}

	// 获取所有会话
	sessions, err := a.GetSessions()
	if err != nil {
		return nil, fmt.Errorf("获取会话列表失败: %w", err)
	}

	// 转换为可搜索的会话格式
	searchableSessions := make([]session.SearchableSession, len(sessions))
	for i, s := range sessions {
		searchableSessions[i] = session.SearchableSession{
			ID:         s.ID,
			Prompt:     s.Prompt,
			Model:      s.Model,
			Branch:     s.Branch,
			ProjectDir: s.ProjectDir,
		}
	}

	query := session.SearchQuery{
		Keyword:   keyword,
		Fields:    fields,
		Tags:      tags,
		Favorited: favorited,
	}

	return a.metaStore.Search(searchableSessions, query), nil
}

// GetSessionMeta 获取会话元数据
func (a *App) GetSessionMeta(sessionID string) (*session.SessionMeta, error) {
	if a.metaStore == nil {
		return nil, fmt.Errorf("元数据存储未初始化")
	}

	meta, ok := a.metaStore.GetMeta(sessionID)
	if !ok {
		// 返回空元数据
		return &session.SessionMeta{
			SessionID: sessionID,
			Tags:      []string{},
			AutoTags:  []string{},
			Favorited: false,
		}, nil
	}

	return meta, nil
}

// SetSessionFavorite 设置会话收藏状态
func (a *App) SetSessionFavorite(sessionID string, favorited bool) error {
	if a.metaStore == nil {
		return fmt.Errorf("元数据存储未初始化")
	}

	return a.metaStore.SetFavorite(sessionID, favorited)
}

// GetFavoriteSessions 获取所有收藏的会话 ID
func (a *App) GetFavoriteSessions() ([]string, error) {
	if a.metaStore == nil {
		return nil, fmt.Errorf("元数据存储未初始化")
	}

	return a.metaStore.GetFavorites(), nil
}

// AddSessionTag 为会话添加标签
func (a *App) AddSessionTag(sessionID, tag string) error {
	if a.metaStore == nil {
		return fmt.Errorf("元数据存储未初始化")
	}

	return a.metaStore.AddTag(sessionID, tag)
}

// RemoveSessionTag 移除会话标签
func (a *App) RemoveSessionTag(sessionID, tag string) error {
	if a.metaStore == nil {
		return fmt.Errorf("元数据存储未初始化")
	}

	return a.metaStore.RemoveTag(sessionID, tag)
}

// GetAllTags 获取所有已使用的标签
func (a *App) GetAllTags() ([]string, error) {
	if a.metaStore == nil {
		return nil, fmt.Errorf("元数据存储未初始化")
	}

	return a.metaStore.GetAllTags(), nil
}

// GetCustomTags 获取用户自定义标签列表
func (a *App) GetCustomTags() ([]session.Tag, error) {
	if a.metaStore == nil {
		return nil, fmt.Errorf("元数据存储未初始化")
	}

	return a.metaStore.GetCustomTags(), nil
}

// AddCustomTag 添加自定义标签
func (a *App) AddCustomTag(name, color, description string) error {
	if a.metaStore == nil {
		return fmt.Errorf("元数据存储未初始化")
	}

	return a.metaStore.AddCustomTag(session.Tag{
		Name:        name,
		Color:       color,
		Description: description,
	})
}

// RemoveCustomTag 删除自定义标签
func (a *App) RemoveCustomTag(name string) error {
	if a.metaStore == nil {
		return fmt.Errorf("元数据存储未初始化")
	}

	return a.metaStore.RemoveCustomTag(name)
}

// SetSessionNote 设置会话备注
func (a *App) SetSessionNote(sessionID, note string) error {
	if a.metaStore == nil {
		return fmt.Errorf("元数据存储未初始化")
	}

	return a.metaStore.SetNote(sessionID, note)
}

// GetSessionNote 获取会话备注
func (a *App) GetSessionNote(sessionID string) string {
	if a.metaStore == nil {
		return ""
	}

	return a.metaStore.GetNote(sessionID)
}

// ApplyAutoTagsToSession 为会话应用自动标签
func (a *App) ApplyAutoTagsToSession(sessionID string) ([]string, error) {
	if a.metaStore == nil {
		return nil, fmt.Errorf("元数据存储未初始化")
	}

	// 获取会话详情
	sess, err := a.getSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// 提取文件路径和命令
	var filePaths []string
	var commands []string
	for _, action := range sess.Actions {
		if action.FilePath != "" {
			filePaths = append(filePaths, action.FilePath)
		}
		if action.Type == session.ActionBash && action.Description != "" {
			commands = append(commands, action.Description)
		}
	}

	newTags := a.metaStore.ApplyAutoTags(sessionID, sess.Prompt, filePaths, commands)
	return newTags, nil
}

// BatchOperation 执行批量操作
func (a *App) BatchOperation(op session.BatchOperation) (*session.BatchOperationResult, error) {
	if a.metaStore == nil {
		return nil, fmt.Errorf("元数据存储未初始化")
	}

	result := &session.BatchOperationResult{
		Success: 0,
		Failed:  0,
		Errors:  []string{},
	}

	for _, sessionID := range op.SessionIDs {
		var err error

		switch op.Action {
		case "favorite":
			err = a.metaStore.SetFavorite(sessionID, true)
		case "unfavorite":
			err = a.metaStore.SetFavorite(sessionID, false)
		case "tag":
			if op.Tag != "" {
				err = a.metaStore.AddTag(sessionID, op.Tag)
			}
		case "untag":
			if op.Tag != "" {
				err = a.metaStore.RemoveTag(sessionID, op.Tag)
			}
		case "delete":
			// 删除会话文件
			err = a.deleteSession(sessionID)
		case "export":
			// 批量导出会话
			if op.Format == "" {
				op.Format = "markdown"
			}
			_, err = a.ExportSession(sessionID, op.Format, op.OutputDir)
		default:
			err = fmt.Errorf("未知的操作类型: %s", op.Action)
		}

		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", sessionID, err))
		} else {
			result.Success++
		}
	}

	return result, nil
}

// deleteSession 删除会话文件
func (a *App) deleteSession(sessionID string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude", "projects")

	// 查找并删除会话文件
	entries, _ := os.ReadDir(claudeDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectDir := filepath.Join(claudeDir, entry.Name())
		jsonlFiles, _ := filepath.Glob(filepath.Join(projectDir, "*.jsonl"))
		for _, jsonlPath := range jsonlFiles {
			if filepath.Base(jsonlPath) == sessionID+".jsonl" || filepath.Base(jsonlPath) == sessionID {
				return os.Remove(jsonlPath)
			}
		}
	}

	return fmt.Errorf("未找到会话文件: %s", sessionID)
}

// BatchExport 批量导出会话
func (a *App) BatchExport(sessionIDs []string, format string, outputDir string) (*session.BatchOperationResult, error) {
	op := session.BatchOperation{
		Action:    "export",
		SessionIDs: sessionIDs,
		Format:    format,
		OutputDir: outputDir,
	}
	return a.BatchOperation(op)
}

// GetSessionDetailWithMeta 获取带元数据的会话详情
func (a *App) GetSessionDetailWithMeta(sessionID string) (map[string]any, error) {
	detail, err := a.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	var meta *session.SessionMeta
	if a.metaStore != nil {
		m, ok := a.metaStore.GetMeta(sessionID)
		if ok {
			meta = m
		}
	}

	if meta == nil {
		meta = &session.SessionMeta{
			SessionID: sessionID,
			Tags:      []string{},
			AutoTags:  []string{},
			Favorited: false,
		}
	}

	return map[string]any{
		"detail": detail,
		"meta":   meta,
	}, nil
}

// ============================================
// Knowledge Management API
// ============================================

// KnowledgeDocInfo 知识文档信息（前端展示用）
type KnowledgeDocInfo struct {
	Path        string            `json:"path"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Project     string            `json:"project"`
	Content     string            `json:"content"`
	Frontmatter map[string]string `json:"frontmatter"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	Size        int64             `json:"size"`
}

// GetKnowledgeDocuments 获取知识文档列表
func (a *App) GetKnowledgeDocuments(docType string, project string) ([]KnowledgeDocInfo, error) {
	if a.knowledge == nil {
		return nil, fmt.Errorf("知识管理引擎未初始化")
	}

	docs, err := a.knowledge.GetAllDocuments(docType, project)
	if err != nil {
		return nil, err
	}

	result := make([]KnowledgeDocInfo, len(docs))
	for i, doc := range docs {
		result[i] = KnowledgeDocInfo{
			Path:        doc.Path,
			Name:        doc.Name,
			Type:        string(doc.Type),
			Project:     doc.Project,
			Content:     doc.Content,
			Frontmatter: doc.Frontmatter,
			CreatedAt:   doc.CreatedAt,
			UpdatedAt:   doc.UpdatedAt,
			Size:        doc.Size,
		}
	}

	return result, nil
}

// GetKnowledgeDocument 获取单个知识文档
func (a *App) GetKnowledgeDocument(path string) (*KnowledgeDocInfo, error) {
	if a.knowledge == nil {
		return nil, fmt.Errorf("知识管理引擎未初始化")
	}

	doc, err := a.knowledge.GetDocument(path)
	if err != nil {
		return nil, err
	}

	return &KnowledgeDocInfo{
		Path:        doc.Path,
		Name:        doc.Name,
		Type:        string(doc.Type),
		Project:     doc.Project,
		Content:     doc.Content,
		Frontmatter: doc.Frontmatter,
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
		Size:        doc.Size,
	}, nil
}

// GetKnowledgeProjects 获取所有项目列表（使用共享逻辑）
func (a *App) GetKnowledgeProjects() ([]string, error) {
	return GetAllProjectDirs()
}

// SaveKnowledgeDocument 保存知识文档
func (a *App) SaveKnowledgeDocument(path string, content string) error {
	if a.knowledge == nil {
		return fmt.Errorf("知识管理引擎未初始化")
	}

	return a.knowledge.SaveDocument(path, content)
}

// DeleteKnowledgeDocument 删除知识文档
func (a *App) DeleteKnowledgeDocument(path string) error {
	if a.knowledge == nil {
		return fmt.Errorf("知识管理引擎未初始化")
	}

	return a.knowledge.DeleteDocument(path)
}

// RenameKnowledgeDocument 重命名知识文档，返回新的文件路径
func (a *App) RenameKnowledgeDocument(path string, newName string) (string, error) {
	if a.knowledge == nil {
		return "", fmt.Errorf("知识管理引擎未初始化")
	}

	newPath, err := a.knowledge.RenameDocument(path, newName)
	if err != nil {
		return "", err
	}

	return newPath, nil
}

// CreateKnowledgeDocument 创建知识文档
func (a *App) CreateKnowledgeDocument(docType string, title string, content string, project string, sessionId string) (string, error) {
	if a.knowledge == nil {
		return "", fmt.Errorf("知识管理引擎未初始化")
	}

	// 如果标题为空，让知识引擎自动生成默认标题
	if title == "" {
		return a.knowledge.CreateDocument(knowledge.DocType(docType), title, content, project, sessionId)
	}

	// 清理标题中的特殊字符（只允许字母、数字、空格、连字符、下划线）
	var cleanTitle strings.Builder
	for _, r := range title {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == ' ' || r == '-' || r == '_' || r == '.' {
			cleanTitle.WriteRune(r)
		}
	}
	cleanedTitle := cleanTitle.String()
	if cleanedTitle == "" {
		return "", fmt.Errorf("title contains only invalid characters")
	}

	return a.knowledge.CreateDocument(knowledge.DocType(docType), cleanedTitle, content, project, sessionId)
}

// SearchKnowledgeDocuments 搜索知识文档
func (a *App) SearchKnowledgeDocuments(query string, types []string, projects []string) ([]KnowledgeDocInfo, error) {
	if a.knowledge == nil {
		return nil, fmt.Errorf("知识管理引擎未初始化")
	}

	// 转换类型
	docTypes := make([]knowledge.DocType, len(types))
	for i, t := range types {
		docTypes[i] = knowledge.DocType(t)
	}

	filters := knowledge.SearchFilters{
		Types:    docTypes,
		Projects: projects,
	}

	docs, err := a.knowledge.SearchDocuments(query, filters)
	if err != nil {
		return nil, err
	}

	result := make([]KnowledgeDocInfo, len(docs))
	for i, doc := range docs {
		result[i] = KnowledgeDocInfo{
			Path:        doc.Path,
			Name:        doc.Name,
			Type:        string(doc.Type),
			Project:     doc.Project,
			Content:     doc.Content,
			Frontmatter: doc.Frontmatter,
			CreatedAt:   doc.CreatedAt,
			UpdatedAt:   doc.UpdatedAt,
			Size:        doc.Size,
		}
	}

	return result, nil
}

// OpenFileLocation 打开文件所在位置
func (a *App) OpenFileLocation(filePath string) error {
	log.Printf("OpenFileLocation: 收到文件路径: %q", filePath)

	// 验证文件路径
	if filePath == "" {
		log.Printf("OpenFileLocation: 文件路径为空")
		return fmt.Errorf("文件路径不能为空")
	}

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		log.Printf("OpenFileLocation: 文件不存在: %v, 路径: %q", err, filePath)
		return fmt.Errorf("文件不存在: %w", err)
	}

	// 如果是文件，获取其所在目录
	dir := filePath
	if !info.IsDir() {
		dir = filepath.Dir(filePath)
	}

	log.Printf("OpenFileLocation: 打开目录: %q", dir)
	// 跨平台打开文件夹
	return openDirectory(dir)
}

// ============================================
// CLAUDE.md Editor APIs
// ============================================

// ClaudeMDTemplateInfo CLAUDE.md 模板信息
type ClaudeMDTemplateInfo struct {
	Sections []knowledge.ClaudeMDSection `json:"sections"`
}

// ClaudeMDProjectInfo CLAUDE.md 项目信息
type ClaudeMDProjectInfo struct {
	Name      string `json:"name"`
	HasCLAUDE bool   `json:"hasClaudeMD"`
	Path      string `json:"path"`
	RootDir   string `json:"rootDir"`
}

// CLAUDEProjectInfo 前端展示的项目信息
type CLAUDEProjectInfo struct {
	Name         string   `json:"name"`
	Language     string   `json:"language"`
	LanguageIcon string   `json:"languageIcon"`
	Framework    string   `json:"framework"`
	BuildTool    string   `json:"buildTool"`
	HasTests     bool     `json:"hasTests"`
	HasCI        bool     `json:"hasCI"`
	HasDocker    bool     `json:"hasDocker"`
	MainDirs     []string `json:"mainDirs"`
	ConfigFiles  []string `json:"configFiles"`
}

// GetClaudeMDTemplate 获取 CLAUDE.md 默认模板
func (a *App) GetClaudeMDTemplate(projectName string) (*ClaudeMDTemplateInfo, error) {
	content := knowledge.GetClaudeMDTemplate(projectName)
	sections := knowledge.ParseClaudeMDSections(content)
	return &ClaudeMDTemplateInfo{Sections: sections}, nil
}

// ParseClaudeMDSections 解析 CLAUDE.md 为分节
func (a *App) ParseClaudeMDSections(path string) ([]knowledge.ClaudeMDSection, error) {
	if a.knowledge == nil {
		return nil, fmt.Errorf("知识管理引擎未初始化")
	}

	doc, err := a.knowledge.GetDocument(path)
	if err != nil {
		return nil, fmt.Errorf("读取 CLAUDE.md 失败: %w", err)
	}

	sections := knowledge.ParseClaudeMDSections(doc.Content)
	return sections, nil
}

// SaveClaudeMDSections 保存分节编辑结果
func (a *App) SaveClaudeMDSections(path string, projectName string, sections []knowledge.ClaudeMDSection) error {
	if a.knowledge == nil {
		return fmt.Errorf("知识管理引擎未初始化")
	}

	content := knowledge.SerializeClaudeMDSections(projectName, sections)
	return a.knowledge.SaveDocument(path, content)
}

// GenerateClaudeMDFromProject 从项目结构自动生成 CLAUDE.md
func (a *App) GenerateClaudeMDFromProject(projectDir string) (string, error) {
	info, err := knowledge.DetectProject(projectDir)
	if err != nil {
		return "", fmt.Errorf("检测项目失败: %w", err)
	}

	content := knowledge.GenerateClaudeMDFromProject(info)
	return content, nil
}

// DetectProjectInfo 检测项目信息
func (a *App) DetectProjectInfo(projectDir string) (*CLAUDEProjectInfo, error) {
	info, err := knowledge.DetectProject(projectDir)
	if err != nil {
		return nil, err
	}

	return &CLAUDEProjectInfo{
		Name:         info.Name,
		Language:     info.Language,
		LanguageIcon: info.LanguageIcon,
		Framework:    info.Framework,
		BuildTool:    info.BuildTool,
		HasTests:     info.HasTests,
		HasCI:        info.HasCI,
		HasDocker:    info.HasDocker,
		MainDirs:     info.MainDirs,
		ConfigFiles:  info.ConfigFiles,
	}, nil
}

// GetCLAUDEMDProjects 获取所有有 CLAUDE.md 的项目列表
func (a *App) GetCLAUDEMDProjects() ([]ClaudeMDProjectInfo, error) {
	if a.knowledge == nil {
		return nil, fmt.Errorf("知识管理引擎未初始化")
	}

	projects, err := a.knowledge.GetClaudeMDProjects()
	if err != nil {
		return nil, err
	}

	result := make([]ClaudeMDProjectInfo, len(projects))
	for i, p := range projects {
		result[i] = ClaudeMDProjectInfo{
			Name:      p.Name,
			HasCLAUDE: p.HasCLAUDE,
			Path:      p.Path,
			RootDir:   p.RootDir,
		}
	}

	return result, nil
}

// CLAUDEMDBatchUpdate 批量更新项
type CLAUDEMDBatchUpdate struct {
	Project string `json:"project"`
	Content string `json:"content"`
}

// BatchCLAUDEMDResult 批量操作结果
type BatchCLAUDEMDResult struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

// BatchUpdateCLAUDEMD 批量更新多个项目的 CLAUDE.md
func (a *App) BatchUpdateCLAUDEMD(updates []CLAUDEMDBatchUpdate) (*BatchCLAUDEMDResult, error) {
	if a.knowledge == nil {
		return nil, fmt.Errorf("知识管理引擎未初始化")
	}

	result := &BatchCLAUDEMDResult{
		Success: 0,
		Failed:  0,
		Errors:  []string{},
	}

	for _, update := range updates {
		_, err := a.knowledge.CreateDocument(knowledge.DocTypeClaudeMD, "CLAUDE.md", update.Content, update.Project, "")
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", update.Project, err))
		} else {
			result.Success++
		}
	}

	return result, nil
}

// ============================================
// Session Continuity API
// ============================================

// ContinuityProjectInfo 项目信息（前端展示用）
type ContinuityProjectInfo struct {
	Name         string    `json:"name"`
	DirName      string    `json:"dirName"`
	SessionCount int       `json:"sessionCount"`
	LastActivity time.Time `json:"lastActivity"`
}

// ContinuityTaskInfo 任务信息（前端展示用）
type ContinuityTaskInfo struct {
	Description   string    `json:"description"`
	SessionID     string    `json:"sessionId"`
	FilesChanged  []string  `json:"filesChanged"`
	VerifiedByGit bool      `json:"verifiedByGit"`
	Timestamp     time.Time `json:"timestamp"`
}

// ContinuityPendingTaskInfo 待办任务信息
type ContinuityPendingTaskInfo struct {
	Description string   `json:"description"`
	Source      string   `json:"source"`
	SessionID   string   `json:"sessionId"`
	FilesHint   []string `json:"filesHint"`
}

// ContinuityDecisionInfo 决策信息
type ContinuityDecisionInfo struct {
	Description string    `json:"description"`
	Context     string    `json:"context"`
	Timestamp   time.Time `json:"timestamp"`
	SessionID   string    `json:"sessionId"`
}

// ContinuityFileInfo 文件信息
type ContinuityFileInfo struct {
	Path         string `json:"path"`
	ChangeCount  int    `json:"changeCount"`
	ActionCount  int    `json:"actionCount"`
	LastAction   string `json:"lastAction"`
	IsTestFile   bool   `json:"isTestFile"`
	IsConfigFile bool   `json:"isConfigFile"`
}

// ContinuitySummary 完整的交接摘要（前端展示用）
type ContinuitySummary struct {
	Project        string                      `json:"project"`
	SessionsUsed   int                         `json:"sessionsUsed"`
	SessionsTotal  int                         `json:"sessionsTotal"`
	Summary        string                      `json:"summary"`
	CompletedTasks []ContinuityTaskInfo        `json:"completedTasks"`
	PendingTasks   []ContinuityPendingTaskInfo `json:"pendingTasks"`
	KeyDecisions   []ContinuityDecisionInfo    `json:"keyDecisions"`
	ModifiedFiles  []ContinuityFileInfo        `json:"modifiedFiles"`
	KnownIssues    []string                    `json:"knownIssues"`
	GeneratedAt    time.Time                   `json:"generatedAt"`
	Quality        ContinuityQualityInfo       `json:"quality"`
}

// ContinuityQualityInfo 质量评分信息
type ContinuityQualityInfo struct {
	Completeness float64 `json:"completeness"`
	Accuracy     float64 `json:"accuracy"`
	Freshness    float64 `json:"freshness"`
	OverallScore float64 `json:"overallScore"`
}

// GetContinuityProjects 获取所有有会话的项目列表
func (a *App) GetContinuityProjects() ([]ContinuityProjectInfo, error) {
	if a.continuity == nil {
		return nil, fmt.Errorf("会话连续性引擎未初始化")
	}

	projects, err := a.continuity.GetAvailableProjects()
	if err != nil {
		return nil, err
	}

	result := make([]ContinuityProjectInfo, len(projects))
	for i, p := range projects {
		result[i] = ContinuityProjectInfo{
			Name:         p.Name,
			DirName:      p.DirName,
			SessionCount: p.SessionCount,
			LastActivity: p.LastActivity,
		}
	}

	return result, nil
}

// GenerateContinuityHandoff 生成会话交接摘要
func (a *App) GenerateContinuityHandoff(project string, sessionCount int) (*ContinuitySummary, error) {
	if a.continuity == nil {
		return nil, fmt.Errorf("会话连续性引擎未初始化")
	}

	summary, err := a.continuity.GenerateHandoff(project, sessionCount)
	if err != nil {
		return nil, err
	}

	// 转换为前端格式
	completedTasks := make([]ContinuityTaskInfo, len(summary.CompletedTasks))
	for i, t := range summary.CompletedTasks {
		completedTasks[i] = ContinuityTaskInfo{
			Description:   t.Description,
			SessionID:     t.SessionID,
			FilesChanged:  t.FilesChanged,
			VerifiedByGit: t.VerifiedByGit,
			Timestamp:     t.Timestamp,
		}
	}

	pendingTasks := make([]ContinuityPendingTaskInfo, len(summary.PendingTasks))
	for i, t := range summary.PendingTasks {
		pendingTasks[i] = ContinuityPendingTaskInfo{
			Description: t.Description,
			Source:      t.Source,
			SessionID:   t.SessionID,
			FilesHint:   t.FilesHint,
		}
	}

	keyDecisions := make([]ContinuityDecisionInfo, len(summary.KeyDecisions))
	for i, d := range summary.KeyDecisions {
		keyDecisions[i] = ContinuityDecisionInfo{
			Description: d.Description,
			Context:     d.Context,
			Timestamp:   d.Timestamp,
			SessionID:   d.SessionID,
		}
	}

	modifiedFiles := make([]ContinuityFileInfo, len(summary.ModifiedFiles))
	for i, f := range summary.ModifiedFiles {
		modifiedFiles[i] = ContinuityFileInfo{
			Path:         f.Path,
			ChangeCount:  f.ChangeCount,
			ActionCount:  f.ActionCount,
			LastAction:   f.LastAction,
			IsTestFile:   f.IsTestFile,
			IsConfigFile: f.IsConfigFile,
		}
	}

	return &ContinuitySummary{
		Project:        summary.Project,
		SessionsUsed:   summary.SessionsUsed,
		SessionsTotal:  summary.SessionsTotal,
		Summary:        summary.Summary,
		CompletedTasks: completedTasks,
		PendingTasks:   pendingTasks,
		KeyDecisions:   keyDecisions,
		ModifiedFiles:  modifiedFiles,
		KnownIssues:    summary.KnownIssues,
		GeneratedAt:    summary.GeneratedAt,
		Quality: ContinuityQualityInfo{
			Completeness: summary.Quality.Completeness,
			Accuracy:     summary.Quality.Accuracy,
			Freshness:    summary.Quality.Freshness,
			OverallScore: summary.Quality.OverallScore,
		},
	}, nil
}

// ExportContinuityToMemory 导出交接摘要到 memory 目录
func (a *App) ExportContinuityToMemory(project string, sessionCount int) (string, error) {
	if a.continuity == nil {
		return "", fmt.Errorf("会话连续性引擎未初始化")
	}

	return a.continuity.ExportToMemory(project, sessionCount)
}

// GenerateContinuityMarkdown 生成 Markdown 格式的交接摘要
func (a *App) GenerateContinuityMarkdown(project string, sessionCount int) (string, error) {
	if a.continuity == nil {
		return "", fmt.Errorf("会话连续性引擎未初始化")
	}

	markdown, _, err := a.continuity.GenerateHandoffMarkdown(project, sessionCount)
	return markdown, err
}

// GenerateContinuityPrompt 生成可粘贴的 prompt 片段
func (a *App) GenerateContinuityPrompt(project string, sessionCount int) (string, error) {
	if a.continuity == nil {
		return "", fmt.Errorf("会话连续性引擎未初始化")
	}

	prompt, _, err := a.continuity.GenerateHandoffPrompt(project, sessionCount)
	return prompt, err
}

// openDirectory 打开指定目录（跨平台）
func openDirectory(dir string) error {
	log.Printf("openDirectory: 打开目录: %q", dir)
	// 使用 cmd /c start 打开文件夹，比直接调用 explorer.exe 更可靠
	cmd := exec.Command("cmd", "/c", "start", "", dir)
	err := cmd.Start()
	if err != nil {
		log.Printf("openDirectory: 打开目录失败: %v", err)
		return fmt.Errorf("打开目录失败: %w", err)
	}
	return nil
}
