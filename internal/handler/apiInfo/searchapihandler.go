package apiInfo

import (
	"net/http"

	api "lexa-engine/internal/logic/apiInfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SearchApiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchDto
		sk := r.URL.Query().Get("keyword")
		req.Keyword = sk
		l := api.NewSearchApiLogic(r.Context(), svcCtx)
		resp, err := l.SearchApi(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
