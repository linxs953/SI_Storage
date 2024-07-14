package syncTask

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"lexa-engine/internal/svc"
)

type GetSyncRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSyncRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSyncRecordLogic {
	return &GetSyncRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSyncRecordLogic) GetSyncRecord() error {
	// todo: add your logic here and delete this line

	return nil
}
