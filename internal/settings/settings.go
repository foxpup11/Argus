// Package settings provides application settings management.
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"agentscope-desktop/internal/session"
)

// Theme 主题类型
type Theme string

const (
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
	ThemeAuto  Theme = "auto"
)

// Settings 应用设置
type Settings struct {
	mu sync.RWMutex `json:"-"`

	// 主题设置
	Theme Theme `json:"theme"`

	// 风险规则
	CustomRules []CustomRule `json:"customRules"`
}

// CustomRule 自定义风险规则
type CustomRule struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Level       session.RiskLevel `json:"level"`
	Pattern     string           `json:"pattern"` // 文件路径匹配模式
	Enabled     bool             `json:"enabled"`
}

// Manager 设置管理器
type Manager struct {
	settings *Settings
	filePath string
}

// NewManager 创建设置管理器
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".agentscope")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	m := &Manager{
		settings: &Settings{
			Theme:       ThemeAuto,
			CustomRules: []CustomRule{},
		},
		filePath: filepath.Join(configDir, "settings.json"),
	}

	// 加载已有的设置
	m.load()

	return m, nil
}

// Get 获取设置
func (m *Manager) Get() *Settings {
	m.settings.mu.RLock()
	defer m.settings.mu.RUnlock()

	// 返回副本（不复制锁）
	result := &Settings{
		Theme:       m.settings.Theme,
		CustomRules: make([]CustomRule, len(m.settings.CustomRules)),
	}
	copy(result.CustomRules, m.settings.CustomRules)
	return result
}

// UpdateTheme 更新主题设置
func (m *Manager) UpdateTheme(theme Theme) error {
	m.settings.mu.Lock()
	defer m.settings.mu.Unlock()

	m.settings.Theme = theme
	return m.save()
}

// AddCustomRule 添加自定义规则
func (m *Manager) AddCustomRule(rule CustomRule) error {
	m.settings.mu.Lock()
	defer m.settings.mu.Unlock()

	m.settings.CustomRules = append(m.settings.CustomRules, rule)
	return m.save()
}

// RemoveCustomRule 删除自定义规则
func (m *Manager) RemoveCustomRule(name string) error {
	m.settings.mu.Lock()
	defer m.settings.mu.Unlock()

	for i, rule := range m.settings.CustomRules {
		if rule.Name == name {
			m.settings.CustomRules = append(
				m.settings.CustomRules[:i],
				m.settings.CustomRules[i+1:]...,
			)
			return m.save()
		}
	}

	return nil
}

// UpdateCustomRule 更新自定义规则
func (m *Manager) UpdateCustomRule(name string, rule CustomRule) error {
	m.settings.mu.Lock()
	defer m.settings.mu.Unlock()

	for i, r := range m.settings.CustomRules {
		if r.Name == name {
			m.settings.CustomRules[i] = rule
			return m.save()
		}
	}

	return nil
}

// load 从文件加载设置
func (m *Manager) load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，使用默认设置
		}
		return err
	}

	m.settings.mu.Lock()
	defer m.settings.mu.Unlock()

	return json.Unmarshal(data, m.settings)
}

// save 保存设置到文件
func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}
