package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DefaultMessageValidator 默认消息验证器
type DefaultMessageValidator struct{}

// Validate 验证消息
func (v *DefaultMessageValidator) Validate(message *Message) error {
	if message == nil {
		return &NotificationError{Message: "消息不能为空"}
	}

	if message.Title == "" {
		return &NotificationError{Message: "消息标题不能为空"}
	}

	if message.SourceID == "" {
		return &NotificationError{Message: "消息源ID不能为空"}
	}

	return nil
}

// NotificationHub 通知中心
type NotificationHub struct {
	// 发送器映射表
	senders map[string]NotificationSender

	// 接收器映射表
	receivers map[string]NotificationReceiver

	// 消息验证器
	validator MessageValidator

	// 路由规则
	routingRules []RoutingRule

	// 锁
	mu sync.RWMutex
}

// NewNotificationHub 创建通知中心
func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		senders:      make(map[string]NotificationSender),
		receivers:    make(map[string]NotificationReceiver),
		validator:    &DefaultMessageValidator{},
		routingRules: make([]RoutingRule, 0),
	}
}

// RegisterSender 注册发送器
func (h *NotificationHub) RegisterSender(sender NotificationSender) error {
	if sender == nil {
		return &NotificationError{Message: "发送器不能为空"}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	id := sender.GetID()
	if _, exists := h.senders[id]; exists {
		return &NotificationError{Message: fmt.Sprintf("发送器ID已存在: %s", id)}
	}

	h.senders[id] = sender
	return nil
}

// UnregisterSender 注销发送器
func (h *NotificationHub) UnregisterSender(senderID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sender, exists := h.senders[senderID]; exists {
		// 先关闭发送器
		if err := sender.Close(context.Background()); err != nil {
			return &NotificationError{
				Message: fmt.Sprintf("关闭发送器失败: %s", senderID),
				Err:     err,
			}
		}
		delete(h.senders, senderID)
		return nil
	}

	return &NotificationError{Message: fmt.Sprintf("发送器不存在: %s", senderID)}
}

// RegisterReceiver 注册接收器
func (h *NotificationHub) RegisterReceiver(receiver NotificationReceiver) error {
	if receiver == nil {
		return &NotificationError{Message: "接收器不能为空"}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	id := receiver.GetID()
	if _, exists := h.receivers[id]; exists {
		return &NotificationError{Message: fmt.Sprintf("接收器ID已存在: %s", id)}
	}

	h.receivers[id] = receiver
	return nil
}

// UnregisterReceiver 注销接收器
func (h *NotificationHub) UnregisterReceiver(receiverID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if receiver, exists := h.receivers[receiverID]; exists {
		// 先关闭接收器
		if err := receiver.Close(context.Background()); err != nil {
			return &NotificationError{
				Message: fmt.Sprintf("关闭接收器失败: %s", receiverID),
				Err:     err,
			}
		}
		delete(h.receivers, receiverID)
		return nil
	}

	return &NotificationError{Message: fmt.Sprintf("接收器不存在: %s", receiverID)}
}

// AddRoutingRule 添加路由规则
func (h *NotificationHub) AddRoutingRule(rule RoutingRule) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 检查发送器是否存在
	if _, exists := h.senders[rule.SenderID]; !exists {
		return &NotificationError{Message: fmt.Sprintf("发送器不存在: %s", rule.SenderID)}
	}

	h.routingRules = append(h.routingRules, rule)
	return nil
}

// RemoveRoutingRule 移除路由规则
func (h *NotificationHub) RemoveRoutingRule(senderID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	removed := false
	newRules := make([]RoutingRule, 0, len(h.routingRules))

	for _, rule := range h.routingRules {
		if rule.SenderID != senderID {
			newRules = append(newRules, rule)
		} else {
			removed = true
		}
	}

	if !removed {
		return &NotificationError{Message: fmt.Sprintf("未找到发送器的路由规则: %s", senderID)}
	}

	h.routingRules = newRules
	return nil
}

// EnableRoutingRule 启用路由规则
func (h *NotificationHub) EnableRoutingRule(senderID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, rule := range h.routingRules {
		if rule.SenderID == senderID {
			h.routingRules[i].Enabled = true
			return nil
		}
	}

	return &NotificationError{Message: fmt.Sprintf("未找到发送器的路由规则: %s", senderID)}
}

// DisableRoutingRule 禁用路由规则
func (h *NotificationHub) DisableRoutingRule(senderID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, rule := range h.routingRules {
		if rule.SenderID == senderID {
			h.routingRules[i].Enabled = false
			return nil
		}
	}

	return &NotificationError{Message: fmt.Sprintf("未找到发送器的路由规则: %s", senderID)}
}

// SetMessageValidator 设置消息验证器
func (h *NotificationHub) SetMessageValidator(validator MessageValidator) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if validator != nil {
		h.validator = validator
	}
}

