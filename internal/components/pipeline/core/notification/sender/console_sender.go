package sender

import (
	"Storage/internal/logic/workflows/notification"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConsoleSenderConfig 控制台发送器配置
type ConsoleSenderConfig struct {
	// 是否启用彩色输出
	ColorEnabled bool `json:"color_enabled"`

	// 是否启用详细日志
	VerboseLogging bool `json:"verbose_logging"`

	// 是否显示时间戳
	ShowTimestamp bool `json:"show_timestamp"`

	// 是否显示消息类型
	ShowType bool `json:"show_type"`

	// 是否显示优先级
	ShowPriority bool `json:"show_priority"`

	// 是否显示元数据
	ShowMetadata bool `json:"show_metadata"`
}

// ConsoleSender 控制台通知发送器
type ConsoleSender struct {
	// 发送器ID
	id string

	// 发送器名称
	name string

	// 配置
	config ConsoleSenderConfig

	// 是否关闭
	closed bool

	// 互斥锁
	mu sync.RWMutex
}

// NewConsoleSender 创建控制台通知发送器
func NewConsoleSender(name string, config ConsoleSenderConfig) *ConsoleSender {
	return &ConsoleSender{
		id:     uuid.New().String(),
		name:   name,
		config: config,
		closed: false,
	}
}

// Send 发送通知
func (s *ConsoleSender) Send(ctx context.Context, message *notification.Message) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return &notification.NotificationError{
			Message: "发送器已关闭",
			Code:    "SENDER_CLOSED",
		}
	}

	// 格式化消息
	formattedMsg := s.formatMessage(message)

	// 打印到控制台
	log.Println(formattedMsg)

	if s.config.VerboseLogging {
		log.Printf("消息 %s 已成功发送到控制台\n", message.ID)
	}

	return nil
}

// BatchSend 批量发送通知
func (s *ConsoleSender) BatchSend(ctx context.Context, messages []*notification.Message) error {
	if len(messages) == 0 {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return &notification.NotificationError{
			Message: "发送器已关闭",
			Code:    "SENDER_CLOSED",
		}
	}

	for _, message := range messages {
		// 格式化并打印每条消息
		formattedMsg := s.formatMessage(message)
		log.Println(formattedMsg)
	}

	if s.config.VerboseLogging {
		log.Printf("批量发送了 %d 条消息到控制台\n", len(messages))
	}

	return nil
}

// formatMessage 格式化消息
func (s *ConsoleSender) formatMessage(message *notification.Message) string {
	var result string

	// 添加时间戳
	if s.config.ShowTimestamp {
		timestamp := message.CreatedAt.Format(time.RFC3339)
		result += fmt.Sprintf("[%s] ", timestamp)
	}

	// 添加消息类型
	if s.config.ShowType {
		typeStr := string(message.Type)
		if s.config.ColorEnabled {
			switch message.Type {
			case notification.TypeInfo:
				typeStr = "\033[34m" + typeStr + "\033[0m" // 蓝色
			case notification.TypeWarning:
				typeStr = "\033[33m" + typeStr + "\033[0m" // 黄色
			case notification.TypeError:
				typeStr = "\033[31m" + typeStr + "\033[0m" // 红色
			}
		}
		result += fmt.Sprintf("[%s] ", typeStr)
	}

	// 添加优先级
	if s.config.ShowPriority {
		priorityStr := string(message.Priority)
		if s.config.ColorEnabled {
			switch message.Priority {
			case notification.PriorityLow:
				priorityStr = "\033[37m" + priorityStr + "\033[0m" // 灰色
			case notification.PriorityNormal:
				priorityStr = "\033[32m" + priorityStr + "\033[0m" // 绿色
			case notification.PriorityHigh:
				priorityStr = "\033[33m" + priorityStr + "\033[0m" // 黄色
			case notification.PriorityUrgent:
				priorityStr = "\033[31m" + priorityStr + "\033[0m" // 红色
			}
		}
		result += fmt.Sprintf("[%s] ", priorityStr)
	}

	// 添加标题和内容
	if s.config.ColorEnabled {
		result += fmt.Sprintf("\033[1m%s\033[0m: %s", message.Title, message.Content)
	} else {
		result += fmt.Sprintf("%s: %s", message.Title, message.Content)
	}

	// 添加元数据
	if s.config.ShowMetadata && len(message.Metadata) > 0 {
		result += "\n元数据: "
		for key, value := range message.Metadata {
			result += fmt.Sprintf("%s=%v ", key, value)
		}
	}

	return result
}

// GetID 获取发送器ID
func (s *ConsoleSender) GetID() string {
	return s.id
}

// GetName 获取发送器名称
func (s *ConsoleSender) GetName() string {
	return s.name
}

// Close 关闭发送器
func (s *ConsoleSender) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	if s.config.VerboseLogging {
		log.Printf("控制台发送器 %s 已关闭\n", s.name)
	}

	return nil
}
