package sync_task

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Synctask struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// SyncTaskID string             `bson:"syncTaskId"`
	SyncType  int       `bson:"syncType"`
	State     int       `bson:"state"`
	TimeStamp int64     `bson:"timestamp"`
	UpdateAt  time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt  time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
