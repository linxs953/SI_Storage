package sceneinfo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"lexa-engine/internal/logic"
)

type SceneInfo struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
	logic.Scene
}
