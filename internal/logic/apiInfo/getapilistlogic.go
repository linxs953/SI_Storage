package apiInfo

import (
	"context"
	"encoding/json"
	mgoutil "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"math"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetApiListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetApiListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApiListLogic {
	return &GetApiListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetApiListLogic) GetApiList(req *types.ApiListDto) (resp *types.ApiListResp, err error) {
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	murl := mgoutil.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	mod := apidetail.NewApidetailModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "ApiInfo")
	totalNum, err := mod.GetListCount(context.Background())
	if err != nil {
		return &types.ApiListResp{
			Code:    0,
			Message: err.Error(),
			Data:    nil,
		}, nil
	}

	if totalNum < int64(req.PageSize) {
		req.PageSize = int(totalNum)
	}

	totalPage := int(math.Ceil(float64(totalNum) / float64(req.PageSize)))

	if req.PageNum > totalPage {
		return &types.ApiListResp{
			Code:    0,
			Message: "获取 Api 列表成功",
			Data:    nil,
		}, nil
	}
	apilist, err := mod.FindApiList(context.Background(), int64(req.PageNum), int64(req.PageSize))
	if err != nil {
		return &types.ApiListResp{
			Code:    500,
			Message: err.Error(),
			Data:    nil,
		}, err
	}
	apilistBytes, err := json.Marshal(apilist)
	if err != nil {
		return &types.ApiListResp{
			Code:    500,
			Message: "序列化api 列表失败",
			Data:    nil,
		}, err
	}
	var apilistMap []map[string]interface{}
	if err = json.Unmarshal(apilistBytes, &apilistMap); err != nil {
		return &types.ApiListResp{
			Code:    500,
			Message: "映射 Api 列表失败",
			Data:    nil,
		}, err
	}

	resp = &types.ApiListResp{
		Code:        0,
		Message:     "success",
		CurrentPage: req.PageNum,
		Total:       int(totalPage),
		TotalNum:    int(totalNum),
		Data:        apilistMap,
	}
	return
}
