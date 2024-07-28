package synchronizer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"

	logic "lexa-engine/internal/logic"
	"lexa-engine/internal/logic/common"
	"lexa-engine/internal/logic/sync/synchronizer/utils"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/svc"
)

var hc common.HttpClient

type DataHook struct {
	Event    string             `json:"event"`
	Data     interface{}        `json:"data"`
	IsEof    bool               `json:"iseof"`
	UpdateId primitive.ObjectID `json:"updateId"`
}

type ApiFoxSynchronizer struct {
	ApiFoxShareUrl        string                `json:"api_fox_share_url"`
	ApiFoxShareAuthUser   string                `json:"api_fox_share_auth_user"`
	ApiFoxShareAuthPasswd string                `json:"api_fox_share_auth_passwd"`
	ApiFoxDetailUrl       string                `json:"api_fox_detail_url"`
	ApiFoxTreeUrl         string                `json:"api_fox_tree_url"`
	ApiInfoList           []apidetail.Apidetail `json:"api_info_list"`
	Hooks                 []DataHook            `json:"hooks"`
}

func BuildApiFox(config ApiFoxSpec) *ApiFoxSynchronizer {
	if config.ShareUrl == "" {
		err := errors.New("apifox 分享链接为空")
		logx.Error(err)
		return nil
	}
	urlParts := strings.Split(config.ShareUrl, "shared-")
	if len(urlParts) == 0 || len(urlParts) < 2 {
		err := errors.New("url 不存在 【shared-】 解析 url 失败")
		logx.Error(fmt.Sprintf("%v, %v", err, config.ShareUrl))
		return nil
	}
	return &ApiFoxSynchronizer{
		ApiFoxShareAuthUser:   urlParts[1],
		ApiFoxShareAuthPasswd: config.ShareDocAuth,
		ApiFoxDetailUrl:       logic.APIFOX_DETAIL_URL,
		ApiFoxTreeUrl:         logic.APIFOX_TREE_URL,
		ApiFoxShareUrl:        logic.APIFOX_DOC_AUTH_URL,
	}
}

func (afSync *ApiFoxSynchronizer) Sync(ctx *svc.ServiceContext, recordId primitive.ObjectID) (err error) {
	logx.Info("同步 apifox 接口中")
	go afSync.apifoxSyncProccess(ctx, recordId)
	return
}

func (afSync *ApiFoxSynchronizer) Store(ctx *svc.ServiceContext, recordId primitive.ObjectID) (err error) {
	logx.Info("数据开始入库")
	for idx, apiInfo := range afSync.ApiInfoList {
		if apiInfo.ID.IsZero() {
			apiInfo.ID = primitive.NewObjectID()
		}
		apiInfo.CreateAt = time.Now()
		apiInfo.UpdateAt = time.Now()
		syncDataEvent := DataHook{
			Event: "sync_data",
			Data:  apiInfo,
		}
		if idx == len(afSync.ApiInfoList)-1 {
			syncDataEvent.IsEof = true
			syncDataEvent.UpdateId = recordId
		}
		afSync.Hooks = append(afSync.Hooks, syncDataEvent)
	}

	for _, hook := range afSync.Hooks {
		eventMsg, err := json.Marshal(hook)
		if err != nil {
			logx.Error("序列化 hook 失败", err)
			return err
		}

		if err := ctx.KqPusherClient.Push(string(eventMsg)); err != nil {
			logx.Error(err)
			return err
		}
	}

	return
}

func (afSync *ApiFoxSynchronizer) apifoxSyncProccess(ctx *svc.ServiceContext, recordId primitive.ObjectID) {
	var err error

	// 获取 apifox 授权
	if err = afSync.apifoxAuth(); err != nil {
		return
	}

	// 获取 apifox 接口列表
	apiIds, err := afSync.getApiDetailIds()
	if err != nil {
		return
	}

	// 构建folder层级关系
	folderLevelMap, err := afSync.BuildFolderLevelMap()
	if err != nil {
		return
	}
	var apiInfo apidetail.Apidetail
	for _, aid := range apiIds {
		if apiInfo, err = afSync.extractApiInfo(aid, folderLevelMap); err != nil {
			logx.Error(err)
			// continue
		}
		afSync.ApiInfoList = append(afSync.ApiInfoList, apiInfo)
	}
	if err = afSync.Store(ctx, recordId); err != nil {
		logx.Error(err)
		return
	}
}

