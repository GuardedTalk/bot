package bot

import (
	"net/http"

	"github.com/GuardedTalk/bot/internal/logic/bot"
	"github.com/GuardedTalk/bot/internal/svc"
	"github.com/GuardedTalk/bot/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func LeaveRoomHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RoomReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := bot.NewLeaveRoomLogic(r.Context(), svcCtx)
		resp, err := l.LeaveRoom(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
