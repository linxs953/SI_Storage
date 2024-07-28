package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskRunLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Type       string             `bson:"type,omitempty" json:"type,omitempty"`
	RequestID  string             `bson:"requestId,omitempty" json:"requestId,omitempty"`
	SceneID    string             `bson:"sceneId,omitempty" json:"sceneId,omitempty"`
	ActionID   string             `bson:"actionId,omitempty" json:"actionId,omitempty"`
	Detail     ActionLog          `bson:"detail,omitempty" json:"detail,omitempty"`
	Total      int                `bson:"total,omitempty" json:"total,omitempty"`
	State      int                `bson:"state,omitempty" json:"state,omitempty"`
	StartTime  time.Time          `bson:"startTime,omitempty" json:"startTime,omitempty"`
	FinishTime time.Time          `bson:"finishTime,omitempty" json:"finishTime,omitempty"`
	UpdateAt   time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt   time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
}

type ActionLog struct {
	URL      string            `bson:"url,omitempty" json:"url,omitempty"`
	Method   string            `bson:"method,omitempty" json:"method,omitempty"`
	Headers  map[string]string `bson:"headers,omitempty" json:"headers,omitempty"`
	Payload  interface{}       `bson:"payload,omitempty" json:"payload,omitempty"`
	Response interface{}       `bson:"response,omitempty" json:"response,omitempty"`
	Verify   []interface{}     `bson:"verify,omitempty" json:"verify,omitempty"`
}
