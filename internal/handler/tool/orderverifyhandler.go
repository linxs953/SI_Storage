package tool

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"lexa-engine/internal/logic/tool"
	"lexa-engine/internal/svc"
)

func OrderVerifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := tool.NewOrderVerifyLogic(r.Context(), svcCtx)
		err := l.OrderVerify()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
