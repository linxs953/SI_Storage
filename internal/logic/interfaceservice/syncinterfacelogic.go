package interfaceservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncInterfaceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncInterfaceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncInterfaceLogic {
	return &SyncInterfaceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SyncInterfaceLogic) SyncInterface(in *storage.SyncInterfaceRequest) (*storage.SyncInterfaceResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.SyncInterfaceResponse{}, nil
}
