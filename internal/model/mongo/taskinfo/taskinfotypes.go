package taskinfo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"lexa-engine/internal/logic"
)

type TaskInfo struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TaskID   string             `bson:"taskId,omitempty" json:"taskId,omitempty"`
	TaskName string             `bson:"taskName,omitempty" json:"taskName,omitempty"`
	Author   string             `bson:"author,omitempty" json:"author,omitempty"`
	Scenes   []logic.Scene      `bson:"scenes,omitempty" json:"scenes,omitempty"`
	UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
