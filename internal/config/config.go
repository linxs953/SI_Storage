package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type MongoConfig struct {
	MongoHost   string `json:"mongoHost"`
	MongoPort   int    `json:"mongoPort"`
	MongoUser   string `json:"mongoUser"`
	MongoPasswd string `json:"mongoPasswd"`
	UseDb       string `json:"database"`
}

type DatabaseConfig struct {
	Mongo MongoConfig `json:"mongo"`
}

type RedisConf struct {
	Host string
	Type string `json:",default=node,options=node|cluster"`
	Pass string `json:",optional"`
	Tls  bool   `json:",optional"`
}

type Config struct {
	zrpc.RpcServerConf
	RedisConf        RedisConf `json:"RedisConf"`
	KqConsumerConf   kq.KqConf
	TaskConsumerConf kq.KqConf
	KqPusherConf     struct {
		Brokers      []string
		Topic        string
		TaskRunTopic string
	}
	Database DatabaseConfig `json:"Database"`
}

type KafkaConfig struct {
	rest.RestConf
	KqConsumerConf kq.KqConf
	KqPusherConf   struct {
		Brokers []string
		Topic   string
	}
}
