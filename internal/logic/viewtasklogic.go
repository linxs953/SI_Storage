package logic

import (
	"context"

	"Storage/internal/svc"
	"Storage/pb/Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type ViewTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewViewTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ViewTaskLogic {
	return &ViewTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查看单个任务详情
func (l *ViewTaskLogic) ViewTask(in *storage.ViewTaskRequest) (*storage.Task, error) {
	// todo: add your logic here and delete this line

	return &storage.Task{}, nil
}
