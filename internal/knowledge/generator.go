package knowledge

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ProjectInfo 项目检测信息
type ProjectInfo struct {
	Name          string   `json:"name"`          // 项目名
	RootDir       string   `json:"rootDir"`       // 项目根目录
	Language      string   `json:"language"`      // 主语言
	LanguageIcon  string   `json:"languageIcon"`  // 语言图标
	Framework     string   `json:"framework"`     // 框架
	BuildTool     string   `json:"buildTool"`     // 构建工具
	PackageMgr    string   `json:"packageMgr"`    // 包管理器
	HasTests      bool     `json:"hasTests"`      // 是否有测试
	HasCI         bool     `json:"hasCI"`         // 是否有 CI 配置
	HasDocker     bool     `json:"hasDocker"`     // 是否有 Docker
	MainDirs      []string `json:"mainDirs"`      // 主要目录结构
	ConfigFiles   []string `json:"configFiles"`   // 配置文件列表
	RecentCommits []string `json:"recentCommits"` // 最近 git commit
}

// DetectProject 扫描项目目录，检测项目信息
func DetectProject(rootDir string) (*ProjectInfo, error) {
	info := &ProjectInfo{
		RootDir: rootDir,
	}

	// 获取项目名
	info.Name = filepath.Base(rootDir)

	// 检测语言和框架
	info.detectLanguage(rootDir)
	info.detectFramework(rootDir)
	info.detectBuildTool(rootDir)
	info.detectPackageManager(rootDir)

	// 检测特性
	info.HasTests = detectTests(rootDir)
	info.HasCI = detectCI(rootDir)
	info.HasDocker = detectDocker(rootDir)

	// 扫描主要目录
	info.MainDirs = detectMainDirs(rootDir)

	// 扫描配置文件
	info.ConfigFiles = detectConfigFiles(rootDir)

	// 获取最近的 git commit
	info.RecentCommits = detectRecentCommits(rootDir)

	return info, nil
}

// detectLanguage 检测项目主要语言
func (p *ProjectInfo) detectLanguage(rootDir string) {
	// Go 项目
	if _, err := os.Stat(filepath.Join(rootDir, "go.mod")); err == nil {
		p.Language = "Go"
		p.LanguageIcon = "🔵"
		return
	}

	// Node.js/TypeScript 项目
	if _, err := os.Stat(filepath.Join(rootDir, "package.json")); err == nil {
		// 检查是否有 TypeScript 配置
		if _, err := os.Stat(filepath.Join(rootDir, "tsconfig.json")); err == nil {
			p.Language = "TypeScript"
			p.LanguageIcon = "🔷"
		} else {
			p.Language = "JavaScript"
			p.LanguageIcon = "🟡"
		}
		return
	}

	// Python 项目
	pythonFiles := []string{"requirements.txt", "pyproject.toml", "setup.py", "Pipfile"}
	for _, f := range pythonFiles {
		if _, err := os.Stat(filepath.Join(rootDir, f)); err == nil {
			p.Language = "Python"
			p.LanguageIcon = "🐍"
			return
		}
	}

	// Rust 项目
	if _, err := os.Stat(filepath.Join(rootDir, "Cargo.toml")); err == nil {
		p.Language = "Rust"
		p.LanguageIcon = "🦀"
		return
	}

	// Java/Kotlin 项目
	if _, err := os.Stat(filepath.Join(rootDir, "pom.xml")); err == nil {
		p.Language = "Java"
		p.LanguageIcon = "☕"
		return
	}
	if _, err := os.Stat(filepath.Join(rootDir, "build.gradle.kts")); err == nil {
		p.Language = "Kotlin"
		p.LanguageIcon = "🟣"
		return
	}

	// 默认
	p.Language = "Unknown"
	p.LanguageIcon = "📄"
}

