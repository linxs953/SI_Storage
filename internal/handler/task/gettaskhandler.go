package task

import (
	"errors"
	"lexa-engine/internal/logic/task"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetTaskHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTaskDto

		req.TaskId = r.URL.Query().Get("taskId")
		if req.TaskId == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("taskId为空"))
			return
		}
		l := task.NewGetTaskLogic(r.Context(), svcCtx)
		resp, err := l.GetTask(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
