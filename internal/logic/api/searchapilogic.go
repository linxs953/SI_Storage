package api

import (
	"context"
	"fmt"

	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchApiLogic {
	return &SearchApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchApiLogic) SearchApi(req *types.SearchDto) (resp *types.ApiSearchResp, err error) {
	resp = &types.ApiSearchResp{}
	if req.Keyword == "" {
		resp.Code = 1
		resp.Message = "搜索关键字为空"
		resp.Data = nil
		return
	}
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	apiMod := apidetail.NewApidetailModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "ApiInfo")
	var apis []apidetail.Apidetail
	var apisMap []map[string]interface{}

	if apis, err = apiMod.FindMatch(context.Background(), req.Keyword); err != nil {
		logx.Error(err)
		return nil, err
	}
	for _, api := range apis {
		apiMap := make(map[string]interface{})
		apiMap["id"] = api.ApiId
		apiMap["name"] = api.ApiName
		apiMap["fullName"] = fmt.Sprintf("%s-%s", api.FolderName, api.ApiName)
		apiMap["path"] = api.ApiPath
		apiMap["method"] = api.ApiMethod
		apisMap = append(apisMap, apiMap)
	}

	resp.Code = 0
	resp.Message = "搜索成功"
	resp.Data = apisMap
	return
}
