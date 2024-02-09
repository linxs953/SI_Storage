package job

import (
	apijob "lexa-engine/internal/logic/job"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func StartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StartDto
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := apijob.NewStartLogic(r.Context(), svcCtx)
		resp, err := l.Start(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
