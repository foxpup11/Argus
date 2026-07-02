package continuity

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"agentscope-desktop/internal/session"
)

// Validator Git 交叉验证器
type Validator struct {
	homeDir string
}

// NewValidator 创建新的验证器
func NewValidator() (*Validator, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户目录失败: %w", err)
	}
	return &Validator{homeDir: homeDir}, nil
}

// ValidateTasks 验证任务是否被 git 记录
func (v *Validator) ValidateTasks(tasks []CompletedTask, cwd string) []CompletedTask {
	if cwd == "" {
		return tasks
	}

	// 获取 git root
	gitRoot := findGitRoot(cwd)
	if gitRoot == "" {
		// 无法找到 git 仓库，所有任务标记为未验证
		return tasks
	}

	// 获取该目录的 git log（文件变更记录）
	gitFiles := v.getGitChangedFiles(gitRoot)

	// 验证每个任务
	for i := range tasks {
		verified := false
		for _, changedFile := range gitFiles {
			for _, taskFile := range tasks[i].FilesChanged {
				// 比较文件路径（支持相对路径和绝对路径匹配）
				if pathsMatch(changedFile, taskFile, gitRoot) {
					verified = true
					break
				}
			}
			if verified {
				break
			}
		}
		tasks[i].VerifiedByGit = verified
	}

	return tasks
}

// getGitChangedFiles 获取 git 中变更的文件列表
func (v *Validator) getGitChangedFiles(gitRoot string) []string {
	var files []string

	// 1. 获取已跟踪文件的变更
	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = gitRoot
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		files = append(files, lines...)
	}

	// 2. 获取已暂存的变更
	cmd = exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = gitRoot
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		files = append(files, lines...)
	}

	// 3. 获取未跟踪的新文件
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = gitRoot
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		files = append(files, lines...)
	}

	// 4. 获取最近的 git log 中的变更（最近 7 天）
	cmd = exec.Command("git", "log", "--since=7.days", "--name-only", "--pretty=format:")
	cmd.Dir = gitRoot
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				files = append(files, line)
			}
		}
	}

	// 去重
	seen := make(map[string]bool)
	var unique []string
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f != "" && !seen[f] {
			seen[f] = true
			unique = append(unique, f)
		}
	}

	return unique
}

// ValidateAgainstSessions 验证任务是否被会话记录
func (v *Validator) ValidateAgainstSessions(tasks []CompletedTask, sessions []*session.Session) []CompletedTask {
	// 构建会话中的文件变更索引
	sessionFiles := make(map[string]bool)
	for _, sess := range sessions {
		for _, action := range sess.Actions {
			if action.FilePath != "" {
				sessionFiles[action.FilePath] = true
			}
		}
	}

	for i := range tasks {
		verified := false
		for _, taskFile := range tasks[i].FilesChanged {
			if sessionFiles[taskFile] {
				verified = true
				break
			}
		}
		// 如果已经通过 git 验证，保持状态
		if !tasks[i].VerifiedByGit {
			tasks[i].VerifiedByGit = verified
		}
	}

	return tasks
}

// findGitRoot 查找 git 仓库根目录
func findGitRoot(startDir string) string {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// pathsMatch 比较两个文件路径是否匹配
func pathsMatch(gitPath, taskPath, gitRoot string) bool {
	// 直接匹配
	if gitPath == taskPath {
		return true
	}

	// 将任务路径转换为相对于 git root 的路径
	if filepath.IsAbs(taskPath) {
		relPath, err := filepath.Rel(gitRoot, taskPath)
		if err == nil && relPath == gitPath {
			return true
		}
		// 也尝试匹配文件名
		if filepath.Base(taskPath) == gitPath || filepath.Base(taskPath) == filepath.Base(gitPath) {
			return true
		}
	} else {
		// 任务路径是相对路径
		if taskPath == gitPath {
			return true
		}
		// 匹配文件名
		if filepath.Base(taskPath) == filepath.Base(gitPath) {
			return true
		}
	}

	return false
}

// GetSessionGitRoot 获取会话对应的 git 根目录
func GetSessionGitRoot(sess *session.Session) string {
	if sess.CWD == "" {
		return ""
	}
	return findGitRoot(sess.CWD)
}

// GetSessionTimeRange 获取会话的时间范围
func GetSessionTimeRange(sessions []*session.Session) (time.Time, time.Time) {
	if len(sessions) == 0 {
		return time.Time{}, time.Time{}
	}

	var earliest, latest time.Time
	for _, sess := range sessions {
		if earliest.IsZero() || sess.StartedAt.Before(earliest) {
			earliest = sess.StartedAt
		}
		endTime := sess.StartedAt.Add(sess.Duration)
		if latest.IsZero() || endTime.After(latest) {
			latest = endTime
		}
	}

	return earliest, latest
}
