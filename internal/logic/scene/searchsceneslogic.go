package scene

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

type SearchScenesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchScenesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchScenesLogic {
	return &SearchScenesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchScenesLogic) SearchScenes(req *types.SearchScenesDto) (resp *types.SearchScentVo, err error) {
	
	return
}
