package task

import (
	"context"

	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPreActionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPreActionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPreActionsLogic {
	return &GetPreActionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPreActionsLogic) GetPreActions(req *types.GetPreActionsDto) (resp *types.GetPreActionsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
