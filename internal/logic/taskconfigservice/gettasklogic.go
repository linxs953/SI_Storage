package taskconfigservicelogic

import (
	"context"
	"time"

	"Storage/internal/errors"
	model "Storage/internal/model/task"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskLogic {
	return &GetTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTaskLogic) GetTask(in *storage.GetTaskRequest) (*storage.TaskResponse, error) {
	response := &storage.TaskResponse{
		Header: &storage.ResponseHeader{
			Code:    int64(errors.Success),
			Message: "获取任务信息成功",
		},
	}

	// 1. 参数校验
	if in.TaskId == "" {
		response.Header.Code = int64(errors.InvalidParameter)
		response.Header.Message = "任务ID不能为空"
		return response, nil
	}

	// 2. 查询任务信息
	taskModel := model.NewTaskModel(l.svcCtx.GetMongoURI(), l.svcCtx.Config.Database.Mongo.UseDb, model.TaskCollectionName)
	task, err := taskModel.FindOneByTaskID(l.ctx, in.TaskId)
	if err != nil {
		response.Header.Code = int64(errors.InternalError)
		response.Header.Message = err.(*errors.Error).WithDetails("获取任务信息失败", err).GetMessage()
		return response, nil
	}

	// 3. 构造响应数据
	response.Meta = &storage.TaskMeta{
		TaskId:   task.TaskId,
		TaskName: task.TaskName,
		TaskDesc: task.TaskDesc,
	}
	response.Type = int64(task.Type)
	response.CreateAt = task.CreateAt.Format(time.RFC3339)
	response.UpdateAt = task.UpdateAt.Format(time.RFC3339)

	// 4. 填充任务配置详情
	switch {
	case task.APISpec != nil:
		response.Spec = &storage.TaskResponse_ApiSpec{
			ApiSpec: convertToAPISpecResponse(task.APISpec),
		}
	case task.SyncSpec != nil:
		response.Spec = &storage.TaskResponse_SyncSpec{
			SyncSpec: convertToSyncSpecResponse(task.SyncSpec),
		}
	}

	response.Header.Message = "获取任务成功"
	response.Header.Code = int64(errors.Success)
	logx.Error(response.Header.Code)
	return response, nil
}
