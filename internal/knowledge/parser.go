package knowledge

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// ParseFrontmatter 解析 YAML frontmatter
func ParseFrontmatter(content string) (map[string]string, string) {
	frontmatter := make(map[string]string)
	body := content

	// 检查是否以 --- 开头
	if !strings.HasPrefix(content, "---") {
		return frontmatter, body
	}

	// 查找结束标记
	endIndex := strings.Index(content[3:], "---")
	if endIndex == -1 {
		return frontmatter, body
	}

	// 提取 frontmatter 部分
	fmContent := content[3 : endIndex+3]
	body = content[endIndex+6:]

	// 简单解析 YAML（支持 key: value 格式）
	lines := strings.Split(fmContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			frontmatter[key] = value
		}
	}

	return frontmatter, strings.TrimSpace(body)
}

// ExtractTitle 从 Markdown 内容提取标题
func ExtractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

// GenerateRandomName 生成随机文件名
func GenerateRandomName() string {
	adjectives := []string{"async", "cheerful", "greedy", "iterative", "logical", "recursive", "swirling", "bold", "calm", "eager"}
	nouns := []string{"greeting", "sparking", "shimmying", "cuddling", "chasing", "wiggling", "dancing", "singing", "reading", "writing"}

	// 使用加密安全的随机数生成器
	b := make([]byte, 8)
	_, _ = rand.Read(b)

	// 使用随机字节生成索引
	adjIndex := int(b[0]) % len(adjectives)
	nounIndex := int(b[1]) % len(nouns)

	return adjectives[adjIndex] + "-" + nouns[nounIndex] + "-" + fmt.Sprintf("%x", b)
}

// ============================================
// CLAUDE.md 分节解析/序列化
// ============================================

// ClaudeMDSection CLAUDE.md 分节
type ClaudeMDSection struct {
	ID      string `json:"id"`      // "overview", "techstack", "conventions", "architecture", "commands"
	Title   string `json:"title"`   // 显示标题
	Content string `json:"content"` // 分节内容（不含标题行）
	Order   int    `json:"order"`   // 排序
}

// 预定义的 CLAUDE.md 分节 ID 和标题
var claudeMDSectionDefs = []struct {
	ID    string
	Title string
	Order int
}{
	{"overview", "Overview", 0},
	{"techstack", "Tech Stack", 1},
	{"conventions", "Conventions", 2},
	{"architecture", "Architecture", 3},
	{"commands", "Commands", 4},
}

// ParseClaudeMDSections 将 CLAUDE.md 内容解析为分节
func ParseClaudeMDSections(content string) []ClaudeMDSection {
	lines := strings.Split(content, "\n")

	// 找到所有 ## 标题的位置
	type heading struct {
		title   string
		lineIdx int
	}
	var headings []heading

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			title := strings.TrimPrefix(trimmed, "## ")
			headings = append(headings, heading{title: title, lineIdx: i})
		}
	}

	sections := make([]ClaudeMDSection, 0, len(headings)+1)

	// 提取第一个 # 标题之前的内容作为 "overview"（如果不在任何 ## 下）
	// 或者将第一个 ## 之前的内容作为概述

	// 构建分节
	for i, h := range headings {
		// 提取分节内容（从标题下一行到下一个标题之前）
		startLine := h.lineIdx + 1
		endLine := len(lines)
		if i+1 < len(headings) {
			endLine = headings[i+1].lineIdx
		}

		// 去除首尾空行
		sectionLines := lines[startLine:endLine]
		for len(sectionLines) > 0 && strings.TrimSpace(sectionLines[0]) == "" {
			sectionLines = sectionLines[1:]
		}
		for len(sectionLines) > 0 && strings.TrimSpace(sectionLines[len(sectionLines)-1]) == "" {
			sectionLines = sectionLines[:len(sectionLines)-1]
		}

		content := strings.Join(sectionLines, "\n")

		// 匹配预定义分节 ID
		id := matchSectionID(h.title)
		order := i

		// 查找预定义顺序
		for _, def := range claudeMDSectionDefs {
			if def.ID == id {
				order = def.Order
				break
			}
		}

		sections = append(sections, ClaudeMDSection{
			ID:      id,
			Title:   h.title,
			Content: content,
			Order:   order,
		})
	}

	// 如果没有找到任何 ## 标题，将整个内容作为 overview
	if len(sections) == 0 && strings.TrimSpace(content) != "" {
		// 去掉第一个 # 标题（如果有）
		body := content
		if idx := strings.Index(body, "\n"); idx != -1 {
			firstLine := strings.TrimSpace(body[:idx])
			if strings.HasPrefix(firstLine, "# ") {
				body = body[idx+1:]
			}
		}
		sections = append(sections, ClaudeMDSection{
			ID:      "overview",
			Title:   "Overview",
			Content: strings.TrimSpace(body),
			Order:   0,
		})
	}

	return sections
}

