package receiver

import (
	"Storage/internal/logic/workflows/notification"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryReceiverConfig 内存接收器配置
type MemoryReceiverConfig struct {
	// 消息缓冲区大小
	BufferSize int `json:"buffer_size"`

	// 消息保留时间（秒）
	RetentionSeconds int `json:"retention_seconds"`

	// 是否阻塞当缓冲区满
	BlockWhenFull bool `json:"block_when_full"`

	// 消息过滤器
	Filter notification.NotificationFilter `json:"-"`
}

// MemoryReceiver 内存通知接收器
type MemoryReceiver struct {
	// 接收器ID
	id string

	// 接收器名称
	name string

	// 配置
	config MemoryReceiverConfig

	// 消息通道
	messageChan chan *notification.Message

	// 已确认消息的集合
	acknowledgedMessages map[string]time.Time

	// 存储的消息映射表 (消息ID -> 消息)
	messages map[string]*notification.Message

	// 是否关闭
	closed bool

	// 互斥锁
	mu sync.RWMutex

	// 清理定时器
	cleanupTicker *time.Ticker

	// 关闭通道
	closeChan chan struct{}
}

// NewMemoryReceiver 创建内存通知接收器
func NewMemoryReceiver(name string, config MemoryReceiverConfig) *MemoryReceiver {
	// 设置默认值
	if config.BufferSize <= 0 {
		config.BufferSize = 100
	}
	if config.RetentionSeconds <= 0 {
		config.RetentionSeconds = 3600 // 1小时
	}

	receiver := &MemoryReceiver{
		id:                   uuid.New().String(),
		name:                 name,
		config:               config,
		messageChan:          make(chan *notification.Message, config.BufferSize),
		acknowledgedMessages: make(map[string]time.Time),
		messages:             make(map[string]*notification.Message),
		closed:               false,
		closeChan:            make(chan struct{}),
	}

	// 启动清理任务
	receiver.startCleanupTask()

	return receiver
}

// startCleanupTask 启动清理任务
func (r *MemoryReceiver) startCleanupTask() {
	r.cleanupTicker = time.NewTicker(time.Duration(r.config.RetentionSeconds/10) * time.Second)
	go func() {
		for {
			select {
			case <-r.cleanupTicker.C:
				r.cleanup()
			case <-r.closeChan:
				r.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanup 清理过期消息
func (r *MemoryReceiver) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	retentionDuration := time.Duration(r.config.RetentionSeconds) * time.Second

	// 清理已确认的过期消息
	for id, acknowledgedTime := range r.acknowledgedMessages {
		if now.Sub(acknowledgedTime) > retentionDuration {
			delete(r.acknowledgedMessages, id)
			delete(r.messages, id)
		}
	}

	// 清理未确认的过期消息
	for id, msg := range r.messages {
		if _, isAcknowledged := r.acknowledgedMessages[id]; !isAcknowledged {
			if now.Sub(msg.CreatedAt) > retentionDuration {
				delete(r.messages, id)
			}
		}
	}
}

// PushMessage 推送消息到接收器
func (r *MemoryReceiver) PushMessage(ctx context.Context, message *notification.Message) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return &notification.NotificationError{
			Message: "接收器已关闭",
			Code:    "RECEIVER_CLOSED",
		}
	}

	// 应用过滤器
	if r.config.Filter != nil && !r.config.Filter.Filter(message) {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	// 存储消息
	r.mu.Lock()
	r.messages[message.ID] = message
	r.mu.Unlock()

	// 发送到通道
	if r.config.BlockWhenFull {
		// 阻塞模式
		select {
		case r.messageChan <- message:
			return nil
		case <-ctx.Done():
			return &notification.NotificationError{
				Message: "发送消息超时",
				Code:    "SEND_TIMEOUT",
				Err:     ctx.Err(),
			}
		}
	} else {
		// 非阻塞模式
		select {
		case r.messageChan <- message:
			return nil
		default:
			// 通道已满，但我们已经存储了消息，所以这不是错误
			return nil
		}
	}
}

// Receive 接收通知
func (r *MemoryReceiver) Receive(ctx context.Context) (*notification.Message, error) {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return nil, &notification.NotificationError{
			Message: "接收器已关闭",
			Code:    "RECEIVER_CLOSED",
		}
	}
	r.mu.RUnlock()

	// 从通道接收消息
	select {
	case message := <-r.messageChan:
		return message, nil
	case <-ctx.Done():
		return nil, &notification.NotificationError{
			Message: "接收消息超时",
			Code:    "RECEIVE_TIMEOUT",
			Err:     ctx.Err(),
		}
	}
}

// Acknowledge 确认消息已处理
func (r *MemoryReceiver) Acknowledge(ctx context.Context, messageID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return &notification.NotificationError{
			Message: "接收器已关闭",
			Code:    "RECEIVER_CLOSED",
		}
	}

	// 检查消息是否存在
	if _, exists := r.messages[messageID]; !exists {
		return &notification.NotificationError{
			Message: fmt.Sprintf("消息不存在: %s", messageID),
			Code:    "MESSAGE_NOT_FOUND",
		}
	}

	// 标记为已确认
	r.acknowledgedMessages[messageID] = time.Now()

	return nil
}

// GetAllMessages 获取所有消息
func (r *MemoryReceiver) GetAllMessages(ctx context.Context) ([]*notification.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, &notification.NotificationError{
			Message: "接收器已关闭",
			Code:    "RECEIVER_CLOSED",
		}
	}

	// 复制所有消息
	result := make([]*notification.Message, 0, len(r.messages))
	for _, msg := range r.messages {
		result = append(result, msg)
	}

	return result, nil
}

