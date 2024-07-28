package taskinfo

import (
	"lexa-engine/internal/logic"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskInfo struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TaskID   string             `bson:"taskId,omitempty" json:"taskId,omitempty"`
	TaskName string             `bson:"taskName,omitempty" json:"taskName,omitempty"`
	Scenes   []logic.Scene      `bson:"scenes,omitempty" json:"scenes,omitempty"`
	UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
