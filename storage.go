package main

import (
	"flag"
	"fmt"

	"Storage/internal/config"
	executeservice "Storage/internal/server/executeservice"
	generateservice "Storage/internal/server/generateservice"
	interfaceservice "Storage/internal/server/interfaceservice"
	reportservice "Storage/internal/server/reportservice"
	sceneconfigservice "Storage/internal/server/sceneconfigservice"
	taskconfigservice "Storage/internal/server/taskconfigservice"
	testdataservice "Storage/internal/server/testdataservice"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/storage.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	s := zrpc.MustNewServer(
		c.RpcServerConf,
		func(grpcServer *grpc.Server) {
			// 注册所有服务
			storage.RegisterExecuteServiceServer(grpcServer, executeservice.NewExecuteServiceServer(ctx))
			storage.RegisterGenerateServiceServer(grpcServer, generateservice.NewGenerateServiceServer(ctx))
			storage.RegisterInterfaceServiceServer(grpcServer, interfaceservice.NewInterfaceServiceServer(ctx))
			storage.RegisterReportServiceServer(grpcServer, reportservice.NewReportServiceServer(ctx))
			storage.RegisterSceneConfigServiceServer(grpcServer, sceneconfigservice.NewSceneConfigServiceServer(ctx))
			storage.RegisterTaskConfigServiceServer(grpcServer, taskconfigservice.NewTaskConfigServiceServer(ctx))
			storage.RegisterTestDataServiceServer(grpcServer, testdataservice.NewTestDataServiceServer(ctx))

			if c.Mode == service.DevMode || c.Mode == service.TestMode {
				reflection.Register(grpcServer)
			}
			// reflection.Register(grpcServer)
		},
	)
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
