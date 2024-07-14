package mqs

import (
	"context"
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"

	"lexa-engine/internal/config"
	"lexa-engine/internal/svc"
)

func StartConsumerGroups() {
	var mqConfigFile = flag.String("m", "etc/lexa-engine-api.yaml", "the config file")
	var c config.Config
	conf.MustLoad(*mqConfigFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	svcCtx := svc.NewServiceContext(c)
	ctx := context.Background()
	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	for _, mq := range Consumers(c, ctx, svcCtx) {
		serviceGroup.Add(mq)
	}
	logx.Info("Start  Consumer Group Successfully")
	serviceGroup.Start()
}
