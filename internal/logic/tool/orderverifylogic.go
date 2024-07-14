package tool

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"lexa-engine/internal/svc"
)

type OrderVerifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderVerifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderVerifyLogic {
	return &OrderVerifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderVerifyLogic) OrderVerify() error {
	// todo: add your logic here and delete this line

	return nil
}
