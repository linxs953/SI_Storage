package apidetail

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Apidetail struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" bson:"id,omitempty"`
	Source        string             `bson:"source"`
	ApiId         int                `bson:"apiId"`
	FolderName    string             `bson:"folderName"`
	ApiName       string             `bson:"name"`
	ApiMethod     string             `bson:"method"`
	ApiAuthType   string             `bson:"authType"` // bearer / no auth
	AuthApiId     string             `bson:"authId"`   // 指定获取授权的 api
	ApiPath       string             `bson:"path"`
	ApiHeaders    []*ApiHeader       `bson:"headers"`
	ApiPayload    *ApiPayload        `bson:"payload"`
	ApiParameters []*ApiParameter    `bson:"parameters"`
	ApiResponse   *ApiResponseSpec   `bson:"response"`
	UpdateAt      time.Time          `bson:"updateAt,omitempty" bson:"updateAt,omitempty"`
	CreateAt      time.Time          `bson:"createAt,omitempty" bson:"createAt,omitempty"`
}

/*
存储到 mongodb 的 api 信息
*/
// type Detail struct {
// 	ApiId         int              `bson:"apiId"`
// 	FolderName    string           `bson:"folderName"`
// 	ApiName       string           `bson:"name"`
// 	ApiMethod     string           `bson:"method"`
// 	ApiAuthType   string           `bson:"authType"` // bearer / no auth
// 	AuthApiId     string           `bson:"authId"`   // 指定获取授权的 api
// 	ApiPath       string           `bson:"path"`
// 	ApiHeaders    []*ApiHeader     `bson:"headers"`
// 	ApiPayload    *ApiPayload      `bson:"payload"`
// 	ApiParameters []*ApiParameter  `bson:"parameters"`
// 	ApiResponse   *ApiResponseSpec `bson:"response"`
// }

type ApiHeader struct {
	HeaderName  string `bson:"headerName"`
	HeaderValue string `bson:"headerValue"`
}

type ApiPayload struct {
	ContentType       string                   `bson:"contentType"`
	PayloadString     string                   `bson:"payloadString"`     // content-type = x-www-form-urlencoded / applicationjson
	PayloadParameters []string                 `bson:"payloadParameters"` // content-type = form-data
	FormPayload       map[string]string        `bson:"form"`
	JsonPayload       map[string]FieldMapValue `bson:"json"`
}

type FieldMapValue struct {
	Type     string `bson:"type"`
	Required bool   `bson:"required"`
}

type ApiParameter struct {
	QueryName  string `bson:"queryName"`
	QueryValue string `bson:"queryValue"`
}

type ApiResponseSpec struct {
	Fields []*ApiResponseField `bson:"fields"`
}

type ApiResponseField struct {
	// response 字段的路径，比如 data.valid.$list.name
	FieldPath string `bson:"fieldPath"`
	Required  bool   `bson:"required"`
	FieldType string `bson:"fieldType"`
}