// detectFramework 检测框架
func (p *ProjectInfo) detectFramework(rootDir string) {
	// Wails
	if _, err := os.Stat(filepath.Join(rootDir, "wails.json")); err == nil {
		p.Framework = "Wails v2"
		return
	}

	// React
	pkgPath := filepath.Join(rootDir, "package.json")
	if _, err := os.Stat(pkgPath); err == nil {
		content, err := os.ReadFile(pkgPath)
		if err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "react") {
				p.Framework = "React"
				return
			}
			if strings.Contains(contentStr, "vue") {
				p.Framework = "Vue"
				return
			}
			if strings.Contains(contentStr, "next") {
				p.Framework = "Next.js"
				return
			}
			if strings.Contains(contentStr, "svelte") {
				p.Framework = "Svelte"
				return
			}
		}
	}

	// Go 框架
	goModPath := filepath.Join(rootDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		content, err := os.ReadFile(goModPath)
		if err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "gin-gonic") {
				p.Framework = "Gin"
				return
			}
			if strings.Contains(contentStr, "gofiber") {
				p.Framework = "Fiber"
				return
			}
			if strings.Contains(contentStr, "echo") {
				p.Framework = "Echo"
				return
			}
		}
	}
}

// detectBuildTool 检测构建工具
func (p *ProjectInfo) detectBuildTool(rootDir string) {
	if _, err := os.Stat(filepath.Join(rootDir, "Makefile")); err == nil {
		p.BuildTool = "Make"
		return
	}
	if _, err := os.Stat(filepath.Join(rootDir, "wails.json")); err == nil {
		p.BuildTool = "Wails CLI"
		return
	}
	if _, err := os.Stat(filepath.Join(rootDir, "package.json")); err == nil {
		p.BuildTool = "npm/yarn"
		return
	}
	if _, err := os.Stat(filepath.Join(rootDir, "go.mod")); err == nil {
		p.BuildTool = "Go"
		return
	}
}

// detectPackageManager 检测包管理器
func (p *ProjectInfo) detectPackageManager(rootDir string) {
	if _, err := os.Stat(filepath.Join(rootDir, "pnpm-lock.yaml")); err == nil {
		p.PackageMgr = "pnpm"
		return
	}
	if _, err := os.Stat(filepath.Join(rootDir, "yarn.lock")); err == nil {
		p.PackageMgr = "yarn"
		return
	}
	if _, err := os.Stat(filepath.Join(rootDir, "package-lock.json")); err == nil {
		p.PackageMgr = "npm"
		return
	}
	if _, err := os.Stat(filepath.Join(rootDir, "Pipfile.lock")); err == nil {
		p.PackageMgr = "pipenv"
		return
	}
	if _, err := os.Stat(filepath.Join(rootDir, "go.sum")); err == nil {
		p.PackageMgr = "Go modules"
		return
	}
}

// detectTests 检测是否有测试
func detectTests(rootDir string) bool {
	// Go 测试 - 使用 filepath.Walk 实现真正的递归匹配
	goTestFound := false
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".go" && strings.HasSuffix(filepath.Base(path), "_test.go") {
			goTestFound = true
			return filepath.SkipDir // 找到一个就够了
		}
		return nil
	})
	if goTestFound {
		return true
	}

	// JavaScript/TypeScript 测试 - 使用 filepath.Walk 实现真正的递归匹配
	jsTestPatterns := []string{".test.js", ".test.ts", ".spec.js", ".spec.ts"}
	for _, pattern := range jsTestPatterns {
		found := false
		filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if strings.HasSuffix(filepath.Base(path), pattern) {
				found = true
				return filepath.SkipDir
			}
			return nil
		})
		if found {
			return true
		}
	}
	// Python 测试文件
	pythonFound := false
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py") {
			pythonFound = true
			return filepath.SkipDir
		}
		return nil
	})
	if pythonFound {
		return true
	}

	// 测试目录
	testDirs := []string{"test", "tests", "__tests__", "testdir"}
	for _, dir := range testDirs {
		if _, err := os.Stat(filepath.Join(rootDir, dir)); err == nil {
			return true
		}
	}

	return false
}

