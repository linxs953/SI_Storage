package provider

import (
	logic "Storage/internal/logic/datahub/interface"
	"Storage/internal/svc"
	"context"
)

// Provider定义
type LogicProvider interface {
	// GetInterfaceLogic 获取接口相关的Logic
	GetInterfaceLogic() InterfaceLogicGroup

	// 其他的Logic
}

type InterfaceLogicGroup interface {
	GetDetail() *logic.GetInterfaceDetailLogic
	GetList() *logic.GetInterfaceListLogic
	Sync() *logic.SyncInterfaceLogic
	Delete() *logic.DeleteInterfaceLogic
}

type DefaultLogicProvider struct {
	svcCtx *svc.ServiceContext // 服务上下文
	ctx    context.Context     // 上下文
}

func NewDefaultLogicProvider(ctx context.Context, svcCtx *svc.ServiceContext) *DefaultLogicProvider {
	return &DefaultLogicProvider{
		svcCtx: svcCtx,
		ctx:    ctx,
	}
}

func (p *DefaultLogicProvider) GetInterfaceLogic() InterfaceLogicGroup {
	return &defaultInterfaceLogicGroup{
		svcCtx: p.svcCtx,
		ctx:    p.ctx,
	}
}

type defaultInterfaceLogicGroup struct {
	svcCtx *svc.ServiceContext
	ctx    context.Context
}

func (g *defaultInterfaceLogicGroup) GetDetail() *logic.GetInterfaceDetailLogic {
	return logic.NewGetInterfaceDetailLogic(g.ctx, g.svcCtx)
}

func (g *defaultInterfaceLogicGroup) GetList() *logic.GetInterfaceListLogic {
	return logic.NewGetInterfaceListLogic(g.ctx, g.svcCtx)
}

func (g *defaultInterfaceLogicGroup) Sync() *logic.SyncInterfaceLogic {
	return logic.NewSyncInterfaceLogic(g.ctx, g.svcCtx)
}

func (g *defaultInterfaceLogicGroup) Delete() *logic.DeleteInterfaceLogic {
	return logic.NewDeleteInterfaceLogic(g.ctx, g.svcCtx)
}
