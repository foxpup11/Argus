package session

import "time"

// MessageType 消息类型
type MessageType string

const (
	MessageTypeUser      MessageType = "user"      // 用户消息
	MessageTypeAssistant MessageType = "assistant" // AI 回复
)

// ContentType 内容块类型
type ContentType string

const (
	ContentTypeText       ContentType = "text"        // 文本
	ContentTypeThinking   ContentType = "thinking"    // 思考过程
	ContentTypeToolUse    ContentType = "tool_use"    // 工具调用
	ContentTypeToolResult ContentType = "tool_result" // 工具结果
)

// Message 对话消息
type Message struct {
	ID        string         `json:"id"`        // 消息 UUID
	Type      MessageType    `json:"type"`       // 消息类型
	Content   []ContentBlock `json:"content"`    // 内容块列表
	Timestamp time.Time      `json:"timestamp"` // 时间戳
	Model     string         `json:"model"`     // 模型名称（AI 消息）
}

// ContentBlock 内容块
type ContentBlock struct {
	Type     ContentType `json:"type"`              // 内容类型
	Text     string      `json:"text,omitempty"`    // 文本内容
	Thinking string      `json:"thinking,omitempty"` // 思考过程
	ToolName string      `json:"toolName,omitempty"` // 工具名称
	ToolID   string      `json:"toolId,omitempty"`   // 工具调用 ID
	Input    any         `json:"input,omitempty"`    // 工具输入参数
	Result   string      `json:"result,omitempty"`   // 工具执行结果
}
