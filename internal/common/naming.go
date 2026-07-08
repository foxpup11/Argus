package common

import "strings"

// FormatProjectName 将项目目录名转换为可读的项目名称
// 例如: "-g-ltch-git-learn-argus-desktop" -> "argus-desktop"
func FormatProjectName(dirName string) string {
	// 去掉开头的连字符
	name := strings.TrimPrefix(dirName, "-")

	// 过滤空字符串并取最后两个段
	parts := strings.Split(name, "-")
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) >= 2 {
		return filtered[len(filtered)-2] + "-" + filtered[len(filtered)-1]
	}
	if len(filtered) == 1 {
		return filtered[0]
	}

	return name
}
