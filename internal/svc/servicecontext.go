package svc

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"Storage/internal/config"
	"Storage/internal/model/api"
	"Storage/internal/model/scene"
)

type ServiceContext struct {
	Config      config.Config
	RedisClient *redis.Redis
	// MongoClient    *mongo.Client
	KqPusherClient *kq.Pusher
	TaskPushClient *kq.Pusher
	SceneTemplateModel func() (scene.ScenetempmodelModel, error)
	ApiModel api.ApiModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%d", c.Database.Mongo.MongoUser, c.Database.Mongo.MongoPasswd, c.Database.Mongo.MongoHost, c.Database.Mongo.MongoPort))
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	// 显式检查连接是否成功
	err = client.Ping(context.Background(), nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to ping MongoDB: %v", err))
	}

	// 初始化 ApiModel
	apiModel := api.NewApiModel(
		fmt.Sprintf("mongodb://%s:%s@%s:%d",
			c.Database.Mongo.MongoUser,
			c.Database.Mongo.MongoPasswd,
			c.Database.Mongo.MongoHost,
			c.Database.Mongo.MongoPort,
		),
		c.Database.Mongo.UseDb,
		api.ApiCollectionName,
	)

	// 初始化 SceneTemplateModel
	sceneTemplateModelFunc := func() (scene.ScenetempmodelModel, error) {
		return scene.NewScenetempmodelModel(
			fmt.Sprintf("mongodb://%s:%s@%s:%d",
				c.Database.Mongo.MongoUser,
				c.Database.Mongo.MongoPasswd,
				c.Database.Mongo.MongoHost,
				c.Database.Mongo.MongoPort,
			),
			c.Database.Mongo.UseDb,
			"scene_template", // 集合名称
		), nil
	}

	return &ServiceContext{
		Config: c,
		// MongoClient: client,
		SceneTemplateModel: sceneTemplateModelFunc,
		ApiModel: apiModel,
	}
}

// 生成MongoDB连接URI
func (svc *ServiceContext) GetMongoURI() string {
	c := svc.Config
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin",
		c.Database.Mongo.MongoUser,
		c.Database.Mongo.MongoPasswd,
		c.Database.Mongo.MongoHost,
		c.Database.Mongo.MongoPort,
		c.Database.Mongo.UseDb,
	)
}
