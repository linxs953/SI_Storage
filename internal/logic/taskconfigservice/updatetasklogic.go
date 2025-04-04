package taskconfigservicelogic

import (
	"context"
	"strings"
	"time"

	"Storage/internal/errors"
	model "Storage/internal/model/task"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskLogic {
	return &UpdateTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateTaskLogic) UpdateTask(in *storage.UpdateTaskRequest) (*storage.TaskResponse, error) {
	response := &storage.TaskResponse{
		Header: &storage.ResponseHeader{
			Code: int64(errors.Success),
		},
	}

	// 1. 参数校验
	if err := validateUpdateRequest(in); err != nil {
		response.Header.Code = int64(errors.InvalidParameter)
		response.Header.Message = err.(*errors.Error).GetMessage()
		return response, nil
	}

	// 2. 获取现有任务
	taskModel := model.NewTaskModel(l.svcCtx.GetMongoURI(), l.svcCtx.Config.Database.Mongo.UseDb, model.TaskCollectionName)
	existingTask, err := taskModel.FindOneByTaskID(l.ctx, in.TaskId)
	if err != nil {
		logx.Errorf("查询任务失败: %v", err)
		response.Header.Code = int64(errors.InternalError)
		response.Header.Message = "系统错误，查询任务失败"
		return response, nil
	}
	if existingTask == nil {
		response.Header.Code = int64(errors.DBNotFound)
		response.Header.Message = "任务不存在"
		return response, nil
	}

	// // 3. 类型一致性检查
	// if (existingTask.Type == int32(storage.TaskType_API_TEST) && in.GetApiSpec() == nil) ||
	// 	(existingTask.Type == int32(storage.TaskType_SYNC) && in.GetSyncSpec() == nil) {
	// 	response.Header.Code = storage.StatusCode_BAD_REQUEST
	// 	response.Header.Message = "任务类型与配置不匹配"
	// 	return response, nil
	// }

	// 4. 更新任务信息
	updateFields := model.Task{
		TaskName: in.Name,
		TaskDesc: in.Desc,
		Enable:   existingTask.Enable,
		Version:  existingTask.Version + 1, // 乐观锁版本控制
		UpdateAt: time.Now(),
		APISpec:  existingTask.APISpec,
		SyncSpec: existingTask.SyncSpec,
	}

	// 处理具体配置更新
	switch spec := in.Spec.(type) {
	case *storage.UpdateTaskRequest_ApiSpec:
		if err := validateAPISpec(spec.ApiSpec); err != nil {
			response.Header.Code = int64(errors.InvalidParameter)
			response.Header.Message = err.(*errors.Error).GetMessage()
			return response, nil
		}
		logx.Error(spec.ApiSpec.Strategy)
		updateFields.APISpec = &model.APITaskSpec{
			Scenarios: convertScenarios(spec.ApiSpec.Scenarios),
			Strategy:  convertStrategy(spec.ApiSpec.Strategy),
		}
	case *storage.UpdateTaskRequest_SyncSpec:
		if err := validateSyncSpec(spec.SyncSpec); err != nil {
			response.Header.Code = int64(errors.InvalidParameter)
			response.Header.Message = err.(*errors.Error).GetMessage()
			return response, nil
		}
		updateFields.SyncSpec = &model.SyncTaskSpec{
			SyncType:    spec.SyncSpec.SyncType,
			Source:      spec.SyncSpec.Source,
			Destination: spec.SyncSpec.Destination,
			Strategy:    spec.SyncSpec.Strategy,
		}
	}

	// 5. 持久化更新
	if err := taskModel.UpdateTask(l.ctx, in.TaskId, &updateFields); err != nil {
		logx.Errorf("更新任务失败: %v", err)
		response.Header.Code = int64(errors.InternalError)
		response.Header.Message = "系统错误，更新任务失败"
		return response, nil
	}

	// 6. 构造响应
	updatedTask, _ := taskModel.FindOneByTaskID(l.ctx, in.TaskId)
	return buildTaskResponse(updatedTask), nil
}

// 参数校验函数
func validateUpdateRequest(in *storage.UpdateTaskRequest) error {
	if in == nil {
		return errors.New(errors.ValidationFailed).WithDetails("请求参数为空", nil)
	}

	// 任务ID校验
	if strings.TrimSpace(in.TaskId) == "" {
		return errors.New(errors.InvalidParameter).WithDetails("任务ID不能为空", nil)
	}

	// 名称校验
	if strings.TrimSpace(in.Name) == "" {
		return errors.New(errors.InvalidParameter).WithDetails("任务名称不能为空", nil)
	}
	if len(in.Name) < 3 || len(in.Name) > 50 {
		return errors.New(errors.InvalidParameter).WithDetails("任务名称长度需在3-50字符之间", nil)
	}

	// 配置校验
	switch {
	case in.GetApiSpec() != nil && in.GetSyncSpec() != nil:
		return errors.New(errors.InvalidParameter).WithDetails("只能指定一种任务配置类型", nil)
	case in.GetApiSpec() == nil && in.GetSyncSpec() == nil:
		return errors.New(errors.InvalidParameter).WithDetails("必须指定任务配置", nil)
	}

	return nil
}

// 响应构造复用函数
func buildTaskResponse(task *model.Task) *storage.TaskResponse {
	resp := &storage.TaskResponse{
		Header: &storage.ResponseHeader{
			Code: int64(errors.Success),
		},
		Meta: &storage.TaskMeta{
			TaskId:   task.TaskId,
			TaskName: task.TaskName,
			TaskDesc: task.TaskDesc,
		},
		Type:     int64(task.Type),
		CreateAt: task.CreateAt.Format(time.RFC3339),
		UpdateAt: task.UpdateAt.Format(time.RFC3339),
	}

	switch {
	case task.APISpec != nil:
		resp.Spec = &storage.TaskResponse_ApiSpec{
			ApiSpec: convertToAPISpecResponse(task.APISpec),
		}
	case task.SyncSpec != nil:
		resp.Spec = &storage.TaskResponse_SyncSpec{
			SyncSpec: convertToSyncSpecResponse(task.SyncSpec),
		}
	}

	return resp
}
