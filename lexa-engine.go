package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	// "os"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"gopkg.in/yaml.v3"

	"lexa-engine/internal/config"
	"lexa-engine/internal/handler"
	"lexa-engine/internal/logic/task"
	"lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/mqs"
	"lexa-engine/internal/svc"
)

var configFile = flag.String("f", "etc/lexa-engine-api.yaml", "the config file")

func main() {
	var serverType string
	var apiids string
	var sceneName string
	flag.StringVar(&serverType, "t", "", "服务启动类型")
	flag.StringVar(&apiids, "a", "", "执行的api列表")
	flag.StringVar(&sceneName, "s", "", "执行的场景名称")
	flag.Parse()
	if serverType == "consumer" {
		mqs.StartConsumerGroups()
	} else if serverType == "tc" {
		var apiidInt []int
		if len(strings.Split(apiids, ",")) == 1 {
			idInt, _ := strconv.ParseInt(apiids, 10, 64)
			apiidInt = append(apiidInt, int(idInt))
		} else {
			for _, apiid := range strings.Split(apiids, ",") {
				idInt, _ := strconv.ParseInt(apiid, 10, 64)
				apiidInt = append(apiidInt, int(idInt))
			}
		}

		tcGenerate(apiidInt, sceneName)
	} else {
		startHttpServer()
	}
}

func startHttpServer() {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	svcCtx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, svcCtx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

func tcGenerate(apiIDs []int, filename string) error {
	var err error
	mongourl := mongo.GetMongoUrl(config.MongoConfig{
		MongoPort:   27017,
		MongoUser:   "admin",
		MongoPasswd: "admin",
		UseDb:       "lct",
		MongoHost:   "47.120.49.73",
	})
	usedb := "lct"
	mod := apidetail.NewApidetailModel(mongourl, usedb, "ApiInfo")
	scenes := task.Scene{
		SceneId:     uuid.New().String(),
		Author:      "自动生成",
		Description: filename,
		EnvKey:      "test",
	}
	actions := []task.Action{}
	for _, apiID := range apiIDs {
		apid, err := mod.FindByApiId(context.Background(), apiID)
		if err != nil {
			logx.Errorf("Error getting API detail: %v", err)
			return err
		}

		action := task.Action{
			EnvKey:       "test",
			DomainKey:    getActionDomainKey("test", apid.ApiPath),
			RelateId:     apid.ApiId,
			SearchKey:    "",
			ActionID:     uuid.New().String(),
			ActionName:   apid.ApiName,
			ActionPath:   apid.ApiPath,
			ActionMethod: strings.ToUpper(apid.ApiMethod),
			Expect: task.Expect{
				Api: []task.Api{},
			},
			Headers:    make(map[string]string),
			Dependency: []task.Dependency{},
		}
		if apid.ApiResponse != nil && apid.ApiResponse.Fields != nil {
			for _, field := range apid.ApiResponse.Fields {
				action.Expect.Api = append(action.Expect.Api, task.Api{
					Type: "api",
					Data: task.Data{
						Name:      field.FieldPath,
						Operation: "equal",
						Type:      field.FieldType,
						Desire:    "",
					},
				})
			}
		}

		logx.Error(apid.ApiAuthType)
		if apid.ApiAuthType == "2" {
			refer := task.Refer{
				Type:     "header",
				DataType: "string",
				Target:   "Authorization",
			}
			action.Dependency = append(action.Dependency, task.Dependency{
				Refer:     refer,
				DataKey:   "data.token",
				ActionKey: "$sid.$aid",
				Type:      "1",
			})
		}

		if apid.ApiParameters != nil {
			for _, q := range apid.ApiParameters {
				if q == nil {
					continue
				}
				refer := task.Refer{
					Type:     "query",
					DataType: q.ValueType,
					Target:   fmt.Sprintf("url.query.$%s", q.QueryName),
				}
				action.Dependency = append(action.Dependency, task.Dependency{
					Refer:     refer,
					DataKey:   "query字段值",
					ActionKey: "",
					Type:      "3",
				})
			}
		}

		if apid.ApiPayload.ContentType == "application/x-www-form-urlencoded" {
			if apid.ApiPayload.FormPayload != nil {
				for fname, fvalue := range apid.ApiPayload.FormPayload {
					var refer task.Refer
					refer.DataType = "string"
					refer.Type = "payload"
					refer.Target = fmt.Sprintf("%s.%s", "payload", fname)
					action.Dependency = append(action.Dependency, task.Dependency{
						Refer:     refer,
						DataKey:   "$data." + fvalue,
						ActionKey: fmt.Sprintf("%s.%s", "$scenename", "$actionname"),
						Type:      "1",
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
					return err
				}
			}
		}
		action.Output = task.Output{
			Key: "$scenename.$actionname",
		}

		if apid.ApiMethod == "GET" {
			action.Headers["Content-Type"] = "application/json; charset=utf-8"
		} else {
			action.Headers["Content-Type"] = apid.ApiPayload.ContentType
		}

		if apid.ApiHeaders != nil {
			for _, header := range apid.ApiHeaders {
				action.Headers[header.HeaderName] = strings.ReplaceAll(header.HeaderValue, " ", "")
			}
		}

		action.Retry = 3
		action.Timeout = 5
		actions = append(actions, action)
	}
	scenes.Actions = actions

	// 使用yaml.Marshal函数将结构体转换为YAML格式
	data, err := yaml.Marshal(scenes)
	if err != nil {
		logx.Errorf("Error marshaling YAML: %v", err)
		return err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Printf("Error while writing to file: %v", err)
		return err
	}
	return err
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
