package apitest

import (
	"context"
	"encoding/json"
	"fmt"
	"lexa-engine/internal/logic/common"
	"lexa-engine/internal/logic/sync/apitest/types"
	apitestUtils "lexa-engine/internal/logic/sync/apitest/utils"
	"lexa-engine/internal/logic/sync/synchronizer/utils"
	"lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/svc"
	"os"
	"regexp"

	"github.com/GUAIK-ORG/go-snowflake/snowflake"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/yaml.v3"
)

type ApiTestJob struct {
	Type    string `json:"type"`
	GitRepo string `json:"gitRepo"`
}

// type ApiRunnerTree struct {
// 	FirstNode *ApiRunnerNode
// 	NextNode  []*ApiRunnerNode
// }

type ApiRunnerNode struct {
	common.Request
	ID   string         `json:"id"`
	Next *ApiRunnerNode `json:"next"`
}

func (atj *ApiTestJob) Build(svcCtx *svc.ServiceContext) (tree *ApiRunnerNode) {
	tc := getDefaultTestCase()
	treeMap := make(map[string][]*ApiRunnerNode)
	if len(tc.Steps) < 1 {
		logx.Error("api testcase 未包含任何节点")
		return nil
	}
	for _, node := range tc.Steps {
		var dependNodes []*ApiRunnerNode
		runnerNode, err := proccessStep(svcCtx, node)
		if err != nil {
			return
		}
		treeMap[runnerNode.ID] = dependNodes
	}
	return
}

// 根据场景中的单个接口进行参数化处理
func proccessStep(svcCtx *svc.ServiceContext, step types.Step) (apiRunnerNode *ApiRunnerNode, err error) {
	apiRunnerNode = &ApiRunnerNode{}
	s, err := snowflake.NewSnowflake(int64(0), int64(0))
	if err != nil {
		return
	}
	id_ := s.NextVal()
	apiRunnerNode.ID = fmt.Sprintf("%v", id_)
	api, err := getApiInfo(svcCtx, step.ApiId)
	if err != nil {
		return
	}
	apiBytes, _ := json.Marshal(api)
	logx.Infof("获取api=[%v]信息\n%v", step.ApiId, string(apiBytes))
	apiRunnerNode, err = buildPre(api, step)
	if err != nil {
		return
	}
	bytes, _ := json.Marshal(apiRunnerNode)
	logx.Info(string(bytes))
	return
}

func getApiInfo(svcCtx *svc.ServiceContext, apiId int) (api *apidetail.Apidetail, err error) {
	murl := mongo.GetMongoUrl(svcCtx.Config.Database.Mongo)
	mod := apidetail.NewApidetailModel(murl, svcCtx.Config.Database.Mongo.UseDb, "ApiInfo")
	api, err = mod.FindByApiId(context.Background(), apiId)
	if err != nil {
		logx.Errorf("根据apiid=[%v]  查找detail 记录失败, %v", apiId, err)
		return
	}
	return
}

// 进行接口参数化
func buildPre(api *apidetail.Apidetail, step types.Step) (apiRunnerNode *ApiRunnerNode, err error) {
	apiRunnerNode = &ApiRunnerNode{
		Request: common.Request{
			ReqUrl:      api.ApiPath,
			Method:      api.ApiMethod,
			PostType:    api.ApiPayload.ContentType,
			NeedCookies: false,
			SetCookies:  false,
		},
		Next: nil,
		ID:   "",
	}
	// 设置Request 请求参数
	var queryList []common.RequestQuery
	if api.ApiParameters != nil {
		for _, q := range api.ApiParameters {
			queryList = append(queryList, common.RequestQuery{
				QueryName:  q.QueryName,
				QueryValue: q.QueryValue,
				Type:       q.ValueType,
			})
		}
		apiRunnerNode.Parameters = queryList
	}

	// 设置Request headers
	if api.ApiHeaders != nil {
		var headers []common.RequestHeader
		for _, h := range api.ApiHeaders {
			headers = append(headers, common.RequestHeader{
				Key:   h.HeaderName,
				Value: h.HeaderValue,
			})
		}
		apiRunnerNode.Headers = headers
	}

	// 设置Request body
	if api.ApiPayload != nil {
		var bodyParamList []common.RequestFormBodyParameter
		if api.ApiPayload.ContentType == "multipart/form-data" || api.ApiPayload.ContentType == "application/x-www-form-urlencoded" {
			for k, v := range api.ApiPayload.FormPayload {
				bodyParamList = append(bodyParamList, common.RequestFormBodyParameter{
					FormName:  k,
					FormValue: v,
				})
			}
		}
		apiRunnerNode.BodyParam = bodyParamList
	}

	return
}

func (atj *ApiTestJob) Run() (err error) {
	return
}

func getDefaultTestCase() (tc types.ApiSuit) {
	return mashralConfig("F:\\go-pipeline-engine\\go-pipeline-engine\\go-pipeline-engine\\etc\\testcase.yaml")
}

func mashralConfig(file string) (tc types.ApiSuit) {
	dataBytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("读取文件失败：", err)
		return
	}

	err = yaml.Unmarshal(dataBytes, &tc)
	if err != nil {
		fmt.Println("解析 yaml 文件失败：", err)
		return
	}
	// sceneId只做标识,标识一个运行的场景
	tc.SceneId = uuid.NewString()

	// 填充output.steId
	for idx, step := range tc.Steps {
		tc.Steps[idx].StepId = uuid.NewString()
		for odx := range step.Output {
			tc.Steps[idx].Output[odx].Key = tc.Steps[idx].StepId
		}
	}

	// 填充expect
	for idx, step := range tc.Steps {
		for edx, field := range step.Expect.Fields {
			desireStr := utils.AssertString(field.Desire)
			reg, err := regexp.Compile(`{TYPE::(?P<type>.*?);VALUE=(?P<value>.*?)}`)
			if err != nil {
				logx.Error("构建正则失败")
				return
			}
			match := reg.FindStringSubmatch(desireStr)
			regResult := make(map[string]string)
			if len(match) > 0 {
				for i, name := range reg.SubexpNames() {
					if i != 0 && name != "" {
						regResult[name] = match[i]
					}
				}
				switch regResult["type"] {
				case "TIMESTAMP":
					{
						tc.Steps[idx].Expect.Fields[edx].Desire = apitestUtils.GetTimeStamp(regResult["value"])
						break
					}
				}
			}
			// logx.Error(tc.Steps[idx].Expect.Fields[edx].Desire)
		}
	}
	return
}
