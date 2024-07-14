package mqs

import (
	"context"
	"encoding/json"
	"errors"
	"lexa-engine/internal/logic/sync/synchronizer"
	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/model/mongo/sync_task"
	"lexa-engine/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	var event synchronizer.DataHook
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)

	if err := json.Unmarshal([]byte(val), &event); err != nil {
		logx.Error("[消费者] 解析消费事件失败", err)
		return err
	}
	switch event.Event {
	case "sync_data":
		{
			var data *apidetail.Apidetail
			eventDataBytes, _ := json.Marshal(event.Data)
			if err := json.Unmarshal(eventDataBytes, &data); err != nil {
				logx.Error("event data 反序列化 Apidetail 失败", err)
				return err
			}
			err := insertApiInfo(murl, l.svcCtx.Config.Database.Mongo.UseDb, data)
			if err != nil {
				return err
			}
			if event.IsEof {
				finishEventHook := synchronizer.DataHook{
					Event: "sync_finish",
					Data:  event.UpdateId,
				}
				eventBytes, err := json.Marshal(finishEventHook)
				if err != nil {
					logx.Error("序列化同步任务完成事件失败")
					return nil
				}
				l.svcCtx.KqPusherClient.Push(string(eventBytes))
			}
			return nil
		}
	case "sync_finish":
		{
			eventDataBytes, err := json.Marshal(event.Data)
			if err != nil {
				logx.Error("反序列化 event.Data 失败, eventType=syncFinsh", err)
				return err
			}
			var recordId primitive.ObjectID
			if err := json.Unmarshal(eventDataBytes, &recordId); err != nil {
				return err
			}
			return updateSyncRecord(murl, l.svcCtx.Config.Database.Mongo.UseDb, recordId)
		}
	default:
		{
			return errors.New("不支持的事件类型: " + event.Event)
		}
	}
}

func insertApiInfo(mongoUrl string, useDb string, data *apidetail.Apidetail) error {
	mod := apidetail.NewApidetailModel(mongoUrl, useDb, "ApiInfo")
	existRecord, err := mod.FindByApiId(context.Background(), data.ApiId)
	if err != nil {
		// 查找记录失败,统一按新的记录处理
		if errors.Is(err, mongo.ErrNoDocuments) {
			if err := mod.Insert(context.Background(), data); err != nil {
				logx.Error("api入库失败", err)
				return err
			}
			logx.Infof("api=[%v]入库成功", data.ApiId)
			return nil
		}
		return err
	}
	data.ID = existRecord.ID
	_, err = mod.Update(context.Background(), data)
	return err
}

func updateSyncRecord(mongoUrl string, useDb string, objId primitive.ObjectID) error {
	mod := sync_task.NewSynctaskModel(mongoUrl, useDb, "SyncTask")
	syncRecord, err := mod.FindOne(context.Background(), objId.Hex())
	if err != nil {
		logx.Error(err)
		return err
	}
	if syncRecord.State != 0 {
		logx.Errorf("同步记录[id=%v]已完成,忽略", syncRecord.ID.String())
		return nil
	}
	syncRecord.State = 1
	_, err = mod.UpdateRecord(context.Background(), syncRecord)
	if err != nil {
		logx.Error(err)
		return nil
	}
	logx.Infof("同步api 任务[id=%v]已完成", syncRecord.ID.String())
	return nil
}
