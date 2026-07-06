package plugin

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ValidateHookConfig validates a hook configuration.
func ValidateHookConfig(hook HookConfig) error {
	return hook.Validate()
}

// ValidateMCPServerConfig validates an MCP server configuration.
func ValidateMCPServerConfig(server MCPServerConfig) error {
	return server.Validate()
}

// ParseSettingsFile parses Claude Code settings.json file content.
func ParseSettingsFile(data []byte) (*PluginSettings, error) {
	return FromJSON(data)
}

// ValidateMatcherPattern validates a hook matcher pattern.
// The matcher should be a pipe-separated list of tool names or regex patterns.
func ValidateMatcherPattern(matcher string) error {
	if strings.TrimSpace(matcher) == "" {
		return fmt.Errorf("matcher pattern cannot be empty")
	}

	// Split by pipe and validate each part
	parts := strings.Split(matcher, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return fmt.Errorf("empty matcher pattern in: %s", matcher)
		}

		// Try to compile as regex
		if _, err := regexp.Compile(part); err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", part, err)
		}
	}

	return nil
}

// ValidateCommand validates a command string.
func ValidateCommand(command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Check for dangerous patterns (basic safety check)
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=/dev/zero",
		"mkfs",
		":(){ :|:& };:",
	}

	lowerCmd := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return fmt.Errorf("command contains potentially dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// ValidateEnvVars validates environment variables map.
func ValidateEnvVars(env map[string]string) error {
	for key, value := range env {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("environment variable key cannot be empty")
		}
		// Check for valid env var name (letters, digits, underscores)
		matched, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, key)
		if err != nil {
			return fmt.Errorf("failed to validate environment variable name: %w", err)
		}
		if !matched {
			return fmt.Errorf("invalid environment variable name: %s", key)
		}

		_ = value // Value can be empty
	}

	return nil
}

// ValidateArgs validates command arguments.
func ValidateArgs(args []string) error {
	for i, arg := range args {
		if strings.TrimSpace(arg) == "" {
			return fmt.Errorf("argument %d cannot be empty", i+1)
		}
	}
	return nil
}

// FormatSettingsAsJSON formats plugin settings as pretty JSON.
func FormatSettingsAsJSON(settings *PluginSettings) (string, error) {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format settings: %w", err)
	}
	return string(data), nil
}

// ExtractHooksFromSettings extracts hook configurations from raw settings JSON.
func ExtractHooksFromSettings(data []byte) ([]HookConfig, error) {
	settings, err := FromJSON(data)
	if err != nil {
		return nil, err
	}
	return settings.Hooks, nil
}

// ExtractMCPServersFromSettings extracts MCP server configurations from raw settings JSON.
func ExtractMCPServersFromSettings(data []byte) ([]MCPServerConfig, error) {
	settings, err := FromJSON(data)
	if err != nil {
		return nil, err
	}
	return settings.MCPServers, nil
}