/*

1. 获取doc授权
2. 获取doc api树
3. 提取Api id 列表
4. 获取api Detail
5. 提取detail信息，推入到kafka

*/

func (afSync *ApiFoxSynchronizer) apifoxAuth() (err error) {
	var body []common.RequestFormBodyParameter
	body = append(body, common.RequestFormBodyParameter{FormName: "id", FormValue: afSync.ApiFoxShareAuthUser})
	body = append(body, common.RequestFormBodyParameter{FormName: "password", FormValue: afSync.ApiFoxShareAuthPasswd})
	var headers []common.RequestHeader
	headers = append(headers, common.RequestHeader{Key: "X-Client-Version", Value: logic.HEADER_APIFOX_VERSION})
	headers = append(headers, common.RequestHeader{Key: "User-Agent", Value: logic.HEADER_USERAGENT_APIFOX})
	req := common.Request{
		Method:      "POST",
		ReqUrl:      afSync.ApiFoxShareUrl,
		PostType:    "form-data",
		SetCookies:  true,
		NeedCookies: false,
		BodyParam:   body,
		Headers:     headers,
	}
	logx.Error(req)
	logx.Info(fmt.Sprintf("获取apifox doc授权 [%s]", req.ReqUrl))
	resp, err := hc.SendRequest(&req)
	if err != nil {
		return
	}
	if resp == nil {
		err = errors.New("响应为空")
		return
	}
	respMap, err := common.ReadBody(resp)
	if err != nil {
		logx.Error("解析response报错")
		return
	}

	if field, ok := respMap["success"].(bool); !ok || !field {
		logx.Error(fmt.Sprintf("apifoxAuth data.success=[%v]", field))
		err = logic.APIFOX_DOC_AUTH_FAILED_ERROR
		return
	}
	return
}

func (afSync *ApiFoxSynchronizer) getApiDetailIds() (apiIds []string, err error) {
	resp, err := afSync.getApiTree()
	if err != nil {
		return
	}
	bodyMap, err := common.ReadBody(resp)
	if err != nil {
		return
	}
	bodyByte, err := json.Marshal(bodyMap)
	if err != nil {
		return
	}

	findApiIdRegex, err := regexp.Compile(`apiDetail.(\d+)`)
	if err != nil {
		return
	}
	for _, group := range findApiIdRegex.FindAllStringSubmatch(string(bodyByte), -1) {
		if len(group) < 2 {
			continue
		}
		apiIds = append(apiIds, group[1])
	}

	return
}

func (afSync *ApiFoxSynchronizer) getApiTree() (resp *http.Response, err error) {
	var headers []common.RequestHeader
	headers = append(headers, common.RequestHeader{Key: "X-Client-Version", Value: logic.HEADER_APIFOX_VERSION})
	headers = append(headers, common.RequestHeader{Key: "User-Agent", Value: logic.HEADER_USERAGENT_APIFOX})

	req := common.Request{
		Method:      "GET",
		ReqUrl:      strings.Replace(logic.APIFOX_TREE_URL, "{docid}", afSync.ApiFoxShareAuthUser, -1),
		SetCookies:  false,
		NeedCookies: true,
		Headers:     headers,
	}
	resp, err = hc.SendRequest(&req)
	if err != nil {
		return
	}
	if resp == nil {
		err = errors.New("响应为空")
		return
	}
	return
}

