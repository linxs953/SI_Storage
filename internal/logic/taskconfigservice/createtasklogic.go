package taskconfigservicelogic

import (
	"context"
	"strings"
	"time"

	model "Storage/internal/model/task"
	"Storage/internal/svc"
	"Storage/storage"

	"Storage/internal/errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTaskLogic {
	return &CreateTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateTaskLogic) CreateTask(in *storage.CreateTaskRequest) (*storage.TaskResponse, error) {
	cresponse := &storage.TaskResponse{
		Header: &storage.ResponseHeader{
			Code: int64(errors.Success),
		},
	}

	// 1. 参数校验
	if err := validateCreateRequest(in); err != nil {
		cresponse.Header.Code = int64(errors.InvalidParameter)
		cresponse.Header.Message = err.(*errors.Error).GetMessage()
		return cresponse, nil
	}

	// 2. 创建任务基础信息
	task := &model.Task{
		TaskName: in.Name,
		TaskDesc: in.Desc,
		Type:     int32(in.Type),
		Version:  1,
		Enable:   true,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}

	// 3. 根据任务类型设置对应的配置
	switch {
	case in.GetApiSpec() != nil:
		// API测试任务
		apiSpec := in.GetApiSpec()
		task.APISpec = &model.APITaskSpec{
			Scenarios: convertScenarios(apiSpec.Scenarios),
			Strategy:  convertStrategy(apiSpec.Strategy),
		}
	case in.GetSyncSpec() != nil:
		// 同步任务
		syncSpec := in.GetSyncSpec()
		task.SyncSpec = &model.SyncTaskSpec{
			SyncType:    syncSpec.SyncType,
			Source:      syncSpec.Source,
			Destination: syncSpec.Destination,
			Strategy:    syncSpec.Strategy,
		}
	default:
		cresponse.Header.Code = int64(errors.InvalidParameter)
		cresponse.Header.Message = "无效的任务配置类型"
		return cresponse, nil
	}

	// 4. 持久化任务
	taskModel := model.NewTaskModel(l.svcCtx.GetMongoURI(), l.svcCtx.Config.Database.Mongo.UseDb, model.TaskCollectionName)
	if err := taskModel.InsertTask(l.ctx, task); err != nil {
		logx.Errorf("任务创建失败: %v", err)
		cresponse.Header.Code = int64(errors.InternalError)
		// cresponse.Header.Message = formatModelError(err).Error()
		cresponse.Header.Message = err.(*errors.Error).GetMessage()
		return cresponse, nil
	}

	// 5. 返回响应
	cresponse.Meta = &storage.TaskMeta{
		TaskId:   task.TaskId,
		TaskName: task.TaskName,
		TaskDesc: task.TaskDesc,
	}
	switch {
	case task.APISpec != nil:
		cresponse.Spec = &storage.TaskResponse_ApiSpec{ApiSpec: convertToAPISpecResponse(task.APISpec)}
	case task.SyncSpec != nil:
		cresponse.Spec = &storage.TaskResponse_SyncSpec{SyncSpec: convertToSyncSpecResponse(task.SyncSpec)}
	}
	cresponse.Type = int64(task.Type)
	cresponse.CreateAt = task.CreateAt.Format(time.RFC3339)
	cresponse.UpdateAt = task.UpdateAt.Format(time.RFC3339)
	cresponse.Header.Message = "创建任务成功"
	return cresponse, nil

}

// 参数校验
func validateCreateRequest(in *storage.CreateTaskRequest) error {
	if in == nil {
		return errors.New(errors.ValidationFailed).WithDetails("请求参数为空 nil", nil)
	}

	// 校验任务名称
	if strings.TrimSpace(in.Name) == "" {
		return errors.New(errors.InvalidParameter).WithDetails("任务名称不能为空", nil)
		// return errors.New("任务名称不能为空")
	}
	if len(in.Name) < 3 || len(in.Name) > 50 {
		return errors.New(errors.InvalidParameter).WithDetails("任务名称长度需在3-50字符之间", nil)
		// return errors.New("任务名称长度需在3-50字符之间")
	}

	// 校验任务类型
	if in.Type == storage.TaskType_TASK_TYPE_UNSPECIFIED {
		return errors.New(errors.InvalidParameter).WithDetails("请指定任务类型", nil)
	}

	// 校验任务配置
	switch {
	case in.GetApiSpec() != nil:
		return validateAPISpec(in.GetApiSpec())
	case in.GetSyncSpec() != nil:
		return validateSyncSpec(in.GetSyncSpec())
	default:
		return errors.New(errors.InvalidParameter).WithDetails("请指定任务配置", nil)
		// return errors.New("请指定任务配置")
	}
}

// 转换场景配置
func convertScenarios(scenarios []*storage.Scenarios) []model.ScenarioRef {
	if len(scenarios) == 0 {
		return nil
	}

	refs := make([]model.ScenarioRef, len(scenarios))
	for i, s := range scenarios {
		refs[i] = model.ScenarioRef{
			ID:   s.Scid,
			Name: s.Scname,
		}
	}
	return refs
}