// matchSectionID 根据标题文本匹配分节 ID
func matchSectionID(title string) string {
	titleLower := strings.ToLower(strings.TrimSpace(title))

	// 精确匹配
	for _, def := range claudeMDSectionDefs {
		if strings.ToLower(def.Title) == titleLower {
			return def.ID
		}
	}

	// 模糊匹配
	aliases := map[string]string{
		"项目概述":    "overview",
		"概述":       "overview",
		"summary":   "overview",
		"技术栈":     "techstack",
		"技术":       "techstack",
		"tech":      "techstack",
		"技术选型":    "techstack",
		"代码规范":    "conventions",
		"规范":       "conventions",
		"约定":       "conventions",
		"coding":    "conventions",
		"style":     "conventions",
		"架构":       "architecture",
		"架构设计":    "architecture",
		"项目结构":    "architecture",
		"structure":  "architecture",
		"目录结构":    "architecture",
		"常用命令":    "commands",
		"命令":       "commands",
		"scripts":    "commands",
		"build":      "commands",
		"运行":       "commands",
	}

	if aliasID, ok := aliases[titleLower]; ok {
		return aliasID
	}

	// 使用标题作为 ID（保留原始大小写，转为 kebab-case）
	return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

// SerializeClaudeMDSections 将分节序列化为 CLAUDE.md 格式
func SerializeClaudeMDSections(projectName string, sections []ClaudeMDSection) string {
	var sb strings.Builder

	// 写入项目标题
	if projectName != "" {
		sb.WriteString("# " + projectName + "\n\n")
	}

	// 按 Order 排序
	sorted := make([]ClaudeMDSection, len(sections))
	copy(sorted, sections)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Order < sorted[i].Order {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	for _, s := range sorted {
		sb.WriteString("## " + s.Title + "\n\n")
		content := strings.TrimSpace(s.Content)
		if content != "" {
			sb.WriteString(content + "\n\n")
		}
	}

	return sb.String()
}

// GetClaudeMDTemplate 获取默认的 CLAUDE.md 模板
func GetClaudeMDTemplate(projectName string) string {
	return `# ` + projectName + `

## Overview

[项目概述：这个项目是做什么的，解决什么问题]

## Tech Stack

- **Language**: [主要编程语言和版本]
- **Framework**: [框架]
- **Build**: [构建工具]

## Conventions

- [代码规范 1]
- [代码规范 2]

## Architecture

- ` + "`" + `app.go` + "`" + ` — [主要入口文件]
- ` + "`" + `internal/` + "`" + ` — [内部包]

## Commands

` + "```" + `bash
# Development
[开发命令]

# Build
[构建命令]

# Test
[测试命令]
` + "```" + `
`
}

// GenerateClaudeMDTemplate 生成 CLAUDE.md 模板（用于 CreateDocument）
func GenerateClaudeMDTemplate(projectName string) string {
	return GetClaudeMDTemplate(projectName)
}

// GenerateTemplate 生成文档模板
func GenerateTemplate(docType DocType, title string) string {
	switch docType {
	case DocTypePlans:
		return `# ` + title + `

## Context

[描述背景和目标]

## Architecture Overview

[架构设计]

## Implementation Steps

### Step 1: [任务描述]

[详细说明]

## Verification

[验证方法]
`
	case DocTypeMemory:
		return `---
name: ` + strings.ToLower(strings.ReplaceAll(title, " ", "-")) + `
description: ` + title + `
metadata:
  node_type: memory
---

` + title + `

**Why:** [为什么需要这个记忆]

**How to apply:** [如何应用这个记忆]
`
	case DocTypeClaudeMD:
		return GetClaudeMDTemplate(title)
	default:
		return "# " + title + "\n\n"
	}
}
