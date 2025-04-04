package api

import (
	"Storage/internal/logic/workflows/api/apirunner/dependency"
	"Storage/internal/logic/workflows/api/apirunner/expect"
	"Storage/internal/logic/workflows/api/apirunner/extract"
	"Storage/internal/logic/workflows/api/apirunner/store"
	"Storage/internal/logic/workflows/core"
	"context"
)

// ApiPipeline的接口定义

type ApiRunnable interface {
	// 准备依赖数据
	PrepareDependencies(ctx context.Context, dependencies []dependency.Dependency) (map[string]interface{}, error)

	// 构建请求
	BuildRequest(ctx context.Context, api *ApiDefinition, dependencies map[string]interface{}) (map[string]interface{}, error)

	// 执行请求
	ExecuteRequest(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error)

	// 验证响应
	ValidateResponse(ctx context.Context, response map[string]interface{}, assertions *expect.AssertionGroup) (*expect.AssertionGroupResult, error)

	// 从响应中提取数据
	ExtractData(ctx context.Context, response map[string]interface{}, extractors []extract.Extractor) (map[string]interface{}, error)

	// 将运行数据存储到公共地方，给其他场景使用
	StoreData(ctx context.Context, data map[string]interface{}, storeConfig []*store.ReportStoreConfig) error

	// 上报指标
	ReportMetrics(ctx context.Context, metrics *ApiMetrics, config *ReportConfig) error
}

// ApiRunner API执行器接口
type ApiRunner interface {
	core.PipelineRunner
	ApiRunnable
}
