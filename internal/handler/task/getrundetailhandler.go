package task

import (
	"errors"
	"lexa-engine/internal/logic/task"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetRunDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetRunDetailDto
		// if err := httpx.Parse(r, &req); err != nil {
		// 	httpx.ErrorCtx(r.Context(), w, err)
		// 	return
		// }
		req.TaskId = r.URL.Query().Get("taskId")
		req.ExecId = r.URL.Query().Get("execId")
		if req.TaskId == "" || req.ExecId == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("taskId or execId is empty"))
			return
		}
		l := task.NewGetRunDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetRunDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