func (afSync *ApiFoxSynchronizer) extractApiInfo(apiId string, foldersLevel map[string]string) (apiDetail apidetail.Apidetail, err error) {
	apiDetail.Source = logic.SYNC_APIFOX
	getDetailUrl := strings.Replace(afSync.ApiFoxDetailUrl, "{docid}", afSync.ApiFoxShareAuthUser, -1)
	getDetailUrl = strings.Replace(getDetailUrl, "{apiid}", apiId, -1)
	logx.Info(getDetailUrl)
	var headers []common.RequestHeader
	headers = append(headers, common.RequestHeader{Key: "X-Client-Version", Value: logic.HEADER_APIFOX_VERSION})
	headers = append(headers, common.RequestHeader{Key: "User-Agent", Value: logic.HEADER_USERAGENT_APIFOX})
	req := common.Request{
		Method:      "GET",
		ReqUrl:      getDetailUrl,
		SetCookies:  false,
		NeedCookies: true,
		Headers:     headers,
	}
	resp, err := hc.SendRequest(&req)
	if err != nil {
		return
	}
	if resp == nil {
		err = errors.New("响应为空")
		return
	}
	bodyMap, err := common.ReadBody(resp)
	if err != nil {
		logx.Error("读取response 失败")
		return
	}

	// 获取resp.data
	dataMap := utils.ReadMapValueObject(bodyMap, "data")
	if dataMap == nil {
		err = errors.New(fmt.Sprintf("apifox detail %v 无data字段", apiId))
		return
	}

	// 获取apiid
	detailApiId := utils.ReadMapValueInteger(dataMap, "id")
	if detailApiId < 0 {
		err = errors.New("获取apiid失败")
		if detailApiId == -1 {
			logx.Error("detail.id不是数值类型")
		}
		if detailApiId == -2 {
			logx.Error("body无id字段")
		}
		return
	}
	apiDetail.ApiId = int(detailApiId)

	// 获取folderId
	folderId := utils.ReadMapValueInteger(dataMap, "folderId")
	if folderId < 0 {
		err = errors.New("获取folderId失败")
		if folderId == -1 {
			logx.Error("detail.folderId不是数值类型")
		}
		if folderId == -2 {
			logx.Error("body无id字段")
		}
		return
	}
	if _, ok := foldersLevel[strconv.Itoa(folderId)]; !ok {
		err = errors.New("读取folder level失败")
		logx.Error(foldersLevel)
		logx.Error(folderId)
		return
	}
	apiDetail.FolderName = foldersLevel[strconv.Itoa(folderId)]

	// 获取method
	method := utils.ReadMapValueString(dataMap, "method")
	if method == "" {
		err = errors.New("获取api method失败")
		return
	}
	apiDetail.ApiMethod = strings.ToUpper(method)

	// 获取api path
	apiPath := utils.ReadMapValueString(dataMap, "path")
	if apiPath == "" {
		err = errors.New("获取api path失败")
		return
	}
	apiDetail.ApiPath = apiPath

	// 获取api name
	apiName := utils.ReadMapValueString(dataMap, "name")
	if apiName == "" {
		err = errors.New("获取api名称失败")
		return
	}
	apiDetail.ApiName = apiName

	// 设置api auth
	auth := utils.ReadMapValueObject(dataMap, "auth")
	if auth == nil {
		err = errors.New("apiDetail没有 resp.data.auth字段")
		logx.Error(err)
		return
	}
	authType := utils.ReadMapValueString(auth, "type")
	if authType == "" {
		// logx.Error(fmt.Sprintf("获取 api=[%v] data.auth 失败 ", apiId))
		apiDetail.ApiAuthType = "0"
	}
	if authType == "noauth" {
		apiDetail.ApiAuthType = "1"
	}
	if authType == "bearer" {
		apiDetail.ApiAuthType = "2"
	}

	apiPayload, err := buildRequestBody(bodyMap)
	if err != nil {
		logx.Error("构建 api payload 失败")
		logx.Error(err)
		return
	}
	if apiPayload == nil {
		err = errors.New("api payload 为空")
		return
	}

	apiDetail.ApiPayload = apiPayload
	logx.Info("构建response 字段映射")
	vmap, err := buildResponseFields(bodyMap)
	if err != nil {
		logx.Error("构建response 字段映射 失败")
		return
	}
	logx.Info("构建response 字段映射 成功")

	// 生成apiInfo
	for fieldName, fieldInfo := range vmap {
		var apiRespField = apidetail.ApiResponseField{
			FieldPath: fieldName,
			FieldType: fieldInfo.Type,
			Required:  fieldInfo.Required,
		}
		if apiDetail.ApiResponse == nil {
			apiDetail.ApiResponse = &apidetail.ApiResponseSpec{}
		}
		apiDetail.ApiResponse.Fields = append(apiDetail.ApiResponse.Fields, &apiRespField)
	}

	// 构建 query
	apiQueries, err := buildRequestQuery(bodyMap)
	if err != nil {
		logx.Error("构建 api query 失败")
		logx.Error(err)
		return
	}
	apiDetail.ApiParameters = apiQueries
	// logx.Errorf("api query %v", apiDetail.ApiParameters[0].QueryName)

	apiHeaders := buildHeaders()
	if apiHeaders != nil {
		apiDetail.ApiHeaders = apiHeaders
	}

	apiInfoB, _ := json.Marshal(apiDetail)
	logx.Info(string(apiInfoB))
	return
}

