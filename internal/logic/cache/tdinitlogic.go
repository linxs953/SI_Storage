package cache

import (
	"context"
	"encoding/json"
	"strings"

	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TdInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTdInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TdInitLogic {
	return &TdInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TdInitLogic) TdInit() (resp *types.TDInitResp, err error) {
	data := getInitData()
	ctx := context.Background()

	resp = &types.TDInitResp{
		Code:    0,
		Message: "测试数据初始化成功",
	}

	// 每次初始化先清空上一次的数据
	keys, err := l.svcCtx.RedisClient.KeysCtx(ctx, "td*")
	for _, key := range keys {
		_, err = l.svcCtx.RedisClient.DelCtx(context.Background(), key)
		if err != nil {
			logx.Error("Error deleting key:", err)
			resp.Code = 1
			resp.Message = err.Error()
			return
		}
	}

	for k, v := range data {
		// 处理缓存类型是list的数据
		if vmap, ok := v.([]map[string]interface{}); ok {
			for _, v := range vmap {
				vby, err := json.Marshal(v)
				if err != nil {
					resp.Code = 1
					resp.Message = err.Error()
					logx.Error(err)
					continue
				}
				_, err = l.svcCtx.RedisClient.LpushCtx(ctx, k, string(vby))
				if err != nil {
					resp.Code = 1
					resp.Message = err.Error()
					logx.Error(err)
					continue
				}
			}
		}
	}
	return
}

func getInitData() map[string]interface{} {
	td := make(map[string]interface{})
	td["td:user"] = getTestUsers()
	td["td:goods"] = getGoods()
	td["td:stores"] = getTestStores()
	return td
}

func getTestUsers() []map[string]interface{} {
	config := "15914795353:497:创客;15914795354:28:创客"
	users := []map[string]interface{}{}
	for _, u := range strings.Split(config, ";") {
		user := make(map[string]interface{})
		uinfo := strings.Split(u, ":")
		user["phone"] = uinfo[0]
		user["userid"] = uinfo[1]
		user["type"] = uinfo[2]
		users = append(users, user)
	}
	return users
}

func getTestStores() []map[string]interface{} {
	stores := []map[string]interface{}{}
	config := "1894:ffff:搜索123455:国内贸易:自营;1892:123Aa:测试1:国内贸易:自营"
	for _, s := range strings.Split(config, ";") {
		store := make(map[string]interface{})
		sinfo := strings.Split(s, ":")
		store["storeid"] = sinfo[0]
		store["storename"] = sinfo[2]
		store["storecode"] = sinfo[1]
		store["storetype"] = sinfo[4]
		store["tradetype"] = sinfo[3]
		stores = append(stores, store)
	}
	return stores
}

func getGoods() []map[string]interface{} {
	goods := []map[string]interface{}{}
	config := "111:4444:547:11:1:2:1:2"
	for _, g := range strings.Split(config, ";") {
		good := make(map[string]interface{})
		ginfo := strings.Split(g, ":")
		good["sku"] = ginfo[0]
		good["spu"] = ginfo[1]
		good["store"] = ginfo[2]
		good["count"] = ginfo[3]
		good["store_settle"] = ginfo[4]
		good["store_retail"] = ginfo[5]
		good["platform_settle"] = ginfo[6]
		good["platform_retail"] = ginfo[7]

		goods = append(goods, good)
	}
	return goods
}