// Send 发送消息
func (h *NotificationHub) Send(ctx context.Context, message *Message) error {
	// 消息预处理
	if err := h.prepareMessage(message); err != nil {
		return err
	}

	// 验证消息
	if err := h.validator.Validate(message); err != nil {
		return &NotificationError{
			Message: "消息验证失败",
			Err:     err,
		}
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// 找到匹配的路由规则
	var matchedSenderIDs []string
	for _, rule := range h.routingRules {
		if !rule.Enabled {
			continue
		}

		if rule.Filter == nil || rule.Filter.Filter(message) {
			matchedSenderIDs = append(matchedSenderIDs, rule.SenderID)
		}
	}

	// 如果没有匹配的规则，使用所有启用的发送器
	if len(matchedSenderIDs) == 0 {
		for _, rule := range h.routingRules {
			if rule.Enabled {
				matchedSenderIDs = append(matchedSenderIDs, rule.SenderID)
			}
		}
	}

	// 如果仍然没有匹配的发送器，返回错误
	if len(matchedSenderIDs) == 0 {
		return &NotificationError{Message: "没有可用的发送器"}
	}

	// 发送消息到每个匹配的发送器
	var lastErr error
	for _, senderID := range matchedSenderIDs {
		if sender, exists := h.senders[senderID]; exists {
			if err := sender.Send(ctx, message); err != nil {
				lastErr = &NotificationError{
					Message: fmt.Sprintf("通过发送器 %s 发送消息失败", senderID),
					Err:     err,
				}
			}
		}
	}

	return lastErr
}

// BatchSend 批量发送消息
func (h *NotificationHub) BatchSend(ctx context.Context, messages []*Message) error {
	if len(messages) == 0 {
		return nil
	}

	// 消息预处理和验证
	validMessages := make([]*Message, 0, len(messages))
	for _, msg := range messages {
		if err := h.prepareMessage(msg); err != nil {
			continue
		}

		if err := h.validator.Validate(msg); err != nil {
			continue
		}

		validMessages = append(validMessages, msg)
	}

	if len(validMessages) == 0 {
		return &NotificationError{Message: "没有有效的消息可发送"}
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// 获取所有启用的发送器
	var enabledSenders []NotificationSender
	for _, rule := range h.routingRules {
		if !rule.Enabled {
			continue
		}

		if sender, exists := h.senders[rule.SenderID]; exists {
			enabledSenders = append(enabledSenders, sender)
		}
	}

	if len(enabledSenders) == 0 {
		return &NotificationError{Message: "没有可用的发送器"}
	}

	// 发送消息到每个启用的发送器
	var lastErr error
	for _, sender := range enabledSenders {
		if err := sender.BatchSend(ctx, validMessages); err != nil {
			lastErr = &NotificationError{
				Message: fmt.Sprintf("通过发送器 %s 批量发送消息失败", sender.GetID()),
				Err:     err,
			}
		}
	}

	return lastErr
}

// prepareMessage 消息预处理
func (h *NotificationHub) prepareMessage(message *Message) error {
	if message == nil {
		return &NotificationError{Message: "消息不能为空"}
	}

	// 确保消息有ID
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	// 确保消息有创建时间
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	// 确保消息类型有效
	if message.Type == "" {
		message.Type = TypeInfo
	}

	// 确保消息优先级有效
	if message.Priority == "" {
		message.Priority = PriorityNormal
	}

	// 确保metadata存在
	if message.Metadata == nil {
		message.Metadata = make(map[string]interface{})
	}

	return nil
}

// Close 关闭通知中心
func (h *NotificationHub) Close(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 关闭所有发送器
	for id, sender := range h.senders {
		if err := sender.Close(ctx); err != nil {
			return &NotificationError{
				Message: fmt.Sprintf("关闭发送器失败: %s", id),
				Err:     err,
			}
		}
	}

	// 关闭所有接收器
	for id, receiver := range h.receivers {
		if err := receiver.Close(ctx); err != nil {
			return &NotificationError{
				Message: fmt.Sprintf("关闭接收器失败: %s", id),
				Err:     err,
			}
		}
	}

	// 清空映射表和规则
	h.senders = make(map[string]NotificationSender)
	h.receivers = make(map[string]NotificationReceiver)
	h.routingRules = make([]RoutingRule, 0)

	return nil
}

// GetSenders 获取所有发送器
func (h *NotificationHub) GetSenders() map[string]NotificationSender {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]NotificationSender, len(h.senders))
	for id, sender := range h.senders {
		result[id] = sender
	}

	return result
}

// GetReceivers 获取所有接收器
func (h *NotificationHub) GetReceivers() map[string]NotificationReceiver {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]NotificationReceiver, len(h.receivers))
	for id, receiver := range h.receivers {
		result[id] = receiver
	}

	return result
}

// GetRoutingRules 获取所有路由规则
func (h *NotificationHub) GetRoutingRules() []RoutingRule {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]RoutingRule, len(h.routingRules))
	copy(result, h.routingRules)

	return result
}

// CreateMessage 创建消息
func (h *NotificationHub) CreateMessage(
	title string,
	content string,
	sourceID string,
	msgType MessageType,
	priority MessagePriority,
	metadata map[string]interface{},
) *Message {
	msg := &Message{
		ID:        uuid.New().String(),
		Title:     title,
		Content:   content,
		SourceID:  sourceID,
		Type:      msgType,
		Priority:  priority,
		CreatedAt: time.Now(),
		Metadata:  metadata,
	}

	if msg.Metadata == nil {
		msg.Metadata = make(map[string]interface{})
	}

	return msg
}