// detectCI 检测是否有 CI 配置
func detectCI(rootDir string) bool {
	ciPaths := []string{
		".github/workflows",
		".gitlab-ci.yml",
		".circleci",
		".travis.yml",
		"Jenkinsfile",
	}
	for _, p := range ciPaths {
		if _, err := os.Stat(filepath.Join(rootDir, p)); err == nil {
			return true
		}
	}
	return false
}

// detectDocker 检测是否有 Docker
func detectDocker(rootDir string) bool {
	dockerFiles := []string{"Dockerfile", "docker-compose.yml", "docker-compose.yaml", ".dockerignore"}
	for _, f := range dockerFiles {
		if _, err := os.Stat(filepath.Join(rootDir, f)); err == nil {
			return true
		}
	}
	return false
}

// detectMainDirs 检测主要目录
func detectMainDirs(rootDir string) []string {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil
	}

	// 优先目录
	priorityDirs := map[string]bool{
		"src": true, "cmd": true, "internal": true, "pkg": true,
		"app": true, "lib": true, "pages": true, "components": true,
		"api": true, "models": true, "services": true, "handlers": true,
		"utils": true, "helpers": true, "config": true, "configs": true,
		"docs": true, "test": true, "tests": true, "scripts": true,
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			if priorityDirs[entry.Name()] || len(dirs) < 8 {
				dirs = append(dirs, entry.Name())
			}
		}
	}

	return dirs
}

// detectConfigFiles 检测配置文件
func detectConfigFiles(rootDir string) []string {
	configFiles := []string{
		"go.mod", "go.sum", "package.json", "tsconfig.json",
		"Makefile", "Dockerfile", "docker-compose.yml",
		".gitignore", ".env.example", "wails.json",
		"pyproject.toml", "requirements.txt", "Cargo.toml",
	}

	var found []string
	for _, f := range configFiles {
		if _, err := os.Stat(filepath.Join(rootDir, f)); err == nil {
			found = append(found, f)
		}
	}

	return found
}

// detectRecentCommits 获取最近的 git commit
func detectRecentCommits(rootDir string) []string {
	gitDir := filepath.Join(rootDir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return nil
	}

	logFile := filepath.Join(gitDir, "logs", "HEAD")
	if _, err := os.Stat(logFile); err != nil {
		return nil
	}

	// 读取 git log 文件的最后几行
	file, err := os.Open(logFile)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > 50 {
			lines = lines[1:] // 只保留最后 50 行
		}
	}

	// 提取 commit message（git log 格式的最后一部分）
	var commits []string
	for i := len(lines) - 1; i >= 0 && len(commits) < 5; i-- {
		parts := strings.SplitN(lines[i], "	", 3)
		if len(parts) >= 3 {
			msg := parts[2]
			if idx := strings.Index(msg, " ("); idx > 0 {
				msg = msg[:idx]
			}
			commits = append(commits, msg)
		}
	}

	return commits
}

// GenerateClaudeMDFromProject 根据项目信息生成 CLAUDE.md
func GenerateClaudeMDFromProject(info *ProjectInfo) string {
	var sb strings.Builder

	// 标题
	sb.WriteString("# " + info.Name + "\n\n")

	// Overview
	sb.WriteString("## Overview\n\n")
	sb.WriteString(generateOverview(info) + "\n\n")

	// Tech Stack
	sb.WriteString("## Tech Stack\n\n")
	sb.WriteString(generateTechStack(info) + "\n\n")

	// Conventions
	sb.WriteString("## Conventions\n\n")
	sb.WriteString(generateConventions(info) + "\n\n")

	// Architecture
	sb.WriteString("## Architecture\n\n")
	sb.WriteString(generateArchitecture(info) + "\n\n")

	// Commands
	sb.WriteString("## Commands\n\n")
	sb.WriteString(generateCommands(info) + "\n")

	return sb.String()
}

