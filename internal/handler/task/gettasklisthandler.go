package task

import (
	"net/http"
	"strconv"

	"lexa-engine/internal/logic/task"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetTaskListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTaskListDto
		pageSize := r.URL.Query().Get("pageSize")
		page := r.URL.Query().Get("page")
		if pageSize == "" {
			req.PageSize = 10
		} else {
			pageSize, err := strconv.ParseInt(pageSize, 10, 64)
			if err != nil {
				req.PageSize = 10
			}
			req.PageSize = int(pageSize)
		}

		if page == "" {
			req.PageNum = 1
		} else {
			page, err := strconv.ParseInt(page, 10, 64)
			if err != nil {
				req.PageNum = 1
			}
			req.PageNum = int(page)
		}

		l := task.NewGetTaskListLogic(r.Context(), svcCtx)
		resp, err := l.GetTaskList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
