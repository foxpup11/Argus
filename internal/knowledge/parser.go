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
	default:
		return "# " + title + "\n\n"
	}
}
