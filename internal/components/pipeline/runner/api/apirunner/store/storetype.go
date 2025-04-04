package store

type ReportRunData struct {
	// 上传数据的类型，1--存储到场景的struct中, 2--存储到db--redis
	Type string `json:"type"`

	// 存储配置
	Config ReportStoreConfig `json:"config"`
}

type ReportStoreConfig struct {
	Scene *SceneStoreConfig `json:"scene"`
	DB    *DBStoreConfig    `json:"db"`
}

type RedisDataStoreType string

const (
	RedisDataStoreTypeString    RedisDataStoreType = "string"
	RedisDataStoreTypeHash      RedisDataStoreType = "hash"
	RedisDataStoreTypeList      RedisDataStoreType = "list"
	RedisDataStoreTypeSet       RedisDataStoreType = "set"
	RedisDataStoreTypeSortedSet RedisDataStoreType = "zset"
)

type SceneStoreConfig struct {
	ExecuteID string                 `json:"execute_id"`
	TaskID    string                 `json:"task_id"`
	SceneID   string                 `json:"scene_id"`
	StepID    string                 `json:"step_id"`
	Data      map[string]interface{} `json:"data"`
	MsgChan   chan<- interface{}     `json:"msg_chan"`
}

type DBStoreConfig struct {
	ExecuteID     string                 `json:"execute_id"`
	TaskID        string                 `json:"task_id"`
	SceneID       string                 `json:"scene_id"`
	StepID        string                 `json:"step_id"`
	Data          map[string]interface{} `json:"data"`
	DataStoreType RedisDataStoreType     `json:"data_store_type"`
	Redis         *RedisStoreConfig      `json:"redis"`
}

type RedisStoreConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}
