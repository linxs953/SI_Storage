package scene

import (
	"context"
	"fmt"
	"lexa-engine/internal/logic"
	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/model/mongo/sceneinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/yaml.v2"
)

type NewSceneLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNewSceneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NewSceneLogic {
	return &NewSceneLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 通过apiid生成场景模板
func (l *NewSceneLogic) NewScene(req *types.CreateSceneDto) (*types.CreateSceneVO, error) {
	logx.Info("开始创建场景")
	resp := &types.CreateSceneVO{
		Code:    0,
		Message: "创建场景成功",
	}
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	var apiIds []int
	for _, apiIdStr := range req.Actions {
		apiId, err := strconv.ParseInt(apiIdStr, 10, 64)
		if err != nil {
			logx.Errorf("Error parsing API ID: %v", err)
			resp.Code = 1
			resp.Message = err.Error()
			return resp, err
		}
		apiIds = append(apiIds, int(apiId))
	}
	scene, err := sceneTempGen(apiIds, req.Scname, req.Description, req.Author, murl)
	if err != nil {
		resp.Code = 2
		resp.Message = err.Error()
		return resp, err
	}
	if len(scene.Actions) == 0 {
		resp.Code = 3
		resp.Message = "生成的场景是空"
		return resp, fmt.Errorf("场景创建失败")
	}
	smod := sceneinfo.NewSceneInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SceneInfo")
	if err := smod.Insert(context.Background(), &sceneinfo.SceneInfo{
		ID:       primitive.NewObjectID(),
		Scene:    *scene,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}); err != nil {
		resp.Code = 4
		resp.Message = err.Error()
		return resp, err
	}
	resp.Data.Author = scene.Author
	resp.Data.SceneId = scene.SceneId
	resp.Data.SceneName = scene.SceneName
	return resp, nil
}

func sceneTempGen(apiIDs []int, sceneName, description, author, mongourl string) (*logic.Scene, error) {
	usedb := "lct"
	mod := apidetail.NewApidetailModel(mongourl, usedb, "ApiInfo")
	scenes := &logic.Scene{
		SceneId:     uuid.New().String(),
		SceneName:   sceneName,
		Author:      author,
		Description: description,
		EnvKey:      "test",
		Timeout:     10,
		Retry:       3,
		SearchKey:   fmt.Sprintf("SC-%s", encodeToBase36(GenerateId())),
	}
	actions := []logic.Action{}
	for _, apiID := range apiIDs {
		apid, err := mod.FindByApiId(context.Background(), apiID)
		if err != nil {
			logx.Errorf("Error getting API detail: %v", err)
			return nil, err
		}

		action := logic.Action{
			EnvKey:       "test",
			DomainKey:    getActionDomainKey("test", apid.ApiPath),
			RelateId:     apid.ApiId,
			SearchKey:    fmt.Sprintf("AC-%s", encodeToBase36(GenerateId())),
			ActionID:     uuid.New().String(),
			ActionName:   apid.ApiName,
			ActionPath:   apid.ApiPath,
			ActionMethod: strings.ToUpper(apid.ApiMethod),
			Expect: logic.Expect{
				Api: []logic.Api{},
			},
			Headers:    make(map[string]string),
			Dependency: []logic.Dependency{},
		}
		if apid.ApiResponse != nil && apid.ApiResponse.Fields != nil {
			for _, field := range apid.ApiResponse.Fields {
				action.Expect.Api = append(action.Expect.Api, logic.Api{
					Type: "api",
					Data: logic.Data{
						Name:      field.FieldPath,
						Operation: "equal",
						Type:      field.FieldType,
						Desire:    "",
					},
				})
			}
		}

		if apid.ApiAuthType == "2" {
			refer := logic.Refer{
				Type:     "headers",
				DataType: "string",
				Target:   "Authorization",
			}
			action.Dependency = append(action.Dependency, logic.Dependency{
				Refer:     refer,
				DataKey:   "data.token",
				ActionKey: "$sid.$aid",
				Type:      "1",
				Mode:      "1",
				Extra:     "",
				DataSource: []logic.DependInject{
					{
						DependId:      fmt.Sprintf("DEPEND_%s", encodeToBase36(GenerateId())),
						Type:          "1",
						DataKey:       "$data.token",
						ActionKey:     "$sid.$aid",
						SearchCondArr: []logic.SearchCond{},
					},
				},
				DsSpec:    make([]logic.DataSourceSpec, 0),
				IsMultiDs: false,
				Output: logic.OutputSpec{
					Value: "",
					Type:  "string",
				},
			})
		}

		if apid.ApiParameters != nil {
			for _, q := range apid.ApiParameters {
				if q == nil {
					continue
				}
				refer := logic.Refer{
					Type:     "query",
					DataType: q.ValueType,
					Target:   fmt.Sprintf("url.query.$%s", q.QueryName),
				}
				action.Dependency = append(action.Dependency, logic.Dependency{
					Refer:     refer,
					DataKey:   "query字段值",
					ActionKey: "",
					Type:      "3",
					Mode:      "1",
					Extra:     "",
					DsSpec:    make([]logic.DataSourceSpec, 0),
					IsMultiDs: false,
					DataSource: []logic.DependInject{
						{
							DependId:      fmt.Sprintf("DEPEND_%s", encodeToBase36(GenerateId())),
							Type:          "3",
							DataKey:       "query字段值",
							ActionKey:     "",
							SearchCondArr: []logic.SearchCond{},
						},
					},
					Output: logic.OutputSpec{
						Value: "",
						Type:  "string",
					},
				})
			}
		}

		if apid.ApiPayload.ContentType == "application/x-www-form-urlencoded" {
			if apid.ApiPayload.FormPayload != nil {
				for fname, fvalue := range apid.ApiPayload.FormPayload {
					var refer logic.Refer
					refer.DataType = "string"
					refer.Type = "payload"
					refer.Target = fmt.Sprintf("%s.%s", "payload", fname)
					action.Dependency = append(action.Dependency, logic.Dependency{
						Refer:     refer,
						DataKey:   "$data." + fvalue,
						ActionKey: fmt.Sprintf("%s.%s", "$scenename", "$actionname"),
						Type:      "1",
						Mode:      "1",
						IsMultiDs: false,
						Extra:     "",
						DsSpec:    make([]logic.DataSourceSpec, 0),
						DataSource: []logic.DependInject{
							{
								DependId:      fmt.Sprintf("DEPEND_%s", encodeToBase36(GenerateId())),
								Type:          "1",
								DataKey:       "$data." + fvalue,
								ActionKey:     fmt.Sprintf("%s.%s", "$scenename", "$actionname"),
								SearchCondArr: []logic.SearchCond{},
							},
						},
						Output: logic.OutputSpec{
							Value: nil,
							Type:  "",
						},
					})
				}
			}
		}

		if apid.ApiPayload.ContentType == "application/json" {
			if apid.ApiPayload.PayloadString != "" {
				payload := make(map[string]interface{})
				err := yaml.Unmarshal([]byte(apid.ApiPayload.PayloadString), &payload)
				if err != nil {
					logx.Errorf("Error unmarshaling YAML: %v", err)
					return nil, err
				}
			}
		}
		action.Output = logic.Output{
			Key: "$scenename.$actionname",
		}

		if apid.ApiMethod == "GET" {
			apid.ApiHeaders = append(apid.ApiHeaders, &apidetail.ApiHeader{
				HeaderName:  "Content-Type",
				HeaderValue: "application/json; charset=utf-8",
			})
		} else {
			apid.ApiHeaders = append(apid.ApiHeaders, &apidetail.ApiHeader{
				HeaderName:  "Content-Type",
				HeaderValue: apid.ApiPayload.ContentType,
			})
		}

		if apid.ApiHeaders != nil {
			for _, header := range apid.ApiHeaders {
				action.Headers[header.HeaderName] = strings.ReplaceAll(header.HeaderValue, " ", "")
			}
		}

		if apid.ApiHeaders != nil {
			for _, header := range apid.ApiHeaders {
				var refer logic.Refer
				refer.DataType = "string"
				refer.Type = "headers"
				refer.Target = header.HeaderName
				action.Dependency = append(action.Dependency, logic.Dependency{
					Refer:     refer,
					DataKey:   header.HeaderValue,
					ActionKey: "",
					Type:      "3",
					Mode:      "1",
					Extra:     "",
					IsMultiDs: false,
					DsSpec:    make([]logic.DataSourceSpec, 0),
					DataSource: []logic.DependInject{
						{
							DependId:      fmt.Sprintf("DEPEND_%s", encodeToBase36(GenerateId())),
							Type:          "3",
							DataKey:       header.HeaderValue,
							ActionKey:     "",
							SearchCondArr: []logic.SearchCond{},
						},
					},
					Output: logic.OutputSpec{
						Value: "",
						Type:  "string",
					},
				})
			}
		}

		action.Retry = 3
		action.Timeout = 5
		actions = append(actions, action)
	}
	scenes.Actions = actions
	return scenes, nil
}

func getActionDomainKey(env string, apipath string) string {
	var domain string
	if env == "test" {
		domain = "demopsu.86yfw.com"
		if strings.Contains(apipath, "adminapi") {
			domain = "demopsuadmin.86yfw.com"
		}
		if strings.Contains(apipath, "merchant") {
			domain = "demopsub.86yfw.com"
		}
		if strings.Contains(apipath, "center") {
			domain = "demopsuopen.86yfw.com"
		}
	}
	return domain
}

const (
	// 时间戳偏移量，可以根据需要调整
	timestampLeftShift = 12 + 10
	// 序列号位数
	sequenceBits = 12
	// 机器ID位数
	machineIdBits = 10
	// 最大序列号
	maxSequence = -1 ^ (-1 << sequenceBits)
	// 序列掩码
	sequenceMask = maxSequence
	// 机器ID掩码
	machineIdMask = -1 ^ (-1 << machineIdBits)
	// 起始时间戳，可以根据需要调整
	startTime int64 = 1577836800000 // 2020-01-01 00:00:00 UTC
)

var (
	lastTimestamp int64 = -1
	sequence      int64 = 0
)

// 固定的机器ID，因为是单机部署
const machineId = 1

// 定义字符集
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func GenerateId() int64 {
	var timestamp int64

	// 获取当前时间戳，单位为毫秒
	timestamp = time.Now().UnixNano() / 1e6

	// 如果当前时间戳小于上一次时间戳，则发生时钟回拨
	if timestamp < lastTimestamp {
		panic(fmt.Sprintf("Clock moved backwards. Refusing to generate id for %d milliseconds", lastTimestamp-timestamp))
	}

	// 如果当前时间戳与上一次相同，则使用序列号
	if timestamp == lastTimestamp {
		sequence = (sequence + 1) & sequenceMask
		if sequence == 0 {
			timestamp = getNextMillisecond(lastTimestamp)
		}
	} else {
		sequence = 0
	}

	lastTimestamp = timestamp

	// 返回拼接后的ID
	return ((timestamp - startTime) << timestampLeftShift) | (machineId << sequenceBits) | sequence
}

func getNextMillisecond(lastTimestamp int64) int64 {
	timestamp := time.Now().UnixNano() / 1e6
	for timestamp <= lastTimestamp {
		timestamp = time.Now().UnixNano() / 1e6
	}
	return timestamp
}

// 将雪花算法生成的ID转换为包含小写字母和数字的字符串
func encodeToBase36(id int64) string {
	b := big.NewInt(id)
	var result strings.Builder
	base := big.NewInt(62)
	for b.Sign() > 0 {
		mod := new(big.Int).Mod(b, base)
		result.WriteByte(charset[mod.Int64()])
		b.Div(b, base)
	}
	// 反转结果字符串
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
