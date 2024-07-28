package model

import (
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
