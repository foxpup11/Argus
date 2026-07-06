// Package plugin provides Claude Code plugin and hook configuration management.
package plugin

import (
	"encoding/json"
	"fmt"
	"strings"
)

// HookType represents the lifecycle event type for a hook.
type HookType string

const (
	HookPreToolUse   HookType = "PreToolUse"
	HookPostToolUse  HookType = "PostToolUse"
	HookStop         HookType = "Stop"
	HookNotification HookType = "Notification"
)

// ValidHookTypes is a list of all valid hook types.
var ValidHookTypes = []HookType{
	HookPreToolUse,
	HookPostToolUse,
	HookStop,
	HookNotification,
}

// IsValid checks if the hook type is valid.
func (ht HookType) IsValid() bool {
	for _, valid := range ValidHookTypes {
		if ht == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the hook type.
func (ht HookType) String() string {
	return string(ht)
}

// HookConfig represents a Claude Code hook configuration.
type HookConfig struct {
	Type     HookType `json:"type"`
	Matcher  string   `json:"matcher"`
	Commands []string `json:"commands"`
	Enabled  bool     `json:"enabled"`
}

// Validate validates the hook configuration.
func (h *HookConfig) Validate() error {
	if !h.Type.IsValid() {
		return fmt.Errorf("invalid hook type: %s", h.Type)
	}

	if strings.TrimSpace(h.Matcher) == "" {
		return fmt.Errorf("matcher cannot be empty")
	}

	if len(h.Commands) == 0 {
		return fmt.Errorf("at least one command is required")
	}

	for i, cmd := range h.Commands {
		if strings.TrimSpace(cmd) == "" {
			return fmt.Errorf("command %d cannot be empty", i+1)
		}
	}

	return nil
}

// TransportType represents the MCP server transport type.
type TransportType string

const (
	TransportStdio TransportType = "stdio"
	TransportSSE   TransportType = "sse"
	TransportHTTP  TransportType = "http"
)

// ValidTransportTypes is a list of all valid transport types.
var ValidTransportTypes = []TransportType{
	TransportStdio,
	TransportSSE,
	TransportHTTP,
}

// IsValid checks if the transport type is valid.
func (tt TransportType) IsValid() bool {
	for _, valid := range ValidTransportTypes {
		if tt == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the transport type.
func (tt TransportType) String() string {
	return string(tt)
}

// MCPServerConfig represents a Claude Code MCP server configuration.
type MCPServerConfig struct {
	Name      string            `json:"name"`
	Transport TransportType    `json:"transport"`
	Command   string            `json:"command,omitempty"`
	URL       string            `json:"url,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Enabled   bool              `json:"enabled"`
}

// Validate validates the MCP server configuration.
func (m *MCPServerConfig) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	if !m.Transport.IsValid() {
		return fmt.Errorf("invalid transport type: %s", m.Transport)
	}

	switch m.Transport {
	case TransportStdio:
		if strings.TrimSpace(m.Command) == "" {
			return fmt.Errorf("command is required for stdio transport")
		}
	case TransportSSE, TransportHTTP:
		if strings.TrimSpace(m.URL) == "" {
			return fmt.Errorf("url is required for %s transport", m.Transport)
		}
	}

	return nil
}

// PluginSettings represents the complete plugin studio configuration.
type PluginSettings struct {
	Hooks      []HookConfig      `json:"hooks"`
	MCPServers []MCPServerConfig `json:"mcpServers"`
}

// Validate validates the complete plugin settings.
func (ps *PluginSettings) Validate() []ValidationError {
	var errors []ValidationError

	for i, hook := range ps.Hooks {
		if err := hook.Validate(); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("hooks[%d]", i),
				Message: err.Error(),
			})
		}
	}

	for i, server := range ps.MCPServers {
		if err := server.Validate(); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("mcpServers[%d]", i),
				Message: err.Error(),
			})
		}
	}

	return errors
}

// ValidationError represents a validation error for a specific field.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// HookTemplate represents a predefined hook configuration template.
type HookTemplate struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Hook        HookConfig `json:"hook"`
}

// ToJSON serializes the plugin settings to JSON for Claude Code settings.json format.
func (ps *PluginSettings) ToJSON() ([]byte, error) {
	// Build the structure expected by Claude Code
	claudeSettings := make(map[string]interface{})

	// Convert hooks to Claude Code format
	if len(ps.Hooks) > 0 {
		claudeHooks := make(map[string]interface{})
		for _, hook := range ps.Hooks {
			if !hook.Enabled {
				continue
			}
			hookKey := string(hook.Type)
			if _, exists := claudeHooks[hookKey]; !exists {
				claudeHooks[hookKey] = []interface{}{}
			}
			hookEntry := map[string]interface{}{
				"matcher": hook.Matcher,
				"hooks": []interface{}{
					map[string]interface{}{
						"type":    "command",
						"command": strings.Join(hook.Commands, " && "),
					},
				},
			}
			claudeHooks[hookKey] = append(claudeHooks[hookKey].([]interface{}), hookEntry)
		}
		claudeSettings["hooks"] = claudeHooks
	}

	// Convert MCP servers to Claude Code format
	if len(ps.MCPServers) > 0 {
		mcpServers := make(map[string]interface{})
		for _, server := range ps.MCPServers {
			if !server.Enabled {
				continue
			}
			serverConfig := map[string]interface{}{
				"transport": string(server.Transport),
			}
			if server.Command != "" {
				serverConfig["command"] = server.Command
			}
			if server.URL != "" {
				serverConfig["url"] = server.URL
			}
			if len(server.Args) > 0 {
				serverConfig["args"] = server.Args
			}
			if len(server.Env) > 0 {
				serverConfig["env"] = server.Env
			}
			mcpServers[server.Name] = serverConfig
		}
		claudeSettings["mcpServers"] = mcpServers
	}

	return json.MarshalIndent(claudeSettings, "", "  ")
}

// FromJSON deserializes Claude Code settings.json format to PluginSettings.
func FromJSON(data []byte) (*PluginSettings, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	settings := &PluginSettings{
		Hooks:      []HookConfig{},
		MCPServers: []MCPServerConfig{},
	}

	// Parse hooks
	if hooksRaw, ok := raw["hooks"].(map[string]interface{}); ok {
		for hookTypeStr, hookEntriesRaw := range hooksRaw {
			hookType := HookType(hookTypeStr)
			if !hookType.IsValid() {
				continue
			}
			if hookEntries, ok := hookEntriesRaw.([]interface{}); ok {
				for _, entryRaw := range hookEntries {
					entry, ok := entryRaw.(map[string]interface{})
					if !ok {
						continue
					}
					matcher, _ := entry["matcher"].(string)
					var commands []string
					if hooks, ok := entry["hooks"].([]interface{}); ok && len(hooks) > 0 {
						if hook, ok := hooks[0].(map[string]interface{}); ok {
							if cmd, ok := hook["command"].(string); ok {
								commands = strings.Split(cmd, " && ")
							}
						}
					}
					if matcher != "" && len(commands) > 0 {
						settings.Hooks = append(settings.Hooks, HookConfig{
							Type:     hookType,
							Matcher:  matcher,
							Commands: commands,
							Enabled:  true,
						})
					}
				}
			}
		}
	}

	// Parse MCP servers
	if mcpRaw, ok := raw["mcpServers"].(map[string]interface{}); ok {
		for name, serverRaw := range mcpRaw {
			server, ok := serverRaw.(map[string]interface{})
			if !ok {
				continue
			}
			transport, _ := server["transport"].(string)
			command, _ := server["command"].(string)
			url, _ := server["url"].(string)
			var args []string
			if argsRaw, ok := server["args"].([]interface{}); ok {
				for _, arg := range argsRaw {
					if argStr, ok := arg.(string); ok {
						args = append(args, argStr)
					}
				}
			}
			env := make(map[string]string)
			if envRaw, ok := server["env"].(map[string]interface{}); ok {
				for k, v := range envRaw {
					if vStr, ok := v.(string); ok {
						env[k] = vStr
					}
				}
			}
			settings.MCPServers = append(settings.MCPServers, MCPServerConfig{
				Name:      name,
				Transport: TransportType(transport),
				Command:   command,
				URL:       url,
				Args:      args,
				Env:       env,
				Enabled:   true,
			})
		}
	}

	return settings, nil
}
