package diff

import (
	"argus-desktop/internal/session"
)

// Matcher 将 diff 文件关联到 Agent 动作
type Matcher struct{}

// NewMatcher 创建新的 Matcher
func NewMatcher() *Matcher {
	return &Matcher{}
}

// MatchResult 匹配结果
type MatchResult struct {
	Diff    DiffResult
	Actions []session.Action // 导致此文件改动的所有 Agent 动作
}

// Match 将 diff 文件与 actions 进行匹配
func (m *Matcher) Match(diffs []DiffResult, actions []session.Action) []MatchResult {
	// 建立文件 -> actions 的映射
	fileActions := make(map[string][]session.Action)
	for _, action := range actions {
		if action.FilePath != "" {
			fileActions[action.FilePath] = append(fileActions[action.FilePath], action)
		}
	}

	var results []MatchResult
	for _, diff := range diffs {
		result := MatchResult{
			Diff:    diff,
			Actions: fileActions[diff.FilePath],
		}
		results = append(results, result)
	}

	return results
}

// MatchWithGitDiff 将 session 的 actions 与 git diff 结合
func (m *Matcher) MatchWithGitDiff(sess *session.Session, diffs []DiffResult) []session.FileChange {
	// 建立文件 -> actions 的映射
	fileActions := make(map[string][]session.Action)
	for _, action := range sess.Actions {
		if action.FilePath != "" {
			fileActions[action.FilePath] = append(fileActions[action.FilePath], action)
		}
	}

	var fileChanges []session.FileChange
	for _, diff := range diffs {
		fc := session.FileChange{
			Path:       diff.FilePath,
			ChangeType: diff.ChangeType,
			Actions:    fileActions[diff.FilePath],
			Diff:       diff.Patch,
		}
		fileChanges = append(fileChanges, fc)
	}

	return fileChanges
}
