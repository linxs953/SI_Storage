package scene

import (
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"

	"lexa-engine/internal/logic/scene"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

func GetSceneListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetSceneListDto
		if r.URL.Query().Get("page") != "" {
			var err error
			req.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
			if err != nil {
				req.Page = 1
			}
		}
		if r.URL.Query().Get("pageSize") != "" {
			var err error
			req.PageSize, err = strconv.Atoi(r.URL.Query().Get("pageSize"))
			if err != nil {
				req.PageSize = 10
			}
		}

		l := scene.NewGetSceneListLogic(r.Context(), svcCtx)
		resp, err := l.GetSceneList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
