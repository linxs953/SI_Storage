package sender

import (
	"Storage/internal/logic/workflows/notification"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// WebhookSenderConfig Webhook发送器配置
type WebhookSenderConfig struct {
	// Webhook URL
	URL string `json:"url"`

	// HTTP方法
	Method string `json:"method"`

	// 请求超时(秒)
	TimeoutSeconds int `json:"timeout_seconds"`

	// 请求头
	Headers map[string]string `json:"headers"`

	// 是否包含元数据
	IncludeMetadata bool `json:"include_metadata"`

	// 重试次数
	RetryCount int `json:"retry_count"`

	// 重试间隔(秒)
	RetryIntervalSeconds int `json:"retry_interval_seconds"`

	// 自定义负载模板
	PayloadTemplate string `json:"payload_template"`
}

// WebhookSender Webhook通知发送器
type WebhookSender struct {
	// 发送器ID
	id string

	// 发送器名称
	name string

	// 配置
	config WebhookSenderConfig

	// HTTP客户端
	client *http.Client

	// 是否关闭
	closed bool

	// 互斥锁
	mu sync.RWMutex
}

// WebhookPayload Webhook负载
type WebhookPayload struct {
	// 消息ID
	ID string `json:"id"`

	// 消息标题
	Title string `json:"title"`

	// 消息内容
	Content string `json:"content"`

	// 消息类型
	Type string `json:"type"`

	// 消息优先级
	Priority string `json:"priority"`

	// 源ID
	SourceID string `json:"source_id"`

	// 创建时间
	CreatedAt string `json:"created_at"`

	// 发送时间
	SentAt string `json:"sent_at"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// 发送器信息
	Sender struct {
		// 发送器ID
		ID string `json:"id"`

		// 发送器名称
		Name string `json:"name"`
	} `json:"sender"`
}

// NewWebhookSender 创建Webhook通知发送器
func NewWebhookSender(name string, config WebhookSenderConfig) (*WebhookSender, error) {
	// 验证配置
	if config.URL == "" {
		return nil, &notification.NotificationError{
			Message: "Webhook URL不能为空",
			Code:    "INVALID_CONFIG",
		}
	}

	// 设置默认值
	if config.Method == "" {
		config.Method = "POST"
	}

	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 10
	}

	if config.Headers == nil {
		config.Headers = map[string]string{
			"Content-Type": "application/json",
		}
	} else if _, exists := config.Headers["Content-Type"]; !exists {
		config.Headers["Content-Type"] = "application/json"
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}

	return &WebhookSender{
		id:     uuid.New().String(),
		name:   name,
		config: config,
		client: client,
		closed: false,
	}, nil
}

// Send 发送通知
func (s *WebhookSender) Send(ctx context.Context, message *notification.Message) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return &notification.NotificationError{
			Message: "发送器已关闭",
			Code:    "SENDER_CLOSED",
		}
	}

	// 准备负载
	payload, err := s.preparePayload(message)
	if err != nil {
		return err
	}

	// 转换为JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return &notification.NotificationError{
			Message: "JSON序列化失败",
			Code:    "JSON_ERROR",
			Err:     err,
		}
	}

	// 发送请求，带重试
	var lastErr error
	for i := 0; i <= s.config.RetryCount; i++ {
		err := s.sendRequest(ctx, jsonPayload)
		if err == nil {
			return nil
		}

		lastErr = err

		// 如果是最后一次尝试，不要等待
		if i == s.config.RetryCount {
			break
		}

		// 等待重试
		retryInterval := time.Duration(s.config.RetryIntervalSeconds) * time.Second
		if retryInterval <= 0 {
			retryInterval = time.Second
		}

		select {
		case <-time.After(retryInterval):
			// 继续下一次重试
		case <-ctx.Done():
			return &notification.NotificationError{
				Message: "发送Webhook请求被取消",
				Code:    "CONTEXT_CANCELED",
				Err:     ctx.Err(),
			}
		}
	}

	// 所有尝试失败
	return &notification.NotificationError{
		Message: fmt.Sprintf("发送Webhook请求失败，已重试%d次", s.config.RetryCount),
		Code:    "WEBHOOK_FAILED",
		Err:     lastErr,
	}
}

// sendRequest 发送HTTP请求
func (s *WebhookSender) sendRequest(ctx context.Context, jsonPayload []byte) error {
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, s.config.Method, s.config.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return &notification.NotificationError{
			Message: "创建Webhook请求失败",
			Code:    "REQUEST_ERROR",
			Err:     err,
		}
	}

	// 设置请求头
	for key, value := range s.config.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return &notification.NotificationError{
			Message: "发送Webhook请求失败",
			Code:    "REQUEST_ERROR",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &notification.NotificationError{
			Message: "读取Webhook响应失败",
			Code:    "RESPONSE_ERROR",
			Err:     err,
		}
	}

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &notification.NotificationError{
			Message: fmt.Sprintf("Webhook请求返回非成功状态码: %d, 响应: %s", resp.StatusCode, string(body)),
			Code:    "HTTP_ERROR",
		}
	}

	return nil
}

// preparePayload 准备Webhook负载
func (s *WebhookSender) preparePayload(message *notification.Message) (*WebhookPayload, error) {
	payload := &WebhookPayload{
		ID:        message.ID,
		Title:     message.Title,
		Content:   message.Content,
		Type:      string(message.Type),
		Priority:  string(message.Priority),
		SourceID:  message.SourceID,
		CreatedAt: message.CreatedAt.Format(time.RFC3339),
		SentAt:    time.Now().Format(time.RFC3339),
	}

	// 包含元数据
	if s.config.IncludeMetadata && message.Metadata != nil {
		payload.Metadata = message.Metadata
	}

	// 设置发送器信息
	payload.Sender.ID = s.id
	payload.Sender.Name = s.name

	return payload, nil
}

// BatchSend 批量发送通知
func (s *WebhookSender) BatchSend(ctx context.Context, messages []*notification.Message) error {
	if len(messages) == 0 {
		return nil
	}

	// 创建批量负载
	batchPayload := make([]*WebhookPayload, 0, len(messages))
	for _, message := range messages {
		payload, err := s.preparePayload(message)
		if err != nil {
			continue
		}
		batchPayload = append(batchPayload, payload)
	}

	if len(batchPayload) == 0 {
		return nil
	}

	// 转换为JSON
	jsonPayload, err := json.Marshal(batchPayload)
	if err != nil {
		return &notification.NotificationError{
			Message: "JSON序列化失败",
			Code:    "JSON_ERROR",
			Err:     err,
		}
	}

	// 发送请求
	return s.sendRequest(ctx, jsonPayload)
}

// GetID 获取发送器ID
func (s *WebhookSender) GetID() string {
	return s.id
}

// GetName 获取发送器名称
func (s *WebhookSender) GetName() string {
	return s.name
}

// Close 关闭发送器
func (s *WebhookSender) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	s.client.CloseIdleConnections()
	return nil
}
