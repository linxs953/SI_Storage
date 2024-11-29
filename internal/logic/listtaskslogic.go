package logic

import (
	"context"

	"Storage/internal/svc"
	"Storage/pb/Storage/storage"

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

// 列出所有任务
func (l *ListTasksLogic) ListTasks(in *storage.Empty) (*storage.TaskList, error) {
	// todo: add your logic here and delete this line

	return &storage.TaskList{}, nil
}
