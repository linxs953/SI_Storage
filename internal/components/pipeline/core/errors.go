package core

import "time"

type PipelineError struct {
	// Error message and code
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`

	// Additional context
	Context map[string]interface{} `json:"context,omitempty"` // 请求参数、响应数据等上下文信息

	// Time information
	Timestamp time.Time `json:"timestamp"`

	// Original error if wrapped
	Cause error `json:"cause,omitempty"`
}
