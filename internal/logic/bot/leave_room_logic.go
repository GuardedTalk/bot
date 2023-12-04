package bot

import (
	"context"

	"github.com/GuardedTalk/bot/internal/svc"
	"github.com/GuardedTalk/bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LeaveRoomLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLeaveRoomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LeaveRoomLogic {
	return &LeaveRoomLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LeaveRoomLogic) LeaveRoom(req *types.RoomReq) (resp *types.BaseMsgResp, err error) {
	// todo: add your logic here and delete this line
	l.svcCtx.Bot[req.RoomID].Disconnect()
	return
}
