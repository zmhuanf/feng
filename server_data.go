package feng

import "sync"

type serverData struct {
	// 路由
	route map[string]any
	// 路由锁
	routeLock sync.RWMutex
	// 中间层
	middlewares []*middleware
	// 中间层锁
	middlewaresLock sync.RWMutex
	// 响应
	responses map[string]*response
	// 响应锁
	responsesLock sync.RWMutex
	// 用户管理
	userManage *userManage
	// 房间管理
	roomManage *roomManage
}

func newServerData(config *serverConfig) *serverData {
	return &serverData{
		route:       make(map[string]any),
		middlewares: make([]*middleware, 0),
		responses:   make(map[string]*response),
		userManage:  newUserManage(config),
		roomManage:  newRoomManage(config),
	}
}

type middleware struct {
	route string
	fn    any
}

type response struct {
	fn any
	ch chan chanData
}
