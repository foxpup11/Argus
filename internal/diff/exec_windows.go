//go:build windows

package diff

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// findGitExecutable 查找 git 可执行文件路径
func findGitExecutable() string {
	// 首先尝试直接调用 git（如果在 PATH 中）
	if path, err := exec.LookPath("git"); err == nil {
		return path
	}

	// 尝试常见的 Windows 安装路径
	possiblePaths := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Git", "bin", "git.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Git", "bin", "git.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Git", "bin", "git.exe"),
		filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs", "Git", "bin", "git.exe"),
		"C:\\Program Files\\Git\\bin\\git.exe",
		"C:\\Program Files (x86)\\Git\\bin\\git.exe",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 返回默认值，让系统尝试调用
	return "git"
}

// 创建隐藏窗口的exec.Command（Windows版本）
func newGitCommand(args ...string) *exec.Cmd {
	gitPath := findGitExecutable()
	cmd := exec.Command(gitPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}
