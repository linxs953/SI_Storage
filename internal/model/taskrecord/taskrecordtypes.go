package taskrecord

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskRecord represents a record of a task execution
// type TaskRecord struct {
// 	ID          primitive.ObjectID `bson:"_id" json:"id"`
// 	RecordID    string            `bson:"recordId" json:"recordId"`
// 	TaskID      string            `bson:"taskId" json:"taskId"`
// 	TaskType    int32             `bson:"taskType" json:"taskType"`
// 	TaskVersion int64             `bson:"taskVersion" json:"taskVersion"`
// 	Status      string            `bson:"status" json:"status"` // pending, running, completed, failed
// 	StartTime   time.Time         `bson:"startTime" json:"startTime"`
// 	EndTime     time.Time         `bson:"endTime,omitempty" json:"endTime,omitempty"`
// 	CreateAt    time.Time         `bson:"createAt" json:"createAt"`
// 	UpdateAt    time.Time         `bson:"updateAt" json:"updateAt"`

// 	// Execution Details
// 	ExecutionDetails *ExecutionDetails `bson:"executionDetails" json:"executionDetails"`
// }

// // ExecutionDetails contains detailed information about the task execution
// type ExecutionDetails struct {
// 	Attempts      int                `bson:"attempts" json:"attempts"`           // Current attempt number
// 	MaxAttempts   int                `bson:"maxAttempts" json:"maxAttempts"`    // Maximum allowed attempts
// 	Duration      time.Duration      `bson:"duration" json:"duration"`          // Total execution duration
// 	Error         *ExecutionError    `bson:"error,omitempty" json:"error,omitempty"`
// 	APIExecution  *APIExecution      `bson:"apiExecution,omitempty" json:"apiExecution,omitempty"`
// 	SyncExecution *SyncExecution     `bson:"syncExecution,omitempty" json:"syncExecution,omitempty"`
// 	Metrics       *ExecutionMetrics  `bson:"metrics" json:"metrics"`
// }

// // ExecutionError represents detailed error information when task execution fails
// type ExecutionError struct {
// 	Code       string `bson:"code" json:"code"`
// 	Message    string `bson:"message" json:"message"`
// 	Details    string `bson:"details,omitempty" json:"details,omitempty"`
// 	RetryAfter int    `bson:"retryAfter,omitempty" json:"retryAfter,omitempty"` // Seconds to wait before retry
// }

// // APIExecution contains details specific to API task execution
// type APIExecution struct {
// 	ScenarioResults []ScenarioResult `bson:"scenarioResults" json:"scenarioResults"`
// 	TotalScenarios  int             `bson:"totalScenarios" json:"totalScenarios"`
// 	PassedScenarios int             `bson:"passedScenarios" json:"passedScenarios"`
// 	FailedScenarios int             `bson:"failedScenarios" json:"failedScenarios"`
// }

// // ScenarioResult represents the result of a scenario execution
// type ScenarioResult struct {
// 	ID          string    `bson:"id" json:"id"`
// 	Name        string    `bson:"name" json:"name"`
// 	Status      string    `bson:"status" json:"status"` // success, failed
// 	StartTime   time.Time `bson:"startTime" json:"startTime"`
// 	EndTime     time.Time `bson:"endTime" json:"endTime"`
// 	Error       string    `bson:"error,omitempty" json:"error,omitempty"`
// }

// // SyncExecution contains details specific to sync task execution
// type SyncExecution struct {
// 	SyncType        string            `bson:"syncType" json:"syncType"`
// 	SourceInfo      map[string]string `bson:"sourceInfo" json:"sourceInfo"`
// 	ProcessedItems  int               `bson:"processedItems" json:"processedItems"`
// 	SuccessfulItems int               `bson:"successfulItems" json:"successfulItems"`
// 	FailedItems     int               `bson:"failedItems" json:"failedItems"`
// }

// // ExecutionMetrics contains performance and resource usage metrics
// type ExecutionMetrics struct {
// 	CPUUsage    float64 `bson:"cpuUsage" json:"cpuUsage"`       // Percentage
// 	MemoryUsage int64   `bson:"memoryUsage" json:"memoryUsage"` // Bytes
// 	IOReads     int64   `bson:"ioReads" json:"ioReads"`
// 	IOWrites    int64   `bson:"ioWrites" json:"ioWrites"`
// }

type TaskRecord struct {
	ID        primitive.ObjectID       `bson:"_id,omitempty"`
	TaskID    string                   `bson:"task_id"`
	TaskType  string                   `bson:"task_type"`
	SubType   string                   `bson:"sub_type"` // 任务子类型, 比如 tasktype=sync, subtype=apifox
	CreatedAt time.Time                `bson:"created_at"`
	UpdatedAt time.Time                `bson:"updated_at"`
	Status    string                   `bson:"status"`
	TaskSpec  map[string]interface{}   `bson:"task_spec,omitempty"`
	Result    []map[string]interface{} `bson:"result,omitempty"`
	// 任务结果,sync 任务,表示是同步 dest 的状态
	// 任务结果,apiruntime 任务,表示是最后的执行结果
}
