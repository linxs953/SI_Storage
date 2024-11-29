package logic

import (
	"context"

	"Storage/internal/svc"
	"Storage/pb/Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTaskLogic {
	return &DeleteTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除任务
func (l *DeleteTaskLogic) DeleteTask(in *storage.DeleteTaskRequest) (*storage.OperationResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.OperationResponse{}, nil
}
