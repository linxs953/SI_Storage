package synchronizer

import (
	"errors"
	"lexa-engine/internal/logic"
	"lexa-engine/internal/svc"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 同步器 job spec
type SyncSpec struct {
	SyncType       string     `json:"syncType"`
	ApiFoxSpec     ApiFoxSpec `json:"apiFoxSpec"`
	ApollConfigUrl string     `json:"apollConfigUrl"`
}

// apifox同步器 spec
type ApiFoxSpec struct {
	// sharedoc/jsonfile
	Type         string `json:"type"`
	ShareUrl     string `json:"shareUrl"`
	ShareDocAuth string `json:"shareDocAuth"`
}

// 同步器接口
type Synchronizer interface {
	Sync(ctx *svc.ServiceContext) (err error)
	Store(ctx *svc.ServiceContext) (err error)
}

/*
解析 apifox detail 数据
*/
type Property struct {
	Type       any            `json:"type"`
	Required   []string       `json:"required"`
	Properties map[string]any `json:"properties"`
	Items      map[string]any `json:"items"`
}

type ApiRequestBody struct {
	Type       string                `json:"type"`
	Parameters []ApiPayloadParameter `json:"parameters"`
	JsonSchema map[string]any        `json:"jsonSchema"`
}

type ApiPayloadParameter struct {
	Description string `json:"description"`
	Enable      bool   `json:"enable"`
	ParamName   string `json:"name"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
}

type ApiFoxTree struct {
	Data []ApiFoxTreeData `json:"data"`
}

type ApiFoxTreeData struct {
	Children []ApiFoxTreeData     `json:"children"`
	Key      string               `json:"key"`
	Name     string               `json:"name"`
	Type     string               `json:"type"`
	Folder   ApiFoxTreeFolderInfo `json:"folder"`
}

type ApiFoxTreeFolderInfo struct {
	Id int `json:"id"`
}

func (as *SyncSpec) Sync(svcCtx *svc.ServiceContext, recordId primitive.ObjectID) error {
	var err error

	if as == nil {
		err = errors.New("同步器对象为空")
		return err
	}
	if as.SyncType == "" || as.ApollConfigUrl == "" {
		err = errors.New("同步器类型或同步配置地址不能为空")
		return err
	}
	switch as.SyncType {
	case logic.APIFOXDOC:
		{
			apifoxSync := BuildApiFox(as.ApiFoxSpec)
			if apifoxSync == nil {
				err = errors.New("构建 ApiSynchronizer 失败")
				return err
			}
			if err = apifoxSync.Sync(svcCtx, recordId); err != nil {
				return err
			}
			break
		}
	default:
		{
			err = errors.New("不支持的同步器类型")
			return err
		}
	}
	return err
}
