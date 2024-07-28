package api

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	mgoutil "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

type GetApiDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetApiDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApiDetailLogic {
	return &GetApiDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetApiDetailLogic) GetApiDetail(req *types.ApiDetailDto) (resp *types.ApiDetailResp, err error) {
	murl := mgoutil.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	mod := apidetail.NewApidetailModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "ApiInfo")
	idInt, err := strconv.ParseInt(req.ApiId, 10, 64)
	if err != nil {
		return nil, err
	}
	data, err := mod.FindByApiId(context.Background(), int(idInt))
	if err != nil {
		return nil, err
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	dataMap := make(map[string]interface{})
	if err = json.Unmarshal(dataBytes, &dataMap); err != nil {
		return nil, err
	}
	resp = &types.ApiDetailResp{
		Code:    0,
		Message: "获取Api信息成功",
		Data:    dataMap,
	}
	return
}
