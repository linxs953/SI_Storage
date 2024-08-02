package model

import (
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskRecord struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TaskType        string             `bson:"type,omitempty" json:"type,omitempty"`
	SceneID         string             `bson:"sceneId,omitempty" json:"sceneId,omitempty"`
	RequestID       string             `bson:"requestId,omitempty" json:"requestId,omitempty"`
	Total           int                `bson:"total" json:"total"`
	SuccessCount    int                `bson:"successCount" json:"successCount"`
	FailedCount     int                `bson:"failedCount" json:"failedCount"`
	HasRun          int                `bson:"hasRun" json:"hasRun"`
	FinishAt        time.Time          `bson:"finishAt,omitempty" json:"finishAt,omitempty"`
	State           int                `bson:"state" json:"state"`
	Duration        time.Duration      `bson:"duration,omitempty" json:"duration,omitempty"`
	ActionRunDetail []ActionRunDetail  `bson:"actionRunDetail,omitempty" json:"actionRunDetail,omitempty"`
	UpdateAt        time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt        time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
}

type ActionRunDetail struct {
	ActionID     string            `bson:"actionId,omitempty" json:"actionId,omitempty"`
	ApiID        int               `bson:"apiId,omitempty" json:"apiId,omitempty"`
	Method       string            `bson:"method,omitempty" json:"method,omitempty"`
	ReqURL       string            `bson:"reqUrl,omitempty" json:"reqUrl,omitempty"`
	Headers      map[string]string `bson:"headers,omitempty" json:"headers,omitempty"`
	Payload      interface{}       `bson:"payload,omitempty" json:"payload,omitempty"`
	RetryTimes   int               `bson:"retryTimes,omitempty" json:"retryTimes,omitempty"`
	State        int               `bson:"state,omitempty" json:"state,omitempty"`
	Response     interface{}       `bson:"response,omitempty" json:"response,omitempty"`
	VerifyResult interface{}       `bson:"verifyResult,omitempty" json:"verifyResult,omitempty"`
	Error        interface{}       `bson:"error,omitempty" json:"error,omitempty"`
}

type TaskRecord2 struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// 通过LogID和TaskID绑定task的一次运行记录
	LogID  string `bson:"logId,omitempty" json:"logId,omitempty"`
	TaskID string `bson:"taskId,omitempty" json:"taskId,omitempty"`

	// 记录类型， 1-任务记录， 2-场景记录
	Type string `bson:"type,omitempty" json:"type,omitempty"`

	// 任务记录详情
	TaskDetail TaskDetail `bson:"taskDetail,omitempty" json:"taskDetail,omitempty"`

	// 场景记录执行详情
	SceneDetail SceneDetail `bson:"sceneDetail,omitempty" json:"sceneDetail,omitempty"`
}

type TaskDetail struct {
	LogID      string        `bson:"logId,omitempty" json:"logId,omitempty"`
	TaskID     string        `bson:"taskId,omitempty" json:"taskId,omitempty"`
	TotalScene int           `bson:"totalScene" json:"totalScene"`
	ExecuteNum int           `bson:"executeNum" json:"executeNum"`
	SuccessNum int           `bson:"successNum" json:"successNum"`
	FailedNum  int           `bson:"failedNum" json:"failedNum"`
	State      int           `bson:"state" json:"state"`
	Duration   time.Duration `bson:"duration,omitempty" json:"duration,omitempty"`
	CreateAt   time.Time     `bson:"createAt,omitempty" json:"createAt,omitempty"`
	UpdateAt   time.Time     `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	FinishAt   time.Time     `bson:"finishAt,omitempty" json:"finishAt,omitempty"`
}

type SceneDetail struct {
	LogID        string         `bson:"logId,omitempty" json:"logId,omitempty"`
	TaskID       string         `bson:"taskId,omitempty" json:"taskId,omitempty"`
	SceneID      string         `bson:"sceneId,omitempty" json:"sceneId,omitempty"`
	TotalNum     int            `bson:"totalScene" json:"totalNum"`
	ExecuteNum   int            `bson:"executeNum" json:"executeNum"`
	SuccessNum   int            `bson:"successNum" json:"successNum"`
	FailedNum    int            `bson:"failedNum" json:"failedNum"`
	ActionDetail []actionDetail `bson:"actionDetail,omitempty" json:"actionDetail,omitempty"`
	State        int            `bson:"state" json:"state"`
	CreateAt     time.Time      `bson:"createAt,omitempty" json:"createAt,omitempty"`
	UpdateAt     time.Time      `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	FinishAt     time.Time      `bson:"finishAt,omitempty" json:"finishAt,omitempty"`
}

type actionDetail struct {
	ActionID     string                 `bson:"actionId,omitempty" json:"actionId,omitempty"`
	Request      http.Request           `bson:"request,omitempty" json:"request,omitempty"`
	Response     map[string]interface{} `bson:"response,omitempty" json:"response,omitempty"`
	VerifyResult interface{}            `bson:"verifyResult,omitempty" json:"verifyResult,omitempty"`
	State        int                    `bson:"state,omitempty" json:"state,omitempty"`
	Duration     time.Duration          `bson:"duration,omitempty" json:"duration,omitempty"`
	RetryTimes   int                    `bson:"retryTimes,omitempty" json:"retryTimes,omitempty"`
	StartTime    time.Time              `bson:"startTime,omitempty" json:"startTime,omitempty"`
	FinishTime   time.Time              `bson:"finishTime,omitempty" json:"finishTime,omitempty"`

	// 错误的类型，请求错误 / 超时 / 断言失败 / 超出重试次数
	ErrorType string `bson:"errorType,omitempty" json:"errorType,omitempty"`

	// 具体的错误类型
	Error interface{} `bson:"error,omitempty" json:"error,omitempty"`
}
