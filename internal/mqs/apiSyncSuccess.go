package mqs

import (
	"context"
	"encoding/json"
	"lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApiSyncSuccess struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApiSyncSuccess(ctx context.Context, svcCtx *svc.ServiceContext) *ApiSyncSuccess {
	return &ApiSyncSuccess{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApiSyncSuccess) Consume(key, val string) error {
	logx.Infof("Receive message %v", val)
	var record apidetail.Apidetail
	if err := json.Unmarshal([]byte(val), &record); err != nil {
		logx.Error("同步 api detail 序列化Apidetail 失败", err)
		return err
	}

	murl := mongo.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	mod := apidetail.NewApidetailModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "ApiInfo")
	if err := mod.Insert(context.Background(), &record); err != nil {
		logx.Error("api入库失败")
		logx.Error(err)
		return err
	}

	logx.Infof("api=[%v]入库成功", record.ApiId)
	return nil
}
