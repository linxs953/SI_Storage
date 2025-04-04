package api

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// // API represents the basic structure of an API endpoint
// type API struct {
// 	ID          primitive.ObjectID `bson:"_id" json:"id"`
// 	ApiID       string            `bson:"apiId" json:"apiId"`
// 	Name        string            `bson:"name" json:"name"`
// 	Description string            `bson:"description" json:"description"`
// 	Version     int64             `bson:"version" json:"version"`
// 	CreateAt    time.Time         `bson:"createAt" json:"createAt"`
// 	UpdateAt    time.Time         `bson:"updateAt" json:"updateAt"`

// 	// API Specification
// 	Spec *APISpecification `bson:"spec" json:"spec"`
// }

// // APISpecification contains the detailed API endpoint information
// type APISpecification struct {
// 	Method      string            `bson:"method" json:"method"`
// 	Path        string            `bson:"path" json:"path"`
// 	Headers     map[string]string `bson:"headers" json:"headers"`
// 	Parameters  []Parameter       `bson:"parameters,omitempty" json:"parameters,omitempty"`
// 	RequestBody *RequestBody      `bson:"requestBody,omitempty" json:"requestBody,omitempty"`
// 	Responses   []Response        `bson:"responses" json:"responses"`
// }

// // Parameter represents an API parameter (query, path, or header)
// type Parameter struct {
// 	Name        string      `bson:"name" json:"name"`
// 	In          string      `bson:"in" json:"in"` // query, path, header
// 	Required    bool        `bson:"required" json:"required"`
// 	Type        string      `bson:"type" json:"type"`
// 	Description string      `bson:"description" json:"description"`
// 	Schema      interface{} `bson:"schema,omitempty" json:"schema,omitempty"`
// }

// // RequestBody represents the API request body specification
// type RequestBody struct {
// 	Required    bool        `bson:"required" json:"required"`
// 	ContentType string      `bson:"contentType" json:"contentType"`
// 	Schema      interface{} `bson:"schema,omitempty" json:"schema,omitempty"`
// }

// // Response represents an API response specification
// type Response struct {
// 	StatusCode  int         `bson:"statusCode" json:"statusCode"`
// 	Description string      `bson:"description" json:"description"`
// 	Headers     Headers     `bson:"headers,omitempty" json:"headers,omitempty"`
// 	ContentType string      `bson:"contentType" json:"contentType"`
// 	Schema      interface{} `bson:"schema,omitempty" json:"schema,omitempty"`
// }

// // Headers represents response headers
// type Headers map[string]Header

// // Header represents a single header specification
// type Header struct {
// 	Description string `bson:"description" json:"description"`
// 	Type        string `bson:"type" json:"type"`
// 	Required    bool   `bson:"required" json:"required"`
// }

// Parameter represents a request parameter in the API
type Parameter struct {
	Name     string `bson:"name" json:"name"`
	Type     string `bson:"type" json:"type"`
	Required bool   `bson:"required" json:"required"`
}

// Header represents an HTTP header in the API
type Header struct {
	Name     string `bson:"name" json:"name"`
	Value    string `bson:"value" json:"value"`
	Required bool   `bson:"required" json:"required"`
}

// Api represents an API document in MongoDB
type Api struct {
	ID          primitive.ObjectID     `bson:"_id" json:"id"`
	ApiID       string                 `bson:"apiId" json:"apiId"`   // Original API ID from ApiFox
	Name        string                 `bson:"name" json:"name"`     // API name
	Method      string                 `bson:"method" json:"method"` // HTTP method
	Path        string                 `bson:"path" json:"path"`     // API path
	Description string                 `bson:"description" json:"description"`
	Headers     []Header               `bson:"headers" json:"headers"`
	Parameters  []Parameter            `bson:"parameters" json:"parameters"`
	Responses   interface{}            `bson:"responses" json:"responses"`
	RawData     map[string]interface{} `bson:"rawData" json:"rawData"`
	ProjectID   string                 `bson:"projectId" json:"projectId"` // ApiFox project ID
	TaskID      string                 `bson:"taskId" json:"taskId"`       // Task ID that created/updated this API
	CreateAt    time.Time              `bson:"createAt" json:"createAt"`
	UpdateAt    time.Time              `bson:"updateAt" json:"updateAt"`
}
