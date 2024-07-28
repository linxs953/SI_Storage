package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"lexa-engine/internal/logic"
	"lexa-engine/internal/svc"
)

func indexHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewIndexLogic(r.Context(), svcCtx)
		resp, err := l.Index()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
