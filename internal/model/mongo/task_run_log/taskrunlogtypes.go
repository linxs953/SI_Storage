package task_run_log

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type TaskRunLog struct {
// 	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
// 	Type       string             `bson:"type,omitempty" json:"type,omitempty"`
// 	RequestID  string             `bson:"requestId,omitempty" json:"requestId,omitempty"`
// 	SceneID    string             `bson:"sceneId,omitempty" json:"sceneId,omitempty"`
// 	ActionID   string             `bson:"actionId,omitempty" json:"actionId,omitempty"`
// 	Detail     ActionLog          `bson:"detail,omitempty" json:"detail,omitempty"`
// 	Total      int                `bson:"total,omitempty" json:"total,omitempty"`
// 	State      int                `bson:"state,omitempty" json:"state,omitempty"`
// 	StartTime  time.Time          `bson:"startTime,omitempty" json:"startTime,omitempty"`
// 	FinishTime time.Time          `bson:"finishTime,omitempty" json:"finishTime,omitempty"`
// 	UpdateAt   time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
// 	CreateAt   time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
// }

// type ActionLog struct {
// 	URL      string            `bson:"url,omitempty" json:"url,omitempty"`
// 	Method   string            `bson:"method,omitempty" json:"method,omitempty"`
// 	Headers  map[string]string `bson:"headers,omitempty" json:"headers,omitempty"`
// 	Payload  interface{}       `bson:"payload,omitempty" json:"payload,omitempty"`
// 	Response interface{}       `bson:"response,omitempty" json:"response,omitempty"`
// 	Verify   []interface{}     `bson:"verify,omitempty" json:"verify,omitempty"`
// }

type TaskRunLog struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ExecID       string             `bson:"execId,omitempty" json:"execId,omitempty"`
	LogType      string             `bson:"logType,omitempty" json:"logType,omitempty"`
	SceneDetail  *SceneLog          `bson:"sceneDetail,omitempty" json:"sceneDetail,omitempty"`
	ActionDetail *ActionLog         `bson:"actionDetail,omitempty" json:"actionDetail,omitempty"`
	CreateAt     time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
	UpdateAt     time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
}

type SceneLog struct {
	SceneID      string      `bson:"sceneId,omitempty" json:"sceneId,omitempty"`
	Events       []EventMeta `bson:"events,omitempty" json:"events,omitempty"`
	FinishCount  int         `bson:"finishCount" json:"finishCount"`
	SuccessCount int         `bson:"successCount" json:"successCount"`
	FailCount    int         `bson:"failCount" json:"failCount"`
	Duration     int         `bson:"duration" json:"duration"`
	State        int         `bson:"state" json:"state"`
	Error        string      `bson:"error" json:"error"`
}

type ActionLog struct {
	SceneID  string                 `bson:"sceneId,omitempty" json:"sceneId,omitempty"`
	ActionID string                 `bson:"actionId,omitempty" json:"actionId,omitempty"`
	Events   []EventMeta            `bson:"events,omitempty" json:"events,omitempty"`
	Request  *RequestMeta           `bson:"request" json:"request"`
	Response map[string]interface{} `bson:"response" json:"response"`
	Error    string                 `bson:"error" json:"error"`
	State    int                    `bson:"state" json:"state"`
	Duration int                    `bson:"duration" json:"duration"`
}

type RequestMeta struct {
	URL        string                   `bson:"url,omitempty" json:"url,omitempty"`
	Method     string                   `bson:"method,omitempty" json:"method,omitempty"`
	Headers    map[string]string        `bson:"headers,omitempty" json:"headers,omitempty"`
	Payload    interface{}              `bson:"payload,omitempty" json:"payload,omitempty"`
	Dependency []map[string]interface{} `bson:"dependency,omitempty" json:"dependency,omitempty"`
}

type EventMeta struct {
	EventName string `bson:"eventName,omitempty" json:"eventName,omitempty"`
	EventType string `bson:"eventType,omitempty" json:"eventType,omitempty"`
	Message   string `bson:"message,omitempty" json:"message,omitempty"`
	Error     string `bson:"error,omitempty" json:"error,omitempty"`
	// Duration    int       `bson:"duration,omitempty" json:"duration,omitempty"`
	TriggerTime time.Time `bson:"triggerTime,omitempty" json:"triggerTime,omitempty"`
	State       int       `bson:"state,omitempty" json:"state,omitempty"`
}
