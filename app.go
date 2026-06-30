package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"agentscope-desktop/internal/diff"
	"agentscope-desktop/internal/export"
	"agentscope-desktop/internal/monitor"
	"agentscope-desktop/internal/risk"
	"agentscope-desktop/internal/session"
	"agentscope-desktop/internal/session/claude"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	monitor *monitor.Monitor
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

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s (文件可能已被移动或删除)", filepath.Base(filePath))
	}

	// 如果文件路径是绝对路径，使用文件所在目录查找 Git 仓库
	if filepath.IsAbs(filePath) {
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
	// 将文件路径转换为相对于 git 根目录的路径
	relPath, err := filepath.Rel(gitRoot, filePath)
	if err != nil {
		// 如果无法计算相对路径，使用原始路径
		relPath = filePath
	}
	patch, err := diffEngine.GetFilePatch(relPath)
	if err != nil {
		return "", fmt.Errorf("获取 diff 失败: %w", err)
	}

	return patch, nil
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
			diffContent = joinStrings(diffParts, "\n\n")
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
			reader := claude.NewReader()
			sess, err := reader.Read(jsonlPath)
			if err != nil {
				continue
			}
			// 检查会话 ID 是否匹配
			if sess.ID == id {
				return sess, nil
			}
		}
	}

	return nil, fmt.Errorf("未找到会话: %s", id)
}

// joinStrings joins a slice of strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
