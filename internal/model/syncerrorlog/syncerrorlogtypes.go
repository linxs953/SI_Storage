package syncerrorlog

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SyncErrorLog represents a sync error log document in MongoDB
type SyncErrorLog struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	TaskID     string             `bson:"taskId" json:"taskId"`             // Task ID that failed
	ApiID      string             `bson:"apiId,omitempty" json:"apiId"`     // API ID if applicable
	ProjectID  string             `bson:"projectId" json:"projectId"`       // Project ID
	ErrorMsg   string             `bson:"errorMsg" json:"errorMsg"`         // Error message
	RawData    string             `bson:"rawData,omitempty" json:"rawData"` // Raw data that failed to process
	RetryCount int                `bson:"retryCount" json:"retryCount"`     // Number of retry attempts
	CreateAt   time.Time          `bson:"createAt" json:"createAt"`         // When the error occurred
}
