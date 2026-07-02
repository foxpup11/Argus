package continuity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HandoffGenerator 生成手交文件
type HandoffGenerator struct {
	homeDir string
}

// NewHandoffGenerator 创建新的手交生成器
func NewHandoffGenerator() (*HandoffGenerator, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户目录失败: %w", err)
	}
	return &HandoffGenerator{homeDir: homeDir}, nil
}

// GenerateMarkdown 生成 Markdown 格式的交接摘要
func (g *HandoffGenerator) GenerateMarkdown(summary *HandoffSummary) string {
	var sb strings.Builder

	// YAML frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: handoff-%s\n", summary.GeneratedAt.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("description: 会话交接摘要 - %s\n", summary.Project))
	sb.WriteString("metadata:\n")
	sb.WriteString("  type: handoff\n")
	sb.WriteString(fmt.Sprintf("  project: %s\n", summary.Project))
	sb.WriteString(fmt.Sprintf("  sessionsAnalyzed: %d\n", summary.SessionsUsed))
	sb.WriteString(fmt.Sprintf("  generatedAt: %s\n", summary.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString("---\n\n")

	// 标题
	sb.WriteString(fmt.Sprintf("# 会话交接摘要 - %s\n\n", summary.Project))

	// 概览
	sb.WriteString("## 概览\n\n")
	sb.WriteString(fmt.Sprintf("- **分析会话数**: %d / %d\n", summary.SessionsUsed, summary.SessionsTotal))
	sb.WriteString(fmt.Sprintf("- **生成时间**: %s\n", summary.GeneratedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("- **涉及文件数**: %d\n", len(summary.ModifiedFiles)))
	sb.WriteString(fmt.Sprintf("- **已完成任务**: %d\n", len(summary.CompletedTasks)))
	sb.WriteString(fmt.Sprintf("- **待办事项**: %d\n", len(summary.PendingTasks)))
	sb.WriteString(fmt.Sprintf("- **关键决策**: %d\n", len(summary.KeyDecisions)))
	sb.WriteString("\n")

	// 已完成任务
	if len(summary.CompletedTasks) > 0 {
		sb.WriteString("## [DONE] 已完成任务\n\n")
		for i, task := range summary.CompletedTasks {
			sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, task.Description))
			sb.WriteString(fmt.Sprintf("   - 会话: `%s`\n", task.SessionID[:8]))
			if task.VerifiedByGit {
				sb.WriteString("   - 状态: [verified] Git 已验证\n")
			} else {
				sb.WriteString("   - 状态: [unverified] 未验证\n")
			}
			if len(task.FilesChanged) > 0 {
				sb.WriteString("   - 文件: ")
				if len(task.FilesChanged) <= 5 {
					sb.WriteString(strings.Join(task.FilesChanged, ", "))
				} else {
					sb.WriteString(strings.Join(task.FilesChanged[:5], ", "))
					sb.WriteString(fmt.Sprintf(" 等 %d 个文件", len(task.FilesChanged)))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	// 待办事项
	if len(summary.PendingTasks) > 0 {
		sb.WriteString("## [TODO] 待办事项\n\n")
		for i, task := range summary.PendingTasks {
			sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, task.Description))
			sb.WriteString(fmt.Sprintf("   - 来源: %s (会话 `%s`)\n", task.Source, task.SessionID[:8]))
			if len(task.FilesHint) > 0 {
				sb.WriteString(fmt.Sprintf("   - 相关文件: %s\n", strings.Join(task.FilesHint, ", ")))
			}
			sb.WriteString("\n")
		}
	}

	// 关键决策
	if len(summary.KeyDecisions) > 0 {
		sb.WriteString("## [DECISION] 关键决策\n\n")
		for i, decision := range summary.KeyDecisions {
			sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, decision.Description))
			if decision.Context != "" {
				sb.WriteString(fmt.Sprintf("   - 背景: %s\n", truncateStr(decision.Context, 150)))
			}
			sb.WriteString("\n")
		}
	}

	// 修改的文件
	if len(summary.ModifiedFiles) > 0 {
		sb.WriteString("## 修改的文件\n\n")
		sb.WriteString("| 文件 | 操作次数 | 最后操作 | 类型 |\n")
		sb.WriteString("|------|---------|---------|------|\n")
		for _, fs := range summary.ModifiedFiles {
			fileType := "code"
			if fs.IsTestFile {
				fileType = "test"
			} else if fs.IsConfigFile {
				fileType = "config"
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %d | %s | %s |\n",
				truncatePath(fs.Path), fs.ActionCount, fs.LastAction, fileType))
		}
		sb.WriteString("\n")
	}

	// 已知问题
	if len(summary.KnownIssues) > 0 {
		sb.WriteString("## [WARN] 已知问题/陷阱\n\n")
		for i, issue := range summary.KnownIssues {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GeneratePrompt 生成可直接粘贴的 prompt 片段
func (g *HandoffGenerator) GeneratePrompt(summary *HandoffSummary) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## 上下文：我正在项目 %s 上工作\n\n", summary.Project))
	sb.WriteString("以下是最近的工作进展，请你在新会话中了解这些上下文：\n\n")

	// 已完成任务
	if len(summary.CompletedTasks) > 0 {
		sb.WriteString("### 已完成的工作\n\n")
		for i, task := range summary.CompletedTasks {
			sb.WriteString(fmt.Sprintf("%d. %s", i+1, task.Description))
			if task.VerifiedByGit {
				sb.WriteString(" [verified]")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// 待办事项
	if len(summary.PendingTasks) > 0 {
		sb.WriteString("### 待完成的工作\n\n")
		for i, task := range summary.PendingTasks {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, task.Description))
		}
		sb.WriteString("\n")
	}

	// 关键决策
	if len(summary.KeyDecisions) > 0 {
		sb.WriteString("### 关键决策\n\n")
		for _, decision := range summary.KeyDecisions {
			sb.WriteString(fmt.Sprintf("- %s\n", decision.Description))
		}
		sb.WriteString("\n")
	}

	// 已知问题
	if len(summary.KnownIssues) > 0 {
		sb.WriteString("### 已知问题\n\n")
		for _, issue := range summary.KnownIssues {
			sb.WriteString(fmt.Sprintf("- [warn] %s\n", issue))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// SaveToMemory 保存交接摘要到 memory 目录
func (g *HandoffGenerator) SaveToMemory(summary *HandoffSummary, content string) (string, error) {
	// 构建 memory 目录路径
	memoryDir := filepath.Join(g.homeDir, ".claude", "projects", summary.Project, "memory")

	// 确保目录存在
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return "", fmt.Errorf("创建 memory 目录失败: %w", err)
	}

	// 生成文件名
	filename := fmt.Sprintf("handoff-%s.md", summary.GeneratedAt.Format("2006-01-02"))
	filePath := filepath.Join(memoryDir, filename)

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("写入交接文件失败: %w", err)
	}

	return filePath, nil
}

// 辅助函数

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func truncatePath(path string) string {
	// 如果路径太短，直接返回
	if len(path) <= 50 {
		return path
	}

	// 尝试保留最后两级目录
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return path
	}

	// 保留最后两个部分
	lastTwo := strings.Join(parts[len(parts)-2:], "/")
	if len(lastTwo) <= 45 {
		return ".../" + lastTwo
	}

	return "..." + path[len(path)-47:]
}
