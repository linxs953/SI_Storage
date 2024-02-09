package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/rest"
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
type Config struct {
	rest.RestConf
	KqConsumerConf kq.KqConf
	KqPusherConf   struct {
		Brokers []string
		Topic   string
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