func buildHeaders() (apiHeaders []*apidetail.ApiHeader) {
	commonHeaderApi := "https://apifox.com/api/v1/shared-docs/2e301f8d-5ee8-4b0d-a111-0a21c30ba557/common-parameters"
	var headers []common.RequestHeader
	headers = append(headers, common.RequestHeader{Key: "X-Client-Version", Value: logic.HEADER_APIFOX_VERSION})
	headers = append(headers, common.RequestHeader{Key: "User-Agent", Value: logic.HEADER_USERAGENT_APIFOX})
	req := common.Request{
		Method:      "GET",
		ReqUrl:      commonHeaderApi,
		SetCookies:  false,
		NeedCookies: true,
		Headers:     headers,
	}
	resp, err := hc.SendRequest(&req)
	if err != nil {
		return nil
	}
	if resp == nil {
		// err = errors.New("响应为空")
		logx.Error("获取公共参数,接口返回空")
		return nil
	}
	bodyMap, err := common.ReadBody(resp)
	if err != nil {
		logx.Error("读取response 失败")
		return nil
	}
	dataMap := utils.ReadMapValueObject(bodyMap, "data")
	paramsMap := utils.ReadMapValueObject(dataMap, "parameters")
	headerList := utils.ReadArrayAny(paramsMap, "header")
	for _, header := range headerList {
		headerAsMap, ok := header.(map[string]any)
		if !ok {
			logx.Error("common header 类型不是 map[string]any")
			continue
		}
		apiHeaders = append(apiHeaders, &apidetail.ApiHeader{
			HeaderName:  headerAsMap["name"].(string),
			HeaderValue: headerAsMap["defaultValue"].(string),
		})
	}
	return
}

func (afSync *ApiFoxSynchronizer) BuildFolderLevelMap() (folderLevelMap map[string]string, err error) {
	resp, err := afSync.getApiTree()
	if err != nil {
		return
	}
	bodyByte := common.ReadBodyByte(resp)
	if string(bodyByte) == "" {
		logx.Error("body返回为空")
		return
	}
	var apiFolders ApiFoxTree
	if err = json.Unmarshal(bodyByte, &apiFolders); err != nil {
		logx.Error(err)
		return
	}
	folderLevelMap = make(map[string]string)
	parseFolderLevel(apiFolders.Data, folderLevelMap, "")
	return
}