// GetUnacknowledgedMessages 获取未确认的消息
func (r *MemoryReceiver) GetUnacknowledgedMessages(ctx context.Context) ([]*notification.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, &notification.NotificationError{
			Message: "接收器已关闭",
			Code:    "RECEIVER_CLOSED",
		}
	}

	// 查找未确认的消息
	result := make([]*notification.Message, 0)
	for id, msg := range r.messages {
		if _, isAcknowledged := r.acknowledgedMessages[id]; !isAcknowledged {
			result = append(result, msg)
		}
	}

	return result, nil
}

// GetAcknowledgedMessages 获取已确认的消息
func (r *MemoryReceiver) GetAcknowledgedMessages(ctx context.Context) ([]*notification.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, &notification.NotificationError{
			Message: "接收器已关闭",
			Code:    "RECEIVER_CLOSED",
		}
	}

	// 查找已确认的消息
	result := make([]*notification.Message, 0, len(r.acknowledgedMessages))
	for id := range r.acknowledgedMessages {
		if msg, exists := r.messages[id]; exists {
			result = append(result, msg)
		}
	}

	return result, nil
}

// GetMessage 获取指定ID的消息
func (r *MemoryReceiver) GetMessage(ctx context.Context, messageID string) (*notification.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, &notification.NotificationError{
			Message: "接收器已关闭",
			Code:    "RECEIVER_CLOSED",
		}
	}

	// 获取指定消息
	if msg, exists := r.messages[messageID]; exists {
		return msg, nil
	}

	return nil, &notification.NotificationError{
		Message: fmt.Sprintf("消息不存在: %s", messageID),
		Code:    "MESSAGE_NOT_FOUND",
	}
}

// GetID 获取接收器ID
func (r *MemoryReceiver) GetID() string {
	return r.id
}

// GetName 获取接收器名称
func (r *MemoryReceiver) GetName() string {
	return r.name
}

// MessageCount 获取消息数量
func (r *MemoryReceiver) MessageCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.messages)
}

// AcknowledgedCount 获取已确认消息数量
func (r *MemoryReceiver) AcknowledgedCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.acknowledgedMessages)
}

// UnacknowledgedCount 获取未确认消息数量
func (r *MemoryReceiver) UnacknowledgedCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.messages) - len(r.acknowledgedMessages)
}

// Close 关闭接收器
func (r *MemoryReceiver) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	// 停止清理任务
	close(r.closeChan)

	// 关闭消息通道
	close(r.messageChan)

	r.closed = true
	return nil
}
