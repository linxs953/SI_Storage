// Code generated by goctl. DO NOT EDIT.
// goctl 1.7.6
// Source: Storage.proto

package server

import (
	"context"

	"Storage/internal/logic/interfaceservice"
	"Storage/internal/svc"
	"Storage/storage"
)

type InterfaceServiceServer struct {
	svcCtx *svc.ServiceContext
	storage.UnimplementedInterfaceServiceServer
}

func NewInterfaceServiceServer(svcCtx *svc.ServiceContext) *InterfaceServiceServer {
	return &InterfaceServiceServer{
		svcCtx: svcCtx,
	}
}

// 接口同步
func (s *InterfaceServiceServer) GetInterfaceList(ctx context.Context, in *storage.Empty) (*storage.GetInterfaceListResponse, error) {
	l := interfaceservicelogic.NewGetInterfaceListLogic(ctx, s.svcCtx)
	return l.GetInterfaceList(in)
}

func (s *InterfaceServiceServer) GetInterfaceDetail(ctx context.Context, in *storage.GetInterfaceRequest) (*storage.GetInterfaceResponse, error) {
	l := interfaceservicelogic.NewGetInterfaceDetailLogic(ctx, s.svcCtx)
	return l.GetInterfaceDetail(in)
}

func (s *InterfaceServiceServer) DeleteInterface(ctx context.Context, in *storage.DeleteInterfaceRequest) (*storage.DeleteResponse, error) {
	l := interfaceservicelogic.NewDeleteInterfaceLogic(ctx, s.svcCtx)
	return l.DeleteInterface(in)
}

func (s *InterfaceServiceServer) SyncInterface(ctx context.Context, in *storage.SyncInterfaceRequest) (*storage.SyncInterfaceResponse, error) {
	l := interfaceservicelogic.NewSyncInterfaceLogic(ctx, s.svcCtx)
	return l.SyncInterface(in)
}
