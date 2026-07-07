package server

import (
	"reflect"

	"github.com/zmhuanf/feng/internal/core"
	"github.com/zmhuanf/feng/internal/pending"
	"github.com/zmhuanf/feng/internal/router"
	"github.com/zmhuanf/feng/internal/session"
)

type channelData struct {
	router  *router.Router
	pending *pending.Store
	users   *session.UserStore
	rooms   *session.RoomStoreImpl
}

func newChannelData(config core.ServerConfig) *channelData {
	return &channelData{
		router:  router.New(reflect.TypeFor[core.ServerContext]()),
		pending: pending.New(config.Timeout),
		users:   session.NewUserStore(config.PageSize),
		rooms:   session.NewRoomStore(config.PageSize),
	}
}
