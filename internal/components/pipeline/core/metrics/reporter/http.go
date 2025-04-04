package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// HTTPReporter HTTP指标上报器
type HTTPReporter struct {
	BaseReporter
	// 上报URL
	Endpoint string
	// HTTP方法
	Method string
	// 认证令牌（如需要）
	AuthToken string
	// HTTP客户端
	client *http.Client
	// 额外的请求头
	headers map[string]string
}

// NewHTTPReporter 创建新的HTTP上报器
func NewHTTPReporter(endpoint, method, authToken string) *HTTPReporter {
	return &HTTPReporter{
		BaseReporter: BaseReporter{name: "http"},
		Endpoint:    endpoint,
		Method:      method,
		AuthToken:   authToken,
		client:      &http.Client{Timeout: 10 * time.Second},
		headers:     make(map[string]string),
	}
}

// AddHeader 添加HTTP请求头
func (r *HTTPReporter) AddHeader(key, value string) {
	r.headers[key] = value
}

// Report 通过HTTP上报指标
func (r *HTTPReporter) Report(ctx context.Context, metrics map[string]interface{}) error {
	// 序列化指标数据
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("序列化指标数据失败: %w", err)
	}
	
	// 创建请求
	req, err := http.NewRequestWithContext(
		ctx, 
		r.Method, 
		r.Endpoint, 
		bytes.NewBuffer(data),
	)
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if r.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.AuthToken)
	}
	
	// 添加自定义请求头
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}
	
	// 发送请求
	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP请求失败: %s", resp.Status)
	}
	
	log.Printf("[指标] 已上报到: %s, 状态: %s", r.Endpoint, resp.Status)
	return nil
}

// Close 实现MetricsReporter接口
func (r *HTTPReporter) Close() error {
	// 关闭HTTP客户端
	r.client.CloseIdleConnections()
	return nil
}

// NewApiFoxReporter 创建用于ApiFox的指标上报器
func NewApiFoxReporter(authToken string) *HTTPReporter {
	reporter := NewHTTPReporter(
		"https://apifox.com/api/v1/metrics", 
		"POST", 
		authToken,
	)
	reporter.AddHeader("Content-Type", "application/json")
	reporter.AddHeader("X-Source", "pipeline-metrics")
	return reporter
}
