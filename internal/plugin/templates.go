package plugin

// GetBuiltinTemplates returns all built-in hook templates.
func GetBuiltinTemplates() []HookTemplate {
	return []HookTemplate{
		// 格式化类模板
		{
			Name:        "保存时 Prettier 格式化",
			Description: "保存文件时自动运行 Prettier 格式化代码",
			Category:    "格式化",
			Hook: HookConfig{
				Type:     HookPostToolUse,
				Matcher:  "Write|Edit",
				Commands: []string{"prettier --write $FILE"},
				Enabled:  true,
			},
		},
		{
			Name:        "保存时 Black 格式化",
			Description: "保存 Python 文件时自动运行 Black 格式化",
			Category:    "格式化",
			Hook: HookConfig{
				Type:     HookPostToolUse,
				Matcher:  "Write|Edit",
				Commands: []string{"black $FILE"},
				Enabled:  true,
			},
		},
		{
			Name:        "保存时 Goimports 格式化",
			Description: "保存 Go 文件时自动运行 goimports",
			Category:    "格式化",
			Hook: HookConfig{
				Type:     HookPostToolUse,
				Matcher:  "Write|Edit",
				Commands: []string{"goimports -w $FILE"},
				Enabled:  true,
			},
		},

		// 测试类模板
		{
			Name:        "提交前运行单元测试",
			Description: "执行 commit 前自动运行单元测试",
			Category:    "测试",
			Hook: HookConfig{
				Type:     HookPreToolUse,
				Matcher:  "Bash",
				Commands: []string{"npm test"},
				Enabled:  false,
			},
		},
		{
			Name:        "提交前运行 Go 测试",
			Description: "执行 commit 前自动运行 Go 测试",
			Category:    "测试",
			Hook: HookConfig{
				Type:     HookPreToolUse,
				Matcher:  "Bash",
				Commands: []string{"go test ./..."},
				Enabled:  false,
			},
		},
		{
			Name:        "提交前运行 Python 测试",
			Description: "执行 commit 前自动运行 pytest",
			Category:    "测试",
			Hook: HookConfig{
				Type:     HookPreToolUse,
				Matcher:  "Bash",
				Commands: []string{"pytest"},
				Enabled:  false,
			},
		},

		// 安全类模板
		{
			Name:        "拦截危险删除命令",
			Description: "拦截 rm -rf 等危险的删除命令",
			Category:    "安全",
			Hook: HookConfig{
				Type:     HookPreToolUse,
				Matcher:  "Bash",
				Commands: []string{"echo '⚠️ 危险命令已被拦截: rm -rf'"},
				Enabled:  true,
			},
		},
		{
			Name:        "拦截权限修改命令",
			Description: "拦截 chmod 777 等危险的权限修改命令",
			Category:    "安全",
			Hook: HookConfig{
				Type:     HookPreToolUse,
				Matcher:  "Bash",
				Commands: []string{"echo '⚠️ 危险命令已被拦截: chmod 777'"},
				Enabled:  true,
			},
		},
		{
			Name:        "拦截管道执行命令",
			Description: "拦截 curl | bash 等管道执行命令",
			Category:    "安全",
			Hook: HookConfig{
				Type:     HookPreToolUse,
				Matcher:  "Bash",
				Commands: []string{"echo '⚠️ 危险命令已被拦截: 管道执行'"},
				Enabled:  true,
			},
		},

		// 日志类模板
		{
			Name:        "记录文件修改日志",
			Description: "记录所有文件修改操作到日志文件",
			Category:    "日志",
			Hook: HookConfig{
				Type:     HookPostToolUse,
				Matcher:  "Write|Edit",
				Commands: []string{"echo \"$(date): Modified $FILE\" >> .file-changes.log"},
				Enabled:  false,
			},
		},
		{
			Name:        "记录 Bash 命令日志",
			Description: "记录所有 Bash 命令执行到日志文件",
			Category:    "日志",
			Hook: HookConfig{
				Type:     HookPostToolUse,
				Matcher:  "Bash",
				Commands: []string{"echo \"$(date): Executed command\" >> .command-changes.log"},
				Enabled:  false,
			},
		},

		// 通知类模板
		{
			Name:        "会话结束通知",
			Description: "会话结束时发送桌面通知",
			Category:    "通知",
			Hook: HookConfig{
				Type:     HookStop,
				Matcher:  "*",
				Commands: []string{"echo '✅ Claude Code 会话已结束'"},
				Enabled:  false,
			},
		},
	}
}

// GetTemplateCategories returns all unique template categories.
func GetTemplateCategories() []string {
	templates := GetBuiltinTemplates()
	seen := make(map[string]bool)
	var categories []string

	for _, t := range templates {
		if !seen[t.Category] {
			seen[t.Category] = true
			categories = append(categories, t.Category)
		}
	}

	return categories
}

// GetTemplatesByCategory returns templates filtered by category.
func GetTemplatesByCategory(category string) []HookTemplate {
	templates := GetBuiltinTemplates()
	var result []HookTemplate

	for _, t := range templates {
		if t.Category == category {
			result = append(result, t)
		}
	}

	return result
}
