package interfaceservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteInterfaceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteInterfaceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteInterfaceLogic {
	return &DeleteInterfaceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteInterfaceLogic) DeleteInterface(in *storage.DeleteInterfaceRequest) (*storage.DeleteResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.DeleteResponse{}, nil
}