func buildRequestQuery(response map[string]any) (apiQuery []*apidetail.ApiParameter, err error) {
	dataMap := utils.ReadMapValueObject(response, "data")
	requestParams := utils.ReadMapValueObject(dataMap, "parameters")
	queries := utils.ReadArrayAny(requestParams, "query")
	for _, query := range queries {
		q, ok := query.(map[string]any)
		if !ok {
			logx.Error("data.parameters.query[] 不是 map[string]any 类型")
			break
		}
		paramName, ok := q["name"].(string)
		if !ok {
			logx.Errorf("data.parameters.query[].name 不是 string, value=[%v]", q["name"])
			break
		}
		valueType, ok := q["type"].(string)
		if !ok {
			logx.Errorf("data.parameters.query[].type 不是 string, value=[%v]", q["type"])
			break
		}
		apiQuery = append(apiQuery, &apidetail.ApiParameter{
			QueryName:  paramName,
			QueryValue: fmt.Sprintf("$%v", paramName),
			ValueType:  valueType,
		})
	}
	return
}

func buildRequestBody(response map[string]any) (apiPayload *apidetail.ApiPayload, err error) {
	apiPayload = &apidetail.ApiPayload{}
	dataMap := utils.ReadMapValueObject(response, "data")
	requestBody := utils.ReadMapValueObject(dataMap, "requestBody")
	requestBodyByte, err := json.Marshal(requestBody)
	if err != nil {
		return
	}
	var apiRequestBody ApiRequestBody
	if err = json.Unmarshal(requestBodyByte, &apiRequestBody); err != nil {
		return
	}

	if apiRequestBody.Type != "none" {
		apiPayload.ContentType = apiRequestBody.Type
	}

	if apiRequestBody.Type == "multipart/form-data" || apiRequestBody.Type == "application/x-www-form-urlencoded" {
		formMap := make(map[string]string)
		for _, p := range apiRequestBody.Parameters {
			if !p.Enable {
				continue
			}
			if p.ParamName == "" {
				return
			}
			formMap[p.ParamName] = fmt.Sprintf("$%v", p.ParamName)
		}
		// 根据form参数生成一个map[string]string, value是参数化的表达式
		apiPayload.FormPayload = formMap
	}

	if apiRequestBody.Type == "application/json" {
		jsonMap := make(map[string]apidetail.FieldMapValue)
		if err = parseRequestJsonSchema(apiRequestBody.JsonSchema, jsonMap); err != nil {
			return
		}
		apiPayload.JsonPayload = jsonMap
	}

	return
}

func buildResponseFields(response map[string]any) (value map[string]apidetail.FieldMapValue, err error) {
	data := utils.ReadMapValueObject(response, "data")
	if data == nil {
		err = errors.New("获取 data 字段失败")
		bmap, _ := json.Marshal(response)
		logx.Error(string(bmap))
		return
	}

	/* 读取response字段*/
	responseSpecArr := utils.ReadArrayAny(data, "responses")
	if responseSpecArr == nil {
		logx.Error("无效接口,响应结构无 responses 字段")
	}
	responseSpecMap, ok := responseSpecArr[0].(map[string]any)
	if !ok {
		err = errors.New("response.data.response.0 不是map")
		logx.Error(err)
		return
	}

	/* 读取jsonSchema字段*/
	jsonSchemaMap := utils.ReadMapValueObject(responseSpecMap, "jsonSchema")

	/* 读取property字段*/
	propertyMap := utils.ReadMapValueObject(jsonSchemaMap, "properties")

	/* 读取required字段 */
	var requiredFields []string
	requiredArr := utils.ReadArrayAny(jsonSchemaMap, "required")
	if requiredArr == nil {
		logx.Error("无效接口,响应结构无 required 字段")
	}
	for _, r := range requiredArr {
		requiredField, ok := r.(string)
		if !ok {
			continue
		}
		requiredFields = append(requiredFields, requiredField)
	}

	value = make(map[string]apidetail.FieldMapValue)
	if err = parseProperty(propertyMap, requiredFields, value, ""); err != nil {
		logx.Error("解析jsonSchema.property 失败")
		return
	}
	return
}

func parseFolderLevel(folders []ApiFoxTreeData, folderLevelMap map[string]string, parentName string) {
	for _, folder := range folders {
		if folder.Type != "apiDetailFolder" {
			continue
		}
		fullFolderName := folder.Name
		if parentName != "" {
			fullFolderName = fmt.Sprintf("%v.%v", parentName, fullFolderName)
		}

		folderLevelMap[strconv.Itoa(folder.Folder.Id)] = fullFolderName
		if len(folder.Children) > 0 {
			parseFolderLevel(folder.Children, folderLevelMap, fullFolderName)
		}
	}
}

