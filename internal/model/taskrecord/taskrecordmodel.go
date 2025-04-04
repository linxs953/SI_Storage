package taskrecord

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskRecord represents a record in the task_record collection

const TaskRecordCollectionName = "task_record" // Collection name for task records

// TaskRecordModel handles operations on the task_record collection
type TaskRecordModel struct {
	collection *mongo.Collection
}

// NewTaskRecordModel creates a new TaskRecordModel instance
func NewTaskRecordModel(db *mongo.Database) *TaskRecordModel {
	return &TaskRecordModel{
		collection: db.Collection("task_record"),
	}
}

// Create inserts a new task record
func (m *TaskRecordModel) Create(ctx context.Context, record *TaskRecord) error {
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	record.UpdatedAt = record.CreatedAt

	_, err := m.collection.InsertOne(ctx, record)
	return err
}

// Update updates an existing task record
func (m *TaskRecordModel) Update(ctx context.Context, id primitive.ObjectID, update interface{}) error {
	updateDoc := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}
	if update != nil {
		updateDoc["$set"].(bson.M)["data"] = update
	}

	_, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		updateDoc,
	)
	return err
}

// UpdateByRecordId updates a task record by its RecordID
func (m *TaskRecordModel) UpdateByRecordId(ctx context.Context, recordID string, update bson.M) error {
	// 1. Query to validate the record exists
	filter := bson.M{"recordId": recordID}
	var existingRecord TaskRecord
	err := m.collection.FindOne(ctx, filter).Decode(&existingRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		return err
	}

	// 2. Update the record, appending the result to the result array

	_, err = m.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes a task record
func (m *TaskRecordModel) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := m.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// FindOne retrieves a single task record by ID
func (m *TaskRecordModel) FindOne(ctx context.Context, id primitive.ObjectID) (*TaskRecord, error) {
	var record TaskRecord
	err := m.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// FindAll retrieves all task records with optional pagination
func (m *TaskRecordModel) FindAll(ctx context.Context, skip, limit int64) ([]*TaskRecord, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := m.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []*TaskRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// FindByTaskID retrieves all records for a specific task ID
func (m *TaskRecordModel) FindByTaskID(ctx context.Context, taskID string) ([]*TaskRecord, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := m.collection.Find(ctx, bson.M{"task_id": taskID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []*TaskRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return records, nil
}
