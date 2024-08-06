package scene

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"lexa-engine/internal/logic/scene"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

func SearchScenesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchScenesDto
		// if err := httpx.Parse(r, &req); err != nil {
		// 	httpx.ErrorCtx(r.Context(), w, err)
		// 	return
		// }
		keyword := r.URL.Query().Get("keyword")
		if keyword == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("keyword is empty"))
			return
		}
		req.Keyword = keyword

		l := scene.NewSearchScenesLogic(r.Context(), svcCtx)
		resp, err := l.SearchScenes(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
