package notification

import (
	"context"
	"time"
)

// MessageType 消息类型
type MessageType string

const (
	// TypeInfo 信息类型消息
	TypeInfo MessageType = "info"

	// TypeWarning 警告类型消息
	TypeWarning MessageType = "warning"

	// TypeError 错误类型消息
	TypeError MessageType = "error"
)

// MessagePriority 消息优先级
type MessagePriority string

const (
	// PriorityLow 低优先级
	PriorityLow MessagePriority = "low"

	// PriorityNormal 普通优先级
	PriorityNormal MessagePriority = "normal"

	// PriorityHigh 高优先级
	PriorityHigh MessagePriority = "high"

	// PriorityUrgent 紧急优先级
	PriorityUrgent MessagePriority = "urgent"
)

// Message 通知消息
type Message struct {
	// 消息ID
	ID string `json:"id"`

	// 消息类型
	Type MessageType `json:"type"`

	// 消息优先级
	Priority MessagePriority `json:"priority"`

	// 消息标题
	Title string `json:"title"`

	// 消息内容
	Content string `json:"content"`

	// 源ID
	SourceID string `json:"source_id"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationSender 通知发送接口
type NotificationSender interface {
	// 发送通知
	Send(ctx context.Context, message *Message) error

	// 批量发送通知
	BatchSend(ctx context.Context, messages []*Message) error

	// 获取发送器ID
	GetID() string

	// 获取发送器名称
	GetName() string

	// 关闭发送器
	Close(ctx context.Context) error
}

// NotificationReceiver 通知接收接口
type NotificationReceiver interface {
	// 接收通知
	Receive(ctx context.Context) (*Message, error)

	// 确认消息已处理
	Acknowledge(ctx context.Context, messageID string) error

	// 获取接收器ID
	GetID() string

	// 获取接收器名称
	GetName() string

	// 关闭接收器
	Close(ctx context.Context) error
}

// NotificationError 通知错误
type NotificationError struct {
	// 错误消息
	Message string `json:"message"`

	// 错误代码
	Code string `json:"code"`

	// 源错误
	Err error `json:"-"`
}

type NotificationManager struct {
}

// Error 实现error接口
func (e *NotificationError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap 获取原始错误
func (e *NotificationError) Unwrap() error {
	return e.Err
}

// 消息验证器
type MessageValidator interface {
	// 验证消息
	Validate(message *Message) error
}

// NotificationFilter 通知过滤器
type NotificationFilter interface {
	// 过滤消息
	Filter(message *Message) bool
}

// 消息路由规则
type RoutingRule struct {
	// 发送器ID
	SenderID string

	// 条件过滤器
	Filter NotificationFilter

	// 是否启用
	Enabled bool
}
