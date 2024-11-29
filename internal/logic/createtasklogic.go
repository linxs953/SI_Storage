package logic

import (
	"Storage/internal/svc"
	"Storage/pb/Storage/storage"
	"context"
	"time"

	"github.com/google/uuid"
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

func (l *CreateTaskLogic) CreateTask(in *storage.CreateTaskRequest) (*storage.CreateTaskResponse, error) {
	taskID := uuid.New().String()
	now := time.Now().Unix()

	// 这里添加您的业务逻辑，比如保存到数据库
	task := &storage.Task{
		Id:          taskID,
		Title:       in.Title,
		Description: in.Description,
		Status:      storage.Task_PENDING,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   in.CreatedBy,
	}
	_ = task

	// TODO: 保存task到存储层

	return &storage.CreateTaskResponse{
		Response: &storage.OperationResponse{
			Code:    200,
			Success: true,
			Message: "任务创建成功",
		},
		TaskId: taskID,
	}, nil
}