// 转换任务策略
func convertStrategy(s *storage.Strategy) model.TaskStrategy {
	if s == nil {
		return model.TaskStrategy{}
	}

	return model.TaskStrategy{
		Timeout: &model.TimeoutSetting{
			Enabled:  true,
			Duration: time.Duration(s.Timeout) * time.Second,
		},
		Retry: &model.RetrySetting{
			Enabled:     s.RetryCount > 0,
			MaxAttempts: int(s.RetryCount),
			Interval:    time.Duration(s.RetryInterval) * time.Second,
		},
		AutoExecute: &model.AutoExecuteSetting{
			Enabled: s.Auto,
			Cron:    s.CronExpression,
		},
	}
}

// 验证 API 任务配置
func validateAPISpec(spec *storage.TaskAPISpec) error {
	if spec == nil {
		return errors.New(errors.InvalidParameter).WithDetails("API任务配置不能为空", nil)
		// return errors.New("API任务配置不能为空")
	}

	if len(spec.Scenarios) == 0 {
		return errors.New(errors.InvalidParameter).WithDetails("API测试场景不能为空", nil)
		// return errors.New("API测试场景不能为空")
	}

	if spec.Strategy == nil {
		return errors.New(errors.InvalidParameter).WithDetails("API任务策略不能为空", nil)
		// return errors.New("API任务策略不能为空")
	}

	return nil
}

// 验证同步任务配置
func validateSyncSpec(spec *storage.TaskSyncSpec) error {
	if spec == nil {
		return errors.New(errors.InvalidParameter).WithDetails("同步任务配置不能为空", nil)
		// return errors.New("同步任务配置不能为空")
	}

	if spec.Source == nil || len(spec.Source) == 0 {
		return errors.New(errors.InvalidParameter).WithDetails("同步源数组为空", nil)
	}

	if spec.Destination == nil || len(spec.Destination) == 0 {
		return errors.New(errors.InvalidParameter).WithDetails("同步存储位置为空", nil)
	}

	if spec.Strategy == nil {
		return errors.New(errors.InvalidParameter).WithDetails("同步任务策略不能为空", nil)
	}

	return nil
}

// 转换为 API 任务响应
func convertToAPISpecResponse(spec *model.APITaskSpec) *storage.TaskAPISpec {
	if spec == nil {
		return nil
	}

	return &storage.TaskAPISpec{
		Scenarios: convertToScenariosResponse(spec.Scenarios),
		Strategy:  convertToStrategyResponse(&spec.Strategy),
	}
}

// 转换为同步任务响应
func convertToSyncSpecResponse(spec *model.SyncTaskSpec) *storage.TaskSyncSpec {
	if spec == nil {
		return nil
	}

	return &storage.TaskSyncSpec{
		SyncType:    spec.SyncType,
		Source:      spec.Source,
		Destination: spec.Destination,
		Strategy:    spec.Strategy,
	}
}

// 转换场景响应
func convertToScenariosResponse(scenarios []model.ScenarioRef) []*storage.Scenarios {
	if len(scenarios) == 0 {
		return nil
	}

	result := make([]*storage.Scenarios, len(scenarios))
	for i, s := range scenarios {
		result[i] = &storage.Scenarios{
			Scid:   s.ID,
			Scname: s.Name,
		}
	}
	return result
}

// 转换策略响应
func convertToStrategyResponse(s *model.TaskStrategy) *storage.Strategy {
	if s == nil {
		return nil
	}

	return &storage.Strategy{
		Auto:           s.AutoExecute.Enabled,
		CronExpression: s.AutoExecute.Cron,
		RetryCount:     int32(s.Retry.MaxAttempts),
		RetryInterval:  int32(s.Retry.Interval.Seconds()),
		Timeout:        int32(s.Timeout.Duration.Seconds()),
	}
}

// 错误处理转换
// func formatModelError(err error) error {
// 	switch {
// 	case strings.Contains(err.Error(), "任务名称已存在"):
// 		return fmt.Errorf("任务名称已被使用")
// 	case strings.Contains(err.Error(), "任务ID已存在"):
// 		return fmt.Errorf("系统错误：任务ID冲突")
// 	default:
// 		return fmt.Errorf("系统繁忙，请稍后重试")
// 	}
// }

// 辅助函数：获取整型参数
// func getIntParam(params map[string]string, key string, defaultValue int) (int, error) {
// 	if val, ok := params[key]; ok {
// 		var result int
// 		_, err := fmt.Sscanf(val, "%d", &result)
// 		if err != nil {
// 			return 0, fmt.Errorf("参数%s格式错误", key)
// 		}
// 		return result, nil
// 	}
// 	return defaultValue, nil
// }

// 辅助函数：获取布尔型参数
// func getBoolParam(params map[string]string, key string, defaultValue bool) bool {
// 	if val, ok := params[key]; ok {
// 		return strings.ToLower(val) == "true"
// 	}
// 	return defaultValue
// }

// 序列化场景配置
// func marshalScenarios(scenarios []model.ScenarioRef) string {
// 	b, _ := json.Marshal(scenarios)
// 	return string(b)
// }
