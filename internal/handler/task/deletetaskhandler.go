package task

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"lexa-engine/internal/logic/task"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

func DeleteTaskHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteTaskDto

		req.TaskId = r.URL.Query().Get("taskId")
		if req.TaskId == "" {
			httpx.ErrorCtx(r.Context(), w, http.ErrMissingRequiredValue)
			return
		}
		l := task.NewDeleteTaskLogic(r.Context(), svcCtx)
		resp, err := l.DeleteTask(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
