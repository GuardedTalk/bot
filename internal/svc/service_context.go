package svc

import (
	"github.com/GuardedTalk/bot/internal/config"
	lksdk "github.com/livekit/server-sdk-go"
)

type ServiceContext struct {
	Config config.Config
	Bot    map[string]*lksdk.Room
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Bot:    make(map[string]*lksdk.Room),
	}
}
