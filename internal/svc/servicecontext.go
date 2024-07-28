package svc

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"lexa-engine/internal/config"
)

type ServiceContext struct {
	Config         config.Config
	RedisClient    *redis.Redis
	KqPusherClient *kq.Pusher
	TaskPushClient *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	conf := redis.RedisConf{
		Host: c.RedisConf.Host,
		Type: c.RedisConf.Type,
		Pass: c.RedisConf.Pass,
		Tls:  c.RedisConf.Tls,
	}
	return &ServiceContext{
		Config:         c,
		RedisClient:    redis.MustNewRedis(conf),
		KqPusherClient: kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic),
		TaskPushClient: kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.TaskRunTopic),
	}
}
