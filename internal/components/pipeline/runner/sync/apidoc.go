package sync

import (
	"context"
	"net/http"
)

// 抽象apidoc的同步

type ApiDocSource struct {
	// HTTP客户端
	client *http.Client

	// 状态
	status SourceStatus

	// 输出的API详情
	apiDetails map[string]*APIDetail
}

// APIDetail represents detailed information about an API endpoint
type APIDetail struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Description string                 `json:"description"`
	Headers     map[string]string      `json:"headers"`
	Parameters  []Parameter            `json:"parameters"`
	Responses   interface{}            `json:"responses"`
	RawData     map[string]interface{} `json:"-"`
}

// Parameter represents a request parameter in the API
type Parameter struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type SyncApiDoc interface {
	Sync(ctx context.Context) (*SyncResult, error)
}
