package syncTask

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"lexa-engine/internal/svc"
)

type UpdateSyncRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateSyncRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSyncRecordLogic {
	return &UpdateSyncRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSyncRecordLogic) UpdateSyncRecord() error {
	// todo: add your logic here and delete this line

	return nil
}
