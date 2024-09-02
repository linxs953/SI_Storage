package apiInfo

import (
	"net/http"

	api "lexa-engine/internal/logic/apiInfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetApiDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApiDetailDto
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := api.NewGetApiDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetApiDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
