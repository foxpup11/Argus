package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"agentscope-desktop/internal/diff"
	"agentscope-desktop/internal/risk"
	"agentscope-desktop/internal/session"
	"agentscope-desktop/internal/session/claude"
)

// App struct
type App struct {
	ctx context.Context
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
	Path       string `json:"path"`
	ChangeType string `json:"changeType"`
	Risk       string `json:"risk"`
	RiskReason string `json:"riskReason"`
	ActionCount int   `actionCount`
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
	riskEngine.EvaluateAll(fileChanges)

	// 转换为前端格式
	fileChangesInfo := make([]FileChangeInfo, 0, len(fileChanges))
	for _, fc := range fileChanges {
		fileChangesInfo = append(fileChangesInfo, FileChangeInfo{
			Path:        fc.Path,
			ChangeType:  string(fc.ChangeType),
			Risk:        string(fc.Risk),
			RiskReason:  fc.RiskReason,
			ActionCount: len(fc.Actions),
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
		return "", fmt.Errorf("未找到 Git 仓库: %w", err)
	}

	diffEngine := diff.NewEngine(gitRoot)
	patch, err := diffEngine.GetFilePatch(filePath)
	if err != nil {
		return "", fmt.Errorf("获取 diff 失败: %w", err)
	}

	return patch, nil
}
