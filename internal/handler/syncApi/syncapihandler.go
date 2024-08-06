package syncApi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"lexa-engine/internal/logic/syncApi"
	"lexa-engine/internal/logic/syncTask"
	"lexa-engine/internal/model/mongo/sync_task"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

func SyncApiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SyncDto
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		resp := &types.SyncVO{
			Code:    1,
			Message: "存在运行中的同步任务",
		}

		syncRecordLogic := syncTask.NewNewSyncRecordLogic(r.Context(), svcCtx)
		rs, err := findSyncRecord(syncRecordLogic)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		data, err := record2Map(rs)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if rs != nil {
			resp.Message = "存在进行中的任务"
			resp.Code = 1
			resp.Data = data
			httpx.OkJsonCtx(r.Context(), w, resp)
			return
		}

		logx.Debug("无进行中的同步记录")
		if err = syncRecordLogic.NewSyncRecord(); err != nil {
			logx.Error("创建同步记录失败")
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		rs, err = findSyncRecord(syncRecordLogic)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if rs == nil {
			logx.Error("无进行中的同步记录")
			err = errors.New("无进行中的同步记录")
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		data, err = record2Map(rs)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		resp.Code = 0
		resp.Message = "创建同步任务成功"
		resp.Data = data

		newCtx := context.WithValue(r.Context(), "SyncRecordID", rs.ID)

		l := syncApi.NewSyncApiLogic(newCtx, svcCtx)

		resp, err = l.SyncApi(&req)
		resp.Data = data
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}

	}
}

func record2Map(v any) (map[string]any, error) {
	rs := make(map[string]any)
	var err error
	bts, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bts, &rs); err != nil {
		return nil, err
	}
	return rs, nil

}

func findSyncRecord(syncRecordLogic *syncTask.NewSyncRecordLogic) (rs *sync_task.Synctask, err error) {
	rs, err = syncRecordLogic.FindSyncRecord()
	if err != nil {
		logx.Error("查找同步中的记录失败")
		return
	}
	return
}
