package knowledge

import (
	"os"
	"path/filepath"
	"strings"
)

// scanMemory 扫描所有项目的 memory 文件
func (e *Engine) scanMemory(project string) ([]KnowledgeDoc, error) {
	projectsDir := filepath.Join(e.homeDir, ".claude", "projects")

	var docs []KnowledgeDoc

	// 遍历项目目录
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 如果指定了项目，只扫描该项目
		if project != "" && entry.Name() != project {
			continue
		}

		memoryDir := filepath.Join(projectsDir, entry.Name(), "memory")
		if _, err := os.Stat(memoryDir); os.IsNotExist(err) {
			continue
		}

		// 扫描 memory 目录下的 .md 文件
		mdFiles, _ := filepath.Glob(filepath.Join(memoryDir, "*.md"))
		for _, mdPath := range mdFiles {
			doc, err := e.readMemoryFile(mdPath, entry.Name())
			if err != nil {
				continue
			}
			docs = append(docs, *doc)
		}
	}

	return docs, nil
}

// readMemoryFile 读取单个 memory 文件
func (e *Engine) readMemoryFile(path string, project string) (*KnowledgeDoc, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 解析 YAML frontmatter
	frontmatter, _ := ParseFrontmatter(string(content))

	// 提取名称（从 frontmatter 的 name 字段或文件名）
	name := frontmatter["name"]
	if name == "" {
		name = filepath.Base(path)
		// 安全地去掉 .md 后缀
		if strings.HasSuffix(name, ".md") && len(name) > 3 {
			name = name[:len(name)-3]
		}
	}

	return &KnowledgeDoc{
		Path:        path,
		Name:        name,
		Type:        DocTypeMemory,
		Project:     project,
		Content:     string(content),
		Frontmatter: frontmatter,
		CreatedAt:   info.ModTime(),
		UpdatedAt:   info.ModTime(),
		Size:        info.Size(),
	}, nil
}
