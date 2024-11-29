package logic

import (
	"context"

	"Storage/internal/svc"
	"Storage/pb/Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchDeleteTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchDeleteTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchDeleteTasksLogic {
	return &BatchDeleteTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量删除任务
func (l *BatchDeleteTasksLogic) BatchDeleteTasks(in *storage.BatchDeleteTasksRequest) (*storage.OperationResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.OperationResponse{}, nil
}
