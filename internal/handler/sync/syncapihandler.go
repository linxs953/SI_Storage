package sync

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.mongodb.org/mongo-driver/bson/primitive"

	apisync "lexa-engine/internal/logic/sync"
	"lexa-engine/internal/logic/syncTask"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

func SyncapiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StartDto
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		syncRecordLogic := syncTask.NewNewSyncRecordLogic(r.Context(), svcCtx)
		l := apisync.NewSyncapiLogic(r.Context(), svcCtx)
		rs, err := syncRecordLogic.FindSyncRecord()
		if err != nil {
			return
		}
		if rs == nil {
			logx.Debug("无进行中的同步记录")
			if err = syncRecordLogic.NewSyncRecord(); err != nil {
				return
			}
			rs, err = syncRecordLogic.FindSyncRecord()
			if err != nil {
				return
			}

			resp, err := l.Syncapi(&req, primitive.NewObjectID())
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
			} else {
				resp.Data = rs
				httpx.OkJsonCtx(r.Context(), w, resp)
			}
		} else {
			resp := types.StartResp{
				Code:    1,
				Message: "存在进行中的同步任务",
				Data:    rs,
			}
			httpx.OkJsonCtx(r.Context(), w, resp)
		}

	}
}
