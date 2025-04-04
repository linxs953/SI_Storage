package sender

import (
	"Storage/internal/logic/workflows/notification"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EmailSenderConfig 邮件发送器配置
type EmailSenderConfig struct {
	// SMTP服务器主机
	Host string `json:"host"`

	// SMTP服务器端口
	Port int `json:"port"`

	// SMTP用户名
	Username string `json:"username"`

	// SMTP密码
	Password string `json:"password"`

	// 发件人邮箱
	FromEmail string `json:"from_email"`

	// 发件人名称
	FromName string `json:"from_name"`

	// 收件人邮箱列表
	ToEmails []string `json:"to_emails"`

	// 是否使用TLS
	UseTLS bool `json:"use_tls"`

	// HTML邮件模板
	HTMLTemplate string `json:"html_template"`

	// 纯文本邮件模板
	TextTemplate string `json:"text_template"`

	// 是否启用HTML邮件
	EnableHTML bool `json:"enable_html"`

	// 自定义邮件头
	Headers map[string]string `json:"headers"`
}

// EmailSender 邮件通知发送器
type EmailSender struct {
	// 发送器ID
	id string

	// 发送器名称
	name string

	// 配置
	config EmailSenderConfig

	// SMTP客户端
	client *smtp.Client

	// HTML模板
	htmlTpl *template.Template

	// 文本模板
	textTpl *template.Template

	// 是否关闭
	closed bool

	// 互斥锁
	mu sync.RWMutex
}

// EmailTemplateData 邮件模板数据
type EmailTemplateData struct {
	Title     string
	Content   string
	Type      string
	Priority  string
	CreatedAt time.Time
	Metadata  map[string]interface{}
}

// DefaultHTMLTemplate 默认HTML模板
const DefaultHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; color: #333; }
        .container { max-width: 600px; margin: 0 auto; border: 1px solid #ddd; border-radius: 5px; overflow: hidden; }
        .header { background-color: #f5f5f5; padding: 15px; border-bottom: 1px solid #ddd; }
        .content { padding: 20px; }
        .footer { background-color: #f5f5f5; padding: 15px; border-top: 1px solid #ddd; font-size: 12px; color: #777; }
        .info { background-color: #e8f5ff; }
        .warning { background-color: #fff8e8; }
        .error { background-color: #ffebeb; }
        .metadata { margin-top: 20px; font-size: 12px; color: #777; }
        .metadata table { width: 100%; border-collapse: collapse; }
        .metadata td { padding: 5px; border-bottom: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header {{.Type}}">
            <h2>{{.Title}}</h2>
            <div>Priority: {{.Priority}} | Time: {{.CreatedAt.Format "2006-01-02 15:04:05"}}</div>
        </div>
        <div class="content">
            <p>{{.Content}}</p>
            {{if .Metadata}}
            <div class="metadata">
                <h4>Metadata:</h4>
                <table>
                    {{range $key, $value := .Metadata}}
                    <tr>
                        <td>{{$key}}</td>
                        <td>{{$value}}</td>
                    </tr>
                    {{end}}
                </table>
            </div>
            {{end}}
        </div>
        <div class="footer">
            This is an automated notification. Please do not reply to this email.
        </div>
    </div>
</body>
</html>
`

// DefaultTextTemplate 默认文本模板
const DefaultTextTemplate = `
{{.Title}}
Priority: {{.Priority}} | Type: {{.Type}} | Time: {{.CreatedAt.Format "2006-01-02 15:04:05"}}

{{.Content}}

{{if .Metadata}}
Metadata:
{{range $key, $value := .Metadata}}
- {{$key}}: {{$value}}
{{end}}
{{end}}

This is an automated notification. Please do not reply to this email.
`

// NewEmailSender 创建邮件通知发送器
func NewEmailSender(name string, config EmailSenderConfig) (*EmailSender, error) {
	sender := &EmailSender{
		id:     uuid.New().String(),
		name:   name,
		config: config,
		closed: false,
	}

	// 初始化HTML模板
	if config.EnableHTML {
		tplContent := config.HTMLTemplate
		if tplContent == "" {
			tplContent = DefaultHTMLTemplate
		}

		htmlTpl, err := template.New("email_html").Parse(tplContent)
		if err != nil {
			return nil, &notification.NotificationError{
				Message: "HTML模板解析失败",
				Code:    "TEMPLATE_ERROR",
				Err:     err,
			}
		}
		sender.htmlTpl = htmlTpl
	}

	// 初始化文本模板
	tplContent := config.TextTemplate
	if tplContent == "" {
		tplContent = DefaultTextTemplate
	}

	textTpl, err := template.New("email_text").Parse(tplContent)
	if err != nil {
		return nil, &notification.NotificationError{
			Message: "文本模板解析失败",
			Code:    "TEMPLATE_ERROR",
			Err:     err,
		}
	}
	sender.textTpl = textTpl

	return sender, nil
}

// connectSMTP 连接SMTP服务器
func (s *EmailSender) connectSMTP() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已经有连接，关闭它
	if s.client != nil {
		s.client.Quit()
		s.client = nil
	}

	// 连接SMTP服务器
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	var err error
	var conn *smtp.Client

	if s.config.UseTLS {
		// 使用TLS连接
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}
		conn, err = smtp.Dial(addr)
		if err != nil {
			return &notification.NotificationError{
				Message: "SMTP连接失败",
				Code:    "SMTP_CONNECT_ERROR",
				Err:     err,
			}
		}

		if err = conn.StartTLS(tlsConfig); err != nil {
			return &notification.NotificationError{
				Message: "SMTP TLS启动失败",
				Code:    "SMTP_TLS_ERROR",
				Err:     err,
			}
		}
	} else {
		// 使用普通连接
		conn, err = smtp.Dial(addr)
		if err != nil {
			return &notification.NotificationError{
				Message: "SMTP连接失败",
				Code:    "SMTP_CONNECT_ERROR",
				Err:     err,
			}
		}
	}

	// 认证
	if s.config.Username != "" && s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err = conn.Auth(auth); err != nil {
			return &notification.NotificationError{
				Message: "SMTP认证失败",
				Code:    "SMTP_AUTH_ERROR",
				Err:     err,
			}
		}
	}

	s.client = conn
	return nil
}

// Send 发送通知
func (s *EmailSender) Send(ctx context.Context, message *notification.Message) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return &notification.NotificationError{
			Message: "发送器已关闭",
			Code:    "SENDER_CLOSED",
		}
	}

	// 检查收件人
	if len(s.config.ToEmails) == 0 {
		return &notification.NotificationError{
			Message: "未配置收件人",
			Code:    "NO_RECIPIENTS",
		}
	}

	// 连接SMTP服务器
	if s.client == nil {
		if err := s.connectSMTP(); err != nil {
			return err
		}
	}

	// 生成邮件内容
	emailContent, err := s.generateEmailContent(message)
	if err != nil {
		return err
	}

	// 设置发件人
	var from string
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	}

	// 发送邮件
	if err := s.client.Mail(from); err != nil {
		_ = s.client.Reset()
		return &notification.NotificationError{
			Message: "设置发件人失败",
			Code:    "SMTP_FROM_ERROR",
			Err:     err,
		}
	}

	// 设置收件人
	for _, to := range s.config.ToEmails {
		if err := s.client.Rcpt(to); err != nil {
			_ = s.client.Reset()
			return &notification.NotificationError{
				Message: fmt.Sprintf("设置收件人失败: %s", to),
				Code:    "SMTP_RCPT_ERROR",
				Err:     err,
			}
		}
	}

	// 发送邮件内容
	w, err := s.client.Data()
	if err != nil {
		_ = s.client.Reset()
		return &notification.NotificationError{
			Message: "准备发送数据失败",
			Code:    "SMTP_DATA_ERROR",
			Err:     err,
		}
	}

	_, err = w.Write([]byte(emailContent))
	if err != nil {
		_ = w.Close()
		_ = s.client.Reset()
		return &notification.NotificationError{
			Message: "发送邮件内容失败",
			Code:    "SMTP_WRITE_ERROR",
			Err:     err,
		}
	}

	err = w.Close()
	if err != nil {
		_ = s.client.Reset()
		return &notification.NotificationError{
			Message: "完成邮件发送失败",
			Code:    "SMTP_CLOSE_ERROR",
			Err:     err,
		}
	}

	return nil
}

// BatchSend 批量发送通知
func (s *EmailSender) BatchSend(ctx context.Context, messages []*notification.Message) error {
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

	// 对每条消息单独发送
	var lastErr error
	for _, message := range messages {
		if err := s.Send(ctx, message); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// generateEmailContent 生成邮件内容
func (s *EmailSender) generateEmailContent(message *notification.Message) (string, error) {
	// 准备模板数据
	data := EmailTemplateData{
		Title:     message.Title,
		Content:   message.Content,
		Type:      string(message.Type),
		Priority:  string(message.Priority),
		CreatedAt: message.CreatedAt,
		Metadata:  message.Metadata,
	}

	// 生成邮件头
	var headers bytes.Buffer
	headers.WriteString(fmt.Sprintf("From: %s\r\n", s.config.FromEmail))
	headers.WriteString(fmt.Sprintf("To: %s\r\n", s.config.ToEmails[0]))
	headers.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Title))
	headers.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	headers.WriteString("MIME-Version: 1.0\r\n")

	// 添加自定义邮件头
	for key, value := range s.config.Headers {
		headers.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	if s.config.EnableHTML && s.htmlTpl != nil {
		// 生成多部分邮件 (HTML + 纯文本)
		boundary := "boundary_" + uuid.New().String()
		headers.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))
		headers.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		headers.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")

		// 生成纯文本部分
		var textContent bytes.Buffer
		if err := s.textTpl.Execute(&textContent, data); err != nil {
			return "", &notification.NotificationError{
				Message: "生成纯文本邮件内容失败",
				Code:    "TEMPLATE_ERROR",
				Err:     err,
			}
		}
		headers.Write(textContent.Bytes())

		// 生成HTML部分
		headers.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		headers.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")

		var htmlContent bytes.Buffer
		if err := s.htmlTpl.Execute(&htmlContent, data); err != nil {
			return "", &notification.NotificationError{
				Message: "生成HTML邮件内容失败",
				Code:    "TEMPLATE_ERROR",
				Err:     err,
			}
		}
		headers.Write(htmlContent.Bytes())

		// 结束边界
		headers.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	} else {
		// 只生成纯文本邮件
		headers.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")

		var textContent bytes.Buffer
		if err := s.textTpl.Execute(&textContent, data); err != nil {
			return "", &notification.NotificationError{
				Message: "生成纯文本邮件内容失败",
				Code:    "TEMPLATE_ERROR",
				Err:     err,
			}
		}
		headers.Write(textContent.Bytes())
	}

	return headers.String(), nil
}

// GetID 获取发送器ID
func (s *EmailSender) GetID() string {
	return s.id
}

// GetName 获取发送器名称
func (s *EmailSender) GetName() string {
	return s.name
}

// Close 关闭发送器
func (s *EmailSender) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	// 关闭SMTP客户端
	if s.client != nil {
		_ = s.client.Quit()
		s.client = nil
	}

	s.closed = true
	return nil
}