// generateOverview 生成概述
func generateOverview(info *ProjectInfo) string {
	parts := []string{}

	if info.Language != "" && info.Language != "Unknown" {
		parts = append(parts, "A "+info.Language+" project")
	}
	if info.Framework != "" {
		parts = append(parts, "built with "+info.Framework)
	}

	if len(parts) == 0 {
		return "[项目概述：这个项目是做什么的，解决什么问题]"
	}

	return strings.Join(parts, " ") + "."
}

// generateTechStack 生成技术栈
func generateTechStack(info *ProjectInfo) string {
	var lines []string

	if info.Language != "" && info.Language != "Unknown" {
		lines = append(lines, "- **Language**: "+info.Language)
	}
	if info.Framework != "" {
		lines = append(lines, "- **Framework**: "+info.Framework)
	}
	if info.BuildTool != "" {
		lines = append(lines, "- **Build**: "+info.BuildTool)
	}
	if info.PackageMgr != "" {
		lines = append(lines, "- **Package Manager**: "+info.PackageMgr)
	}
	if info.HasDocker {
		lines = append(lines, "- **Container**: Docker")
	}

	if len(lines) == 0 {
		return "- **Language**: [主要编程语言和版本]\n- **Framework**: [框架]\n- **Build**: [构建工具]"
	}

	return strings.Join(lines, "\n")
}

// generateConventions 生成代码规范
func generateConventions(info *ProjectInfo) string {
	var lines []string

	switch info.Language {
	case "Go":
		lines = append(lines,
			"- Follow Go standard project layout",
			"- Use `internal/` for private packages",
			"- Use `snake_case` for file names",
			"- Run `go fmt` before commit",
		)
	case "TypeScript", "JavaScript":
		lines = append(lines,
			"- Use ESLint for code quality",
			"- Prefer TypeScript over JavaScript",
			"- Use descriptive variable names",
		)
	case "Python":
		lines = append(lines,
			"- Follow PEP 8 style guide",
			"- Use type hints for functions",
			"- Use `snake_case` for files and functions",
		)
	default:
		lines = append(lines,
			"- [代码规范 1]",
			"- [代码规范 2]",
		)
	}

	return strings.Join(lines, "\n")
}

// generateArchitecture 生成架构说明
func generateArchitecture(info *ProjectInfo) string {
	var lines []string

	for _, dir := range info.MainDirs {
		lines = append(lines, "- `"+dir+"/` — [目录说明]")
	}

	if len(lines) == 0 {
		lines = append(lines,
			"- `src/` — [源代码目录]",
			"- `docs/` — [文档目录]",
		)
	}

	return strings.Join(lines, "\n")
}

// generateCommands 生成常用命令
func generateCommands(info *ProjectInfo) string {
	var lines []string

	switch info.Language {
	case "Go":
		lines = append(lines,
			"```bash",
			"# Development",
			"go run .",
			"",
			"# Build",
			"go build -o "+info.Name,
			"",
			"# Test",
			"go test ./...",
			"```",
		)
	case "TypeScript", "JavaScript":
		lines = append(lines,
			"```bash",
			"# Install dependencies",
			info.PackageMgr+" install",
			"",
			"# Development",
			info.PackageMgr+" run dev",
			"",
			"# Build",
			info.PackageMgr+" run build",
			"",
			"# Test",
			info.PackageMgr+" test",
			"```",
		)
	case "Python":
		lines = append(lines,
			"```bash",
			"# Install dependencies",
			"pip install -r requirements.txt",
			"",
			"# Run",
			"python main.py",
			"",
			"# Test",
			"pytest",
			"```",
		)
	default:
		lines = append(lines,
			"```bash",
			"# [开发命令]",
			"# [构建命令]",
			"# [测试命令]",
			"```",
		)
	}

	return strings.Join(lines, "\n")
}
