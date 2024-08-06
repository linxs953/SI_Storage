package syncApi

import "go.mongodb.org/mongo-driver/bson/primitive"

type ApiFoxTree struct {
	Data []ApiFoxTreeData `json:"data"`
}

type ApiFoxTreeData struct {
	Children []ApiFoxTreeData     `json:"children"`
	Key      string               `json:"key"`
	Name     string               `json:"name"`
	Type     string               `json:"type"`
	Folder   ApiFoxTreeFolderInfo `json:"folder"`
}

type ApiFoxTreeFolderInfo struct {
	Id int `json:"id"`
}

type ApiRequestBody struct {
	Type       string                `json:"type"`
	Parameters []ApiPayloadParameter `json:"parameters"`
	JsonSchema map[string]any        `json:"jsonSchema"`
}

type ApiPayloadParameter struct {
	Description string `json:"description"`
	Enable      bool   `json:"enable"`
	ParamName   string `json:"name"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
}

type Property struct {
	Type       any            `json:"type"`
	Required   []string       `json:"required"`
	Properties map[string]any `json:"properties"`
	Items      map[string]any `json:"items"`
}

type ApiSyncEvent struct {
	Type     string             `json:"type"`
	Data     interface{}        `json:"data"`
	IsEof    bool               `json:"iseof"`
	UpdateId primitive.ObjectID `json:"updateId"`
}
