package logic

import (
	"context"

	"Storage/internal/svc"
	"Storage/pb/Storage/storage"

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

// 更新任务
func (l *UpdateTaskLogic) UpdateTask(in *storage.UpdateTaskRequest) (*storage.OperationResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.OperationResponse{}, nil
}
