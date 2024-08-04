package logic

import "time"

// Action represents an action with its dependencies, output, and expected results.
type Action struct {
	ActionID     string            `json:"actionId"`
	SearchKey    string            `json:"searchKey"`
	ActionName   string            `json:"actionName"`
	ActionPath   string            `json:"actionPath"`
	ActionMethod string            `json:"actionMethod"`
	RelateId     int               `json:"relateId"`
	EnvKey       string            `json:"envKey"`
	DomainKey    string            `json:"domainKey"`
	Headers      map[string]string `json:"headers"`
	Dependency   []Dependency      `json:"dependency"`
	Output       Output            `json:"output"`
	Expect       Expect            `json:"expect"`
	Retry        int               `json:"retry"`
	Timeout      int               `json:"timeout"`
}

// Dependency represents the dependencies of an action.
type Dependency struct {
	Type      string `json:"type"`
	ActionKey string `json:"actionKey"`
	DataKey   string `json:"dataKey"`
	Refer     Refer  `json:"refer"`
}

// Refer represents references within a dependency.
type Refer struct {
	Type     string `json:"type"`
	Target   string `json:"target"`
	DataType string `json:"dataType"`
	// Field    string `json:"field"`
	// Match    string `json:"match"`
	// Location string `json:"location"`
}

// Output represents the output of an action.
type Output struct {
	// Event     string      `json:"event"`
	// EventBody []EventBody `json:"meta"`
	Key string `json:"key"`
}

// Meta represents metadata of an action output.
type EventBody struct {
	FieldName  string `json:"fieldName"`
	FieldValue string `json:"fieldValue"`
	DataType   string `json:"dataType"`
}

// Expect defines the expected outcomes of an action.
type Expect struct {
	Sql Sql   `json:"sql"`
	Api []Api `json:"api"`
}

// Sql represents SQL related expectations.
type Sql struct {
	Sql    string   `json:"sql"`
	Table  string   `json:"table"`
	Verify []Verify `json:"verify"`
}

// Verify represents verification details for SQL expectations.
type Verify struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// Api represents API related expectations.
type Api struct {
	Type string `json:"type"`
	Data Data   `json:"data"`
}

// Data holds the detailed data for an API expectation.
type Data struct {
	Name      string      `json:"name"`
	Operation string      `json:"operation"`
	Type      string      `json:"type"`
	Desire    interface{} `json:"desire"`
}

// Scene represents a scenario including actions, results, and webhook details.
type Scene struct {
	Description string   `json:"description"`
	SceneName   string   `json:"sceneName"`
	Type        string   `json:"type"`
	Author      string   `json:"author"`
	SceneId     string   `json:"sceneId"`
	Retry       int      `json:"retry"`
	Timeout     int      `json:"timeout"`
	SearchKey   string   `json:"searchKey"`
	EnvKey      string   `json:"envKey"`
	Actions     []Action `json:"actions"`
}

// Result represents the result of an action within a scene.
type Result struct {
	ActionId int    `json:"actionId"`
	ApiId    string `json:"apiId"`
	ApiName  string `json:"apiName"`
	Request  string `json:"request"`
	Duration int    `json:"duration"`
	Error    string `json:"error"`
	Resp     string `json:"resp"`
}

// WebHook represents the webhook details within a scene.
type WebHook struct {
	HookAddr string `json:"hookAddr"`
	HookType string `json:"hookType"`
}

// 定义 kafka 推送事件 struct
type TaskEvent struct {
	EventType string   `json:"eventType"`
	EventMsg  EventMsg `json:"meta"`
}

type EventMsg struct {
	RequestID string
	TaskID    string
	Total     int
	Execute   int
	StartAt   time.Time
	FinishAt  time.Time
	Duration  time.Duration
	State     int
}

type DependencyReference struct {
	Type     string `yaml:"type"`     // payload, path, header之一
	Target   string `yaml:"target"`   // 引用的目标，如header.Authorization
	DataType string `yaml:"dataType"` // 数据类型
}

// Dependency 表示动作之间的依赖关系
type Dependency1 struct {
	Type      string                `yaml:"type"`      // 内部/外部
	ActionKey string                `yaml:"actionKey"` // 依赖的action名称
	DataKey   string                `yaml:"dataKey"`   // 依赖的action输出的key
	Refer     []DependencyReference `yaml:"refer"`     // 引用详情
}

// OutputField 表示动作的输出
type OutputField struct {
	Key   string            `yaml:"key"`   // 输出字段名
	Value map[string]string `yaml:"value"` // 输出的具体值
}

// ExpectationAPIField 期望的API响应字段验证
type ExpectationAPIField struct {
	Type       string      `yaml:"type"`      // field 或 api
	FieldOrAPI string      `yaml:"fieldName"` // 字段名称或api整体
	Operation  string      `yaml:"operation"` // 验证操作，如eq
	DataType   string      `yaml:"dataType"`  // 数据类型
	Desire     interface{} `yaml:"desire"`    // 期望值，可以是多种类型
}

// Expectation 定义了动作执行后的预期结果
type Expectation struct {
	API []ExpectationAPIField `yaml:"api"`
}

// ActionConfig 单个动作的配置
type ActionConfig struct {
	ActionName string        `yaml:"actionName"`
	ApiID      string        `yaml:"apiID"`
	ActionID   string        `yaml:"actionID"`
	Retry      int           `yaml:"retry"`
	Timeout    int           `yaml:"timeout"` // 注意：此处的timeout单位应与yaml文件中一致，为秒
	Dependency []Dependency1 `yaml:"dependency"`
	Output     []OutputField `yaml:"output"`
	Expect     Expectation   `yaml:"expect"`
}

// SceneConfig 定义整个场景的配置
type SceneConfig struct {
	Description string         `yaml:"description"`
	Type        string         `yaml:"type"`
	Author      string         `yaml:"author"`
	SceneID     string         `yaml:"sceneId"`
	Total       int            `yaml:"total"`
	Timeout     int            `yaml:"timeout"` // 场景整体的超时时间
	Actions     []ActionConfig `yaml:"actions"`
}
