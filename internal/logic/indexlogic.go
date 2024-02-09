package logic

import (
	"context"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexLogic {
	return &IndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IndexLogic) Index() (resp *types.IndexVO, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.IndexVO)
	resp.Message = "hello, wen_lin_project"
	//if err := l.svcCtx.NsqProducerClient.SendMsg("task","hello world");err != nil {
	//	logc.Error(context.Background(),err)
	//}
	return
}
