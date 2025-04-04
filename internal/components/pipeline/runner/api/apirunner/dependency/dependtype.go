package dependency

import "encoding/json"

type Dependency struct {
	// 依赖ID
	DependID string `json:"depend_id"`

	// 依赖类型，场景，基础数据，自定义数据，数据生成
	Type string `json:"type"`

	// 依赖项名称
	Name string `json:"name"`

	// 依赖配置
	Key DConfig `json:"key"`

	// 获取依赖之后，存储的数据
	Value interface{} `json:"value"`
}

type DataSourceType string

const (
	DataSourceScene     DataSourceType = "scene"     // DataSourceScene 场景运行时数据
	DataSourceRedis     DataSourceType = "redis"     // DataSourceRedis Redis预设数据
	DataSourceEnv       DataSourceType = "env"       // DataSourceEnv 环境变量
	DataSourceGenerator DataSourceType = "generator" // DataSourceGenerator 数据生成器
	DataSourceCustom    DataSourceType = "custom"    // DataSourceCustom 自定义数据
)

// GeneratorType 生成器类型
type GeneratorType string

const (
	GenTypeTimestamp    GeneratorType = "timestamp"     // GenTypeTimestamp 时间戳
	GenTypeRandomInt    GeneratorType = "random_int"    // GenTypeRandomInt 随机整数
	GenTypeRandomString GeneratorType = "random_string" // GenTypeRandomString 随机字符串
	GenTypeUUID         GeneratorType = "uuid"          // GenTypeUUID UUID
	GenTypeSequence     GeneratorType = "sequence"      // GenTypeSequence 自增序列
	GenTypeCurrentTime  GeneratorType = "current_time"  // GenTypeCurrentTime 当前时间
)

// SceneDataSelector 场景数据选择器
type SceneDataSelector struct {
	SceneID      string      `json:"scene_id"`                // 场景ID
	StepID       string      `json:"step_id"`                 // 步骤ID
	JsonPath     string      `json:"json_path"`               // JsonPath 表达式，用于从响应中提取数据
	DefaultValue interface{} `json:"default_value,omitempty"` // 默认值，当数据不存在时使用

}

// GeneratorConfig 生成器配置
type GeneratorConfig struct {
	Type   GeneratorType          `json:"type"`             // 生成器类型
	Params map[string]interface{} `json:"params,omitempty"` // 生成器参数
	Cache  Cache                  `json:"cache,omitempty"`  // 缓存配置
}

type Cache struct {
	Enable              bool   `json:"enable"`                          // 是否启用缓存
	TTL                 int    `json:"ttl"`                             // 缓存时间（秒）
	KeyPrefix           string `json:"key_prefix,omitempty"`            // 缓存key前缀
	RefreshBeforeExpire bool   `json:"refresh_before_expire,omitempty"` // 是否在TTL过期前刷新缓存
	RefreshTTL          int    `json:"refresh_ttl,omitempty"`           // TTL过期前多少秒刷新缓存
}

type DConfig struct {
	// 数据来源类型
	SourceType    DataSourceType     `json:"source_type"`               // 数据来源类型
	SceneData     *SceneDataSelector `json:"scene_data,omitempty"`      // 场景数据来源配置
	RedisKey      string             `json:"redis_key,omitempty"`       // Redis数据来源配置
	RedisDataType string             `json:"redis_data_type,omitempty"` // Redis数据类型
	RedisField    string             `json:"redis_field,omitempty"`     // Redis Hash/ZSet 的字段名
	EnvName       string             `json:"env_name,omitempty"`        // 环境变量名
	Generator     *GeneratorConfig   `json:"generator,omitempty"`       // 数据生成器配置
	DefaultValue  interface{}        `json:"default_value,omitempty"`   // type=5, 自定义默认值
	Strategy      FetchStrategy      `json:"fetch_strategy,omitempty"`  // 获取数据失败时的处理策略
	Transform     struct {
		EnableTypeConversion bool   `json:"enable_type_conversion"` // 是否需要类型转换
		TargetType           string `json:"target_type,omitempty"`  // 目标类型
		TimeFormat           string `json:"time_format,omitempty"`  // 日期时间格式化模板
	} `json:"transform,omitempty"`
}

type FetchStrategy struct {
	OnFailure     string `json:"on_failure"`               // 当获取数据失败时的处理策略：fail/retry/use_default
	MaxRetries    int    `json:"max_retries,omitempty"`    // 重试次数
	RetryInterval int    `json:"retry_interval,omitempty"` // 重试间隔（秒）
}

// GetValue 获取指定类型的值，如果类型断言成功则返回该类型的值和 true，否则返回零值和 false
func GetValue[T any](d *Dependency) (T, bool) {
	var zero T
	if d.Value == nil {
		return zero, false
	}

	// 处理 map[string]interface{} 的情况
	if mapValue, ok := d.Value.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(mapValue)
		if err != nil {
			return zero, false
		}

		var result T
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return zero, false
		}
		return result, true
	}

	// 直接类型断言
	if value, ok := d.Value.(T); ok {
		return value, true
	}

	return zero, false
}
