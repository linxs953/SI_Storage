package interfaceservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInterfaceListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInterfaceListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInterfaceListLogic {
	return &GetInterfaceListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 接口同步
func (l *GetInterfaceListLogic) GetInterfaceList(in *storage.Empty) (*storage.GetInterfaceListResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.GetInterfaceListResponse{}, nil
}
