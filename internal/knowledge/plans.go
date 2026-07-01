package knowledge

import (
	"os"
	"path/filepath"
	"strings"
)

// scanPlans 扫描 plans 目录
func (e *Engine) scanPlans() ([]KnowledgeDoc, error) {
	plansDir := filepath.Join(e.homeDir, ".claude", "plans")

	if _, err := os.Stat(plansDir); os.IsNotExist(err) {
		return nil, nil
	}

	var docs []KnowledgeDoc

	mdFiles, _ := filepath.Glob(filepath.Join(plansDir, "*.md"))
	for _, mdPath := range mdFiles {
		doc, err := e.readPlansFile(mdPath)
		if err != nil {
			continue
		}
		docs = append(docs, *doc)
	}

	return docs, nil
}

// readPlansFile 读取单个 plans 文件
func (e *Engine) readPlansFile(path string) (*KnowledgeDoc, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 从内容提取标题（第一个 # 开头的行）
	name := ExtractTitle(string(content))
	if name == "" {
		// 使用文件名
		name = filepath.Base(path)
		// 安全地去掉 .md 后缀
		if strings.HasSuffix(name, ".md") && len(name) > 3 {
			name = name[:len(name)-3]
		}
	}

	return &KnowledgeDoc{
		Path:      path,
		Name:      name,
		Type:      DocTypePlans,
		Content:   string(content),
		CreatedAt: info.ModTime(),
		UpdatedAt: info.ModTime(),
		Size:      info.Size(),
	}, nil
}
