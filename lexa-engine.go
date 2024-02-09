package main

import (
	"flag"
	"fmt"
	"lexa-engine/internal/config"
	"lexa-engine/internal/handler"
	"lexa-engine/internal/mqs"
	"lexa-engine/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/lexa-engine-api.yaml", "the config file")

func main() {
	var serverType string
	flag.StringVar(&serverType, "t", "", "服务启动类型")
	flag.Parse()
	if serverType == "consumer" {
		mqs.StartConsumerGroups()
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
