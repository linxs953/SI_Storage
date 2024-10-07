package task

import (
	"errors"
	"lexa-engine/internal/logic/task"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteTaskRunRecordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteTaskRunRecordDto
		// if err := httpx.Parse(r, &req); err != nil {
		// 	httpx.ErrorCtx(r.Context(), w, err)
		// 	return
		// }
		req.ExecId = r.URL.Query().Get("execId")
		if req.ExecId == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("execId is empty"))
			return
		}
		l := task.NewDeleteTaskRunRecordLogic(r.Context(), svcCtx)
		resp, err := l.DeleteTaskRunRecord(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
