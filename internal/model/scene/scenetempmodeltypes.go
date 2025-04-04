package scene

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Scenetempmodel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SceneId    string             `bson:"sceneId,omitempty" json:"sceneId,omitempty"`
	SceneName  string             `bson:"sceneName,omitempty" json:"sceneName,omitempty"`
	SceneDesc  string             `bson:"sceneDesc,omitempty" json:"sceneDesc,omitempty"`
	RelatedApi []*RelatedApi      `bson:"relatedApi,omitempty" json:"relatedApi,omitempty"`
	Strategy   *SceneStrategy     `bson:"strategy,omitempty" json:"strategy,omitempty"` // 场景策略
	UpdateAt   time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt   time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
}

type RelatedApi struct {
	ApiId      string `bson:"apiId,omitempty" json:"apiId,omitempty"`
	Name       string `bson:"name,omitempty" json:"name,omitempty"`
	Enabled    bool   `bson:"enabled,omitempty" json:"enabled,omitempty"`
	Dependency string `bson:"dependency,omitempty" json:"dependency,omitempty"`
	Expect     string `bson:"expect,omitempty" json:"expect,omitempty"`
	Extractor  string `bson:"extractor,omitempty" json:"extractor,omitempty"`
}

type SceneStrategy struct {
	Timeout *SceneTimeoutSetting `bson:"timeout,omitempty" json:"timeout,omitempty"` // 超时配置
	Retry   *SceneRetrySetting   `bson:"retry,omitempty" json:"retry,omitempty"`     // 重试配置
}

type SceneTimeoutSetting struct {
	Duration int  `bson:"duration,omitempty" json:"duration,omitempty"` // 超时时间
	Enabled  bool `bson:"enabled,omitempty" json:"enabled,omitempty"`   // 是否启用超时
}

type SceneRetrySetting struct {
	Enabled  bool `bson:"enabled,omitempty" json:"enabled,omitempty"`   // 是否启用重试
	MaxRetry int  `bson:"maxRetry,omitempty" json:"maxRetry,omitempty"` // 最大重试次数
	Interval int  `bson:"interval,omitempty" json:"interval,omitempty"` // 重试间隔
}
