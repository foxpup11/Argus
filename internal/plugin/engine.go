package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Engine manages Claude Code plugin and hook configurations.
type Engine struct {
	homeDir string
	mu      sync.RWMutex
}

// NewEngine creates a new plugin engine.
func NewEngine() (*Engine, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	return &Engine{
		homeDir: homeDir,
	}, nil
}

// GetClaudeSettingsPath returns the path to Claude Code's settings.json file.
func (e *Engine) GetClaudeSettingsPath(projectDir string) string {
	if projectDir != "" {
		return filepath.Join(projectDir, ".claude", "settings.json")
	}
	return filepath.Join(e.homeDir, ".claude", "settings.json")
}

// LoadSettings loads plugin settings from Claude Code's settings.json.
func (e *Engine) LoadSettings(projectDir string) (*PluginSettings, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	path := e.GetClaudeSettingsPath(projectDir)

	// If file doesn't exist, return empty settings
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &PluginSettings{
			Hooks:      []HookConfig{},
			MCPServers: []MCPServerConfig{},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	return FromJSON(data)
}

// SaveSettings saves plugin settings to Claude Code's settings.json.
func (e *Engine) SaveSettings(projectDir string, settings *PluginSettings) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	// Validate settings
	if validationErrors := settings.Validate(); len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %v", validationErrors)
	}

	path := e.GetClaudeSettingsPath(projectDir)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Load existing settings to preserve other fields
	existingData, err := os.ReadFile(path)
	var existingMap map[string]interface{}
	if err == nil {
		_ = json.Unmarshal(existingData, &existingMap)
	}
	if existingMap == nil {
		existingMap = make(map[string]interface{})
	}

	// Update hooks and mcpServers
	claudeSettings, err := settings.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize settings: %w", err)
	}

	var newMap map[string]interface{}
	if err := json.Unmarshal(claudeSettings, &newMap); err != nil {
		return fmt.Errorf("failed to parse new settings: %w", err)
	}

	// Merge: only update hooks and mcpServers, preserve other fields
	if hooks, ok := newMap["hooks"]; ok {
		existingMap["hooks"] = hooks
	} else {
		delete(existingMap, "hooks")
	}

	if mcpServers, ok := newMap["mcpServers"]; ok {
		existingMap["mcpServers"] = mcpServers
	} else {
		delete(existingMap, "mcpServers")
	}

	// Write back
	data, err := json.MarshalIndent(existingMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// AddHook adds a new hook configuration.
func (e *Engine) AddHook(projectDir string, hook HookConfig) error {
	settings, err := e.LoadSettings(projectDir)
	if err != nil {
		return err
	}

	if err := hook.Validate(); err != nil {
		return err
	}

	settings.Hooks = append(settings.Hooks, hook)
	return e.SaveSettings(projectDir, settings)
}

// UpdateHook updates an existing hook configuration by index.
func (e *Engine) UpdateHook(projectDir string, index int, hook HookConfig) error {
	settings, err := e.LoadSettings(projectDir)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(settings.Hooks) {
		return fmt.Errorf("hook index out of range: %d", index)
	}

	if err := hook.Validate(); err != nil {
		return err
	}

	settings.Hooks[index] = hook
	return e.SaveSettings(projectDir, settings)
}

// RemoveHook removes a hook configuration by index.
func (e *Engine) RemoveHook(projectDir string, index int) error {
	settings, err := e.LoadSettings(projectDir)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(settings.Hooks) {
		return fmt.Errorf("hook index out of range: %d", index)
	}

	settings.Hooks = append(settings.Hooks[:index], settings.Hooks[index+1:]...)
	return e.SaveSettings(projectDir, settings)
}

// AddMCPServer adds a new MCP server configuration.
func (e *Engine) AddMCPServer(projectDir string, server MCPServerConfig) error {
	settings, err := e.LoadSettings(projectDir)
	if err != nil {
		return err
	}

	if err := server.Validate(); err != nil {
		return err
	}

	settings.MCPServers = append(settings.MCPServers, server)
	return e.SaveSettings(projectDir, settings)
}

// UpdateMCPServer updates an existing MCP server configuration by index.
func (e *Engine) UpdateMCPServer(projectDir string, index int, server MCPServerConfig) error {
	settings, err := e.LoadSettings(projectDir)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(settings.MCPServers) {
		return fmt.Errorf("MCP server index out of range: %d", index)
	}

	if err := server.Validate(); err != nil {
		return err
	}

	settings.MCPServers[index] = server
	return e.SaveSettings(projectDir, settings)
}

// RemoveMCPServer removes an MCP server configuration by index.
func (e *Engine) RemoveMCPServer(projectDir string, index int) error {
	settings, err := e.LoadSettings(projectDir)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(settings.MCPServers) {
		return fmt.Errorf("MCP server index out of range: %d", index)
	}

	settings.MCPServers = append(settings.MCPServers[:index], settings.MCPServers[index+1:]...)
	return e.SaveSettings(projectDir, settings)
}

// GetHookTemplates returns all available hook templates.
func (e *Engine) GetHookTemplates() []HookTemplate {
	return GetBuiltinTemplates()
}

// ApplyHookTemplate applies a hook template to the settings.
func (e *Engine) ApplyHookTemplate(projectDir string, template HookTemplate) error {
	return e.AddHook(projectDir, template.Hook)
}

// ValidateSettings validates plugin settings without saving.
func (e *Engine) ValidateSettings(settings *PluginSettings) []ValidationError {
	if settings == nil {
		return []ValidationError{
			{Field: "settings", Message: "settings cannot be nil"},
		}
	}
	return settings.Validate()
}
