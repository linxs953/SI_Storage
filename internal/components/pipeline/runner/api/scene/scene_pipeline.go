package scene

import (
	api "Storage/internal/logic/workflows/api/apirunner"
	"Storage/internal/logic/workflows/core"
	"context"
)

// 实现ScenePipeline
// 子任务级别

func NewScenePipeline(name string, description string, apiPipelines []*api.ApiPipeline) *ScenePipeline {
	return &ScenePipeline{}
}

func (s *ScenePipeline) Initialize(ctx context.Context) error {
	// 根据sceneID查询场景关联的apiID
	// 根据apiID查询api
	// 初始化apiPipelines
	return nil
}

func (s *ScenePipeline) Execute(ctx context.Context, spec map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

// Validate 验证pipeline配置
func (s *ScenePipeline) Validate(ctx context.Context) error {
	return nil
}

// Cancel 取消pipeline执行
func (s *ScenePipeline) Cancel(ctx context.Context) error {
	return nil
}

// GetStatus 获取pipeline状态
func (s *ScenePipeline) GetStatus(ctx context.Context) core.TaskStatus {
	return core.TaskStatusPending
}

// GetProgress 获取执行进度
func (s *ScenePipeline) GetProgress(ctx context.Context) (float64, error) {
	return 0, nil
}

// GetMetrics 获取执行指标
func (s *ScenePipeline) GetMetrics(ctx context.Context) map[string]interface{} {
	return nil
}

// Cleanup 清理资源
func (s *ScenePipeline) Cleanup(ctx context.Context) error {
	return nil
}

func (s *ScenePipeline) StartAllApiPipelines(ctx context.Context) error {
	return nil
}

func (s *ScenePipeline) ReceiveMetrics(ctx context.Context, metrics *api.ApiMetrics) error {
	return nil
}

func (s *ScenePipeline) ReportMetrics(ctx context.Context) error {
	return nil
}
