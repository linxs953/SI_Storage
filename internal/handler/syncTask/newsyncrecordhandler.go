package syncTask

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"lexa-engine/internal/logic/syncTask"
	"lexa-engine/internal/svc"
)

func NewSyncRecordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := syncTask.NewNewSyncRecordLogic(r.Context(), svcCtx)
		err := l.NewSyncRecord()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
