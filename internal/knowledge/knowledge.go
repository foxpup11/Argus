// Package knowledge provides knowledge management capabilities for
// organizing and managing Claude Code's plans, memory, and other markdown documents.
package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// DocType 文档类型
type DocType string

const (
	DocTypePlans    DocType = "plans"
	DocTypeMemory   DocType = "memory"
	DocTypeClaudeMD DocType = "claudemd"
)

// KnowledgeDoc 知识文档
type KnowledgeDoc struct {
	Path        string            `json:"path"`        // 文件路径
	Name        string            `json:"name"`        // 显示名称
	Type        DocType           `json:"type"`        // 文档类型
	Project     string            `json:"project"`     // 所属项目
	Content     string            `json:"content"`     // Markdown 内容
	Frontmatter map[string]string `json:"frontmatter"` // YAML frontmatter
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	Size        int64             `json:"size"`
}

// SearchFilters 搜索筛选条件
type SearchFilters struct {
	Types    []DocType `json:"types,omitempty"`    // 按类型筛选
	Projects []string  `json:"projects,omitempty"` // 按项目筛选
}

// Engine 知识管理引擎
type Engine struct {
	homeDir string
	mu      sync.RWMutex
}

// NewEngine 创建新的知识管理引擎
func NewEngine() (*Engine, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &Engine{
		homeDir: homeDir,
	}, nil
}

// GetAllDocuments 获取所有文档
func (e *Engine) GetAllDocuments(docType string, project string) ([]KnowledgeDoc, error) {
	var docs []KnowledgeDoc

	// 判断是否需要扫描 plans
	scanPlans := docType == "" || docType == "all" || docType == string(DocTypePlans)
	// 判断是否需要扫描 memory
	scanMemory := docType == "" || docType == "all" || docType == string(DocTypeMemory)

	// 扫描 plans/
	if scanPlans {
		plansDocs, err := e.scanPlans()
		if err == nil {
			docs = append(docs, plansDocs...)
		}
	}

	// 扫描 projects/*/memory/
	if scanMemory {
		memoryDocs, err := e.scanMemory(project)
		if err == nil {
			docs = append(docs, memoryDocs...)
		}
	}

	// 按修改时间倒序排列
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].UpdatedAt.After(docs[j].UpdatedAt)
	})

	return docs, nil
}

// GetDocument 获取单个文档
func (e *Engine) GetDocument(path string) (*KnowledgeDoc, error) {
	return e.readDocument(path)
}

// SaveDocument 保存文档
func (e *Engine) SaveDocument(path string, content string) error {
	// 验证路径安全性
	if err := e.validatePath(path); err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(path, []byte(content), 0644)
}

// DeleteDocument 删除文档
func (e *Engine) DeleteDocument(path string) error {
	// 验证路径安全性
	if err := e.validatePath(path); err != nil {
		return err
	}

	return os.Remove(path)
}

// validatePath 验证文件路径是否在允许的目录内
func (e *Engine) validatePath(path string) error {
	// 解析路径为绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// 检查路径是否在允许的目录内
	allowedDirs := []string{
		filepath.Join(e.homeDir, ".claude"),
	}

	for _, dir := range allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absDir) {
			return nil
		}
	}

	return fmt.Errorf("access denied: path outside allowed directory")
}

// CreateDocument 创建新文档
func (e *Engine) CreateDocument(docType DocType, title string, content string, project string) (string, error) {
	var path string

	switch docType {
	case DocTypePlans:
		// plans/ 目录下使用随机文件名
		filename := GenerateRandomName() + ".md"
		path = filepath.Join(e.homeDir, ".claude", "plans", filename)
	case DocTypeMemory:
		// memory/ 需要指定项目
		if project == "" {
			// 尝试获取第一个可用的项目
			projectsDir := filepath.Join(e.homeDir, ".claude", "projects")
			entries, err := os.ReadDir(projectsDir)
			if err != nil || len(entries) == 0 {
				return "", fmt.Errorf("no projects found, please specify a project")
			}
			// 使用第一个项目
			for _, entry := range entries {
				if entry.IsDir() {
					project = entry.Name()
					break
				}
			}
		}
		// 确保 memory 目录存在
		memoryDir := filepath.Join(e.homeDir, ".claude", "projects", project, "memory")
		if err := os.MkdirAll(memoryDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create memory directory: %w", err)
		}
		path = filepath.Join(memoryDir, title+".md")
	default:
		path = filepath.Join(e.homeDir, ".claude", "plans", title+".md")
	}

	// 如果没有提供内容，使用模板
	if content == "" {
		content = GenerateTemplate(docType, title)
	}

	// 保存文件
	if err := e.SaveDocument(path, content); err != nil {
		return "", err
	}

	return path, nil
}

// SearchDocuments 搜索文档
func (e *Engine) SearchDocuments(query string, filters SearchFilters) ([]KnowledgeDoc, error) {
	docs, err := e.GetAllDocuments("", "")
	if err != nil {
		return nil, err
	}

	var results []KnowledgeDoc
	query = strings.ToLower(query)

	for _, doc := range docs {
		// 应用类型筛选
		if len(filters.Types) > 0 {
			found := false
			for _, t := range filters.Types {
				if doc.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 应用项目筛选
		if len(filters.Projects) > 0 {
			found := false
			for _, p := range filters.Projects {
				if doc.Project == p {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 关键词搜索
		if query != "" {
			if strings.Contains(strings.ToLower(doc.Name), query) ||
				strings.Contains(strings.ToLower(doc.Content), query) {
				results = append(results, doc)
			}
		} else if len(filters.Types) > 0 || len(filters.Projects) > 0 {
			// 没有查询但有筛选条件，返回筛选后的结果
			results = append(results, doc)
		}
		// 如果没有查询也没有筛选条件，不返回任何结果（避免返回所有文档）
	}

	return results, nil
}

// readDocument 读取单个文档
func (e *Engine) readDocument(path string) (*KnowledgeDoc, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 判断文档类型
	docType := DocTypePlans
	project := ""

	// 检查路径结构判断类型（更精确的判断）
	// memory 文件路径格式: ~/.claude/projects/<project>/memory/*.md
	// plans 文件路径格式: ~/.claude/plans/*.md
	parts := strings.Split(path, string(os.PathSeparator))

	// 检查是否在 projects 目录下且包含 memory 目录
	inProjectsDir := false
	inMemoryDir := false
	for i, part := range parts {
		if part == "projects" {
			inProjectsDir = true
			// 下一个目录是项目名
			if i+1 < len(parts) {
				project = parts[i+1]
			}
		}
		if inProjectsDir && part == "memory" {
			inMemoryDir = true
		}
	}

	if inProjectsDir && inMemoryDir {
		docType = DocTypeMemory
	}

	// 解析 YAML frontmatter
	frontmatter, body := ParseFrontmatter(string(content))

	// 提取名称
	name := frontmatter["name"]
	if name == "" {
		// 从内容提取标题
		name = ExtractTitle(body)
	}
	if name == "" {
		// 使用文件名
		name = filepath.Base(path)
		name = strings.TrimSuffix(name, ".md")
	}

	return &KnowledgeDoc{
		Path:        path,
		Name:        name,
		Type:        docType,
		Project:     project,
		Content:     string(content),
		Frontmatter: frontmatter,
		CreatedAt:   info.ModTime(),
		UpdatedAt:   info.ModTime(),
		Size:        info.Size(),
	}, nil
}
