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

type ListTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTasksLogic {
	return &ListTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListTasksLogic) ListTasks(in *storage.Empty) (*storage.TaskListResponse, error) {
	response := &storage.TaskListResponse{
		Header: &storage.ResponseHeader{
			Code: int64(errors.Success),
		},
	}

	// 1. 查询所有任务
	taskModel := model.NewTaskModel(l.svcCtx.GetMongoURI(), l.svcCtx.Config.Database.Mongo.UseDb, model.TaskCollectionName)
	tasks, err := taskModel.FindAllTask(l.ctx)
	if err != nil {
		logx.Errorf("获取任务列表失败: %v", err)
		response.Header.Code = int64(errors.InternalError)
		response.Header.Message = err.(*errors.Error).WithDetails("获取任务列表失败", err).GetMessage()
		return response, nil
	}

	// 2. 转换为TaskListResponse_TaskItem格式
	taskItems := make([]*storage.TaskListResponse_TaskItem, 0, len(tasks))
	for _, task := range tasks {
		item := &storage.TaskListResponse_TaskItem{
			Meta: &storage.TaskMeta{
				TaskId:   task.TaskId,
				TaskName: task.TaskName,
				TaskDesc: task.TaskDesc,
			},
			Type:     int64(task.Type),
			CreateAt: task.CreateAt.Format(time.RFC3339),
			UpdateAt: task.UpdateAt.Format(time.RFC3339),
		}

		// 处理任务规格
		switch {
		case task.APISpec != nil:
			item.Spec = &storage.TaskListResponse_TaskItem_ApiSpec{
				ApiSpec: convertToAPISpecResponse(task.APISpec),
			}
		case task.SyncSpec != nil:
			item.Spec = &storage.TaskListResponse_TaskItem_SyncSpec{
				SyncSpec: convertToSyncSpecResponse(task.SyncSpec),
			}
		}
		taskItems = append(taskItems, item)
	}

	// 3. 构造最终响应
	response.Data = taskItems
	response.Header.Message = "成功获取任务列表"
	return response, nil
}