func parseRequestJsonSchema(requestBody map[string]any, jsonMap map[string]apidetail.FieldMapValue) (err error) {
	schemaMap := utils.ReadMapValueObject(requestBody, "jsonSchema")

	/* 读取property字段*/
	propertyMap := utils.ReadMapValueObject(schemaMap, "properties")

	/* 读取required字段 */
	var requiredFields []string
	requiredArr := utils.ReadArrayAny(schemaMap, "required")
	if requiredArr == nil {
		logx.Error("无效接口,请求结构无 required 字段")
	}
	for _, r := range requiredArr {
		requiredField, ok := r.(string)
		if !ok {
			continue
		}
		requiredFields = append(requiredFields, requiredField)
	}

	if err = parseProperty(propertyMap, requiredFields, jsonMap, ""); err != nil {
		return
	}
	return
}

func parseProperty(property map[string]any, required []string, fieldMap map[string]apidetail.FieldMapValue, parentKey string) error {
	for key, value := range property {
		var fieldKey string
		if parentKey != "" {
			fieldKey = fmt.Sprintf("%v.%v", parentKey, key)
		} else {
			fieldKey = key
		}
		var propertyInfo Property
		valueByte, err := json.Marshal(value)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(valueByte, &propertyInfo); err != nil {
			// logx.Error("vaule: %v", string(valueByte))
			return err
		}
		// 断言类型
		dataType := utils.AssertString(propertyInfo.Type)
		if dataType == "" {
			dataTypeList := utils.AssertArrayAny(propertyInfo.Type)
			if len(dataTypeList) == 0 || dataTypeList == nil {
				err = errors.New(fmt.Sprintf("解析 api detail 失败, 字段类型不是Array[] / string, %v", propertyInfo.Type))
				logx.Error(err)
				return err
			}
			// todo: 暂时不处理字段类型 array, 值为空的情况
			dataType = utils.AssertString(dataTypeList[0])
		}
		if propertyInfo.Type == "object" {
			return parseProperty(propertyInfo.Properties, propertyInfo.Required, fieldMap, fieldKey)
		}
		if propertyInfo.Type == "array" {
			// 非同步自apifox，自定义的response，默认给true
			fieldMap[fieldKey] = apidetail.FieldMapValue{
				Type:     dataType,
				Required: true,
			}
			fieldKey += ".$idx"
			itemsByte, err := json.Marshal(propertyInfo.Items)
			if err != nil {
				return err
			}
			var itemsProperty Property
			if err = json.Unmarshal(itemsByte, &itemsProperty); err != nil {
				return err
			}
			itemDataType := utils.AssertString(itemsProperty.Type)
			if itemDataType == "" {
				dataTypeList := utils.AssertArrayAny(itemsProperty.Type)
				if len(dataTypeList) == 0 {
					err = errors.New(fmt.Sprintf("解析 api detail 失败, 字段类型不是Array[] / string, %v", itemsProperty.Type))
					logx.Error(err)
					return err
				}
				// todo: 暂时不处理字段类型 array, 值为空的情况
				itemDataType = utils.AssertString(dataTypeList[0])
			}
			if itemDataType == "object" {
				return parseProperty(itemsProperty.Properties, itemsProperty.Required, fieldMap, fieldKey)
			}
			fieldMap[fieldKey] = apidetail.FieldMapValue{
				Type:     dataType,
				Required: isFieldRequired(key, itemsProperty.Required),
			}
		}
		fieldMap[fieldKey] = apidetail.FieldMapValue{
			Type:     dataType,
			Required: isFieldRequired(key, required),
		}
	}
	return nil
}

func isFieldRequired(fieldName string, required []string) bool {
	for _, rk := range required {
		if fieldName == rk {
			return true
		}
	}
	return false
}
