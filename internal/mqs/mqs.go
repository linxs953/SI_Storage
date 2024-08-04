package mqs

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"lexa-engine/internal/config"
	"lexa-engine/internal/logic/syncApi"
	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/svc"
)

func Consumers(c config.Config, ctx context.Context, svcContext *svc.ServiceContext) []service.Service {
	return []service.Service{
		//Listening for changes in consumption flow status
		// kq.MustNewQueue(c.KqConsumerConf, NewApiSyncSuccess(ctx, svcContext)),
		// kq.MustNewQueue(c.TaskConsumerConf, NewTaskSuccess(ctx, svcContext)),
		newRedisConsumer(&c.RedisConf, NewApiSyncSuccess2(ctx, svcContext)),
	}
}

func NewApiSyncSuccess2(ctx context.Context, svcContext *svc.ServiceContext) func([]byte) {
	return func(message []byte) {
		var event syncApi.ApiSyncEvent
		if err := json.Unmarshal(message, &event); err != nil {
			logx.Error(err)
			return
		}
		switch event.Type {
		case "sync_apidetail":
			{
				var apiInfo apidetail.Apidetail
				evtDataBts, err := json.Marshal(event.Data)
				if err != nil {
					logx.Error(err)
					return
				}
				if err := json.Unmarshal(evtDataBts, &apiInfo); err != nil {
					logx.Error(err)
					return
				}
				mgoUrl := mong.GetMongoUrl(svcContext.Config.Database.Mongo)
				useDb := svcContext.Config.Database.Mongo.UseDb
				if err := insertApiInfo(mgoUrl, useDb, &apiInfo); err != nil {
					logx.Error(err)
					return
				}
				logx.Infof("sync api success. %v", apiInfo.ApiId)
				if event.IsEof {
					finishEvent := syncApi.ApiSyncEvent{
						Type:     "sync_finish",
						UpdateId: event.UpdateId,
					}
					feBts, err := json.Marshal(finishEvent)
					if err != nil {
						logx.Error(err)
						return
					}
					if _, err := svcContext.RedisClient.Rpush("SyncApi", string(feBts)); err != nil {
						logx.Error(err)
						return
					}
					logx.Info("Api任务同步完成")
				}
				break
			}
		case "sync_finish":
			{
				mgoUrl := mong.GetMongoUrl(svcContext.Config.Database.Mongo)
				useDb := svcContext.Config.Database.Mongo.UseDb

				oldId := event.UpdateId
				if err := updateSyncRecord(mgoUrl, useDb, oldId); err != nil {
					logx.Error(err)
					return
				}
				break
			}
		default:
			{
				logx.Errorf("不支持的事件类型: %s", event.Type)
				return
			}
		}
	}
}

type redisConsumer struct {
	redisConf *config.RedisConf
	handler   func([]byte)
	rdb       *redis.Redis // 添加Redis连接实例
	stopCh    chan struct{}
}

func (r *redisConsumer) Start() {
	r.stopCh = make(chan struct{}) // 创建停止通道
	var wg sync.WaitGroup          // 使用sync.WaitGroup来等待所有goroutine完成

	rdb := redis.MustNewRedis(redis.RedisConf{
		Host: r.redisConf.Host,
		Pass: r.redisConf.Pass,
		Type: r.redisConf.Type,
	})
	r.rdb = rdb

	// 使用 blpop 命令监听队列
	queueName := "SyncApi"
	for {
		select {
		case <-r.stopCh: // 如果收到停止信号，则退出循环
			return
		default:
			val, err := rdb.Rpop(queueName)
			if err != nil {
				logx.Errorf("Failed to read from Redis: %v", err)
				time.Sleep(time.Second * 5) // 等待一段时间后重试
				continue
			}

			wg.Add(1) // 增加一个goroutine
			go func(val string) {
				defer wg.Done()        // goroutine完成时调用Done
				r.handler([]byte(val)) // 打印消息
			}(val)
		}
		wg.Wait() // 等待所有goroutine完成
	}
}

// Stop implements service.Service.
func (r *redisConsumer) Stop() {
}

func newRedisConsumer(redisConf *config.RedisConf, handler func([]byte)) service.Service {
	return &redisConsumer{
		redisConf: redisConf,
		handler:   handler,
	}
}
