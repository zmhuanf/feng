package feng

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IServer interface {
	GetConfig() *serverConfig

	Start() error
	Stop() error
	AddHandler(string, any) error
	GetRoom(id string) (IRoom, error)
	GetAllRooms() []IRoom
	GetUser(id string) (IUser, error)
	GetAllUsers() []IUser
}

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
	// 用户列表
	users map[string]*user
	// 用户锁
	usersLock sync.RWMutex
	// 房间列表
	rooms map[string]*room
	// 房间锁
	roomsLock sync.RWMutex
}

type serverStatus struct {
	// 地址
	Url string `json:"url"`
	// 负载
	Load int `json:"load"`
	// id
	ID string `json:"id"`
	// 上报时间
	ReportTime time.Time `json:"reportTime"`
}

type server struct {
	// 配置
	config *serverConfig
	// 用户数据
	userData serverData
	// 系统数据
	sysData serverData
	// 服务器状态
	status serverStatus
	// 状态锁
	statusLock sync.RWMutex
	// 其他服务器状态
	otherStatus map[string]*serverStatus
	// 其他服务器状态锁
	otherStatusLock sync.RWMutex
}

type middleware struct {
	route string
	fn    any
}

type response struct {
	fn any
	ch chan chanData
}

type chanData struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

func NewServer(config *serverConfig) IServer {
	return &server{
		config: config,
		userData: serverData{
			route:       make(map[string]any),
			middlewares: make([]*middleware, 0),
			responses:   make(map[string]*response),
			users:       make(map[string]*user),
			rooms:       make(map[string]*room),
		},
		sysData: serverData{
			route:       make(map[string]any),
			middlewares: make([]*middleware, 0),
			responses:   make(map[string]*response),
			users:       make(map[string]*user),
			rooms:       make(map[string]*room),
		},
		status: serverStatus{
			Url:        config.Addr,
			Load:       0,
			ID:         uuid.New().String(),
			ReportTime: time.Now(),
		},
		otherStatus: make(map[string]*serverStatus),
	}
}

func (s *server) GetConfig() *serverConfig {
	return s.config
}

func (s *server) Start() error {
	r := gin.Default()

	r.GET("/game", serverHandle(s, false))
	r.GET("/system", serverHandle(s, true))

	s.addSystemHandler()

	ser := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.Addr, s.config.Port),
		Handler: r,
	}
	if s.config.CertFile != "" && s.config.KeyFile != "" {
		err := ser.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
		if err != nil {
			return err
		}
	} else {
		err := ser.ListenAndServe()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *server) Stop() error {
	return nil
}

func (s *server) AddHandler(route string, fn any) error {
	return s.addHandler(route, fn, false)
}

func (s *server) addHandler(route string, fn any, isSys bool) error {
	// 检查函数签名
	err := checkFuncTypeServer(fn)
	if err != nil {
		return err
	}
	if isSys {
		s.sysData.routeLock.Lock()
		defer s.sysData.routeLock.Unlock()
		s.sysData.route[route] = fn
	} else {
		s.userData.routeLock.Lock()
		defer s.userData.routeLock.Unlock()
		s.userData.route[route] = fn
	}
	return nil
}

func (s *server) AddMiddleware(name string, fn func(IServerContext, any) error) error {
	return s.addMiddleware(name, fn, false)
}

func (s *server) addMiddleware(route string, fn func(IServerContext, any) error, isSys bool) error {
	// 检查函数签名
	err := checkFuncTypeServer(fn)
	if err != nil {
		return err
	}
	if isSys {
		s.sysData.middlewaresLock.Lock()
		defer s.sysData.middlewaresLock.Unlock()
		s.sysData.middlewares = append(s.sysData.middlewares, &middleware{
			route: route,
			fn:    fn,
		})
	} else {
		s.userData.middlewaresLock.Lock()
		defer s.userData.middlewaresLock.Unlock()
		s.userData.middlewares = append(s.userData.middlewares, &middleware{
			route: route,
			fn:    fn,
		})
	}
	return nil
}

func (s *server) getResponse(id string, isSys bool) (*response, bool) {
	if isSys {
		s.sysData.responsesLock.RLock()
		defer s.sysData.responsesLock.RUnlock()
		res, ok := s.sysData.responses[id]
		return res, ok
	} else {
		s.userData.responsesLock.RLock()
		defer s.userData.responsesLock.RUnlock()
		res, ok := s.userData.responses[id]
		return res, ok
	}
}

func (s *server) deleteResponse(id string, isSys bool) {
	if isSys {
		s.sysData.responsesLock.Lock()
		defer s.sysData.responsesLock.Unlock()
		delete(s.sysData.responses, id)
	} else {
		s.userData.responsesLock.Lock()
		defer s.userData.responsesLock.Unlock()
		delete(s.userData.responses, id)
	}
}

func (s *server) addResponse(id string, res *response, isSys bool) {
	if isSys {
		s.sysData.responsesLock.Lock()
		defer s.sysData.responsesLock.Unlock()
		s.sysData.responses[id] = res
	} else {
		s.userData.responsesLock.Lock()
		defer s.userData.responsesLock.Unlock()
		s.userData.responses[id] = res
	}
}

func (s *server) getRoute(route string, isSys bool) (any, bool) {
	if isSys {
		s.sysData.routeLock.RLock()
		defer s.sysData.routeLock.RUnlock()
		handler, ok := s.sysData.route[route]
		return handler, ok
	} else {
		s.userData.routeLock.RLock()
		defer s.userData.routeLock.RUnlock()
		handler, ok := s.userData.route[route]
		return handler, ok
	}
}

func (s *server) GetRoom(id string) (IRoom, error) {
	return s.getRoom(id, false)
}

func (s *server) getRoom(id string, isSys bool) (IRoom, error) {
	if isSys {
		s.sysData.roomsLock.RLock()
		defer s.sysData.roomsLock.RUnlock()
		room, ok := s.sysData.rooms[id]
		if !ok {
			return nil, fmt.Errorf("room %s not found", id)
		}
		return room, nil
	} else {
		s.userData.roomsLock.RLock()
		defer s.userData.roomsLock.RUnlock()
		room, ok := s.userData.rooms[id]
		if !ok {
			return nil, fmt.Errorf("room %s not found", id)
		}
		return room, nil
	}
}

func (s *server) GetAllRooms() []IRoom {
	return s.getAllRooms(false)
}

func (s *server) getAllRooms(isSys bool) []IRoom {
	if isSys {
		s.sysData.roomsLock.RLock()
		defer s.sysData.roomsLock.RUnlock()
		rooms := make([]IRoom, 0, len(s.sysData.rooms))
		for _, room := range s.sysData.rooms {
			rooms = append(rooms, room)
		}
		return rooms
	} else {
		s.userData.roomsLock.RLock()
		defer s.userData.roomsLock.RUnlock()
		rooms := make([]IRoom, 0, len(s.userData.rooms))
		for _, room := range s.userData.rooms {
			rooms = append(rooms, room)
		}
		return rooms
	}
}

func (s *server) GetUser(id string) (IUser, error) {
	return s.getUser(id, false)
}

func (s *server) getUser(id string, isSys bool) (IUser, error) {
	if isSys {
		s.sysData.usersLock.RLock()
		defer s.sysData.usersLock.RUnlock()
		user, ok := s.sysData.users[id]
		if !ok {
			return nil, fmt.Errorf("user %s not found", id)
		}
		return user, nil
	} else {
		s.userData.usersLock.RLock()
		defer s.userData.usersLock.RUnlock()
		user, ok := s.userData.users[id]
		if !ok {
			return nil, fmt.Errorf("user %s not found", id)
		}
		return user, nil
	}
}

func (s *server) GetAllUsers() []IUser {
	return s.getAllUsers(false)
}

func (s *server) getAllUsers(isSys bool) []IUser {
	if isSys {
		s.sysData.usersLock.RLock()
		defer s.sysData.usersLock.RUnlock()
		users := make([]IUser, 0, len(s.sysData.users))
		for _, user := range s.sysData.users {
			users = append(users, user)
		}
		return users
	} else {
		s.userData.usersLock.RLock()
		defer s.userData.usersLock.RUnlock()
		users := make([]IUser, 0, len(s.userData.users))
		for _, user := range s.userData.users {
			users = append(users, user)
		}
		return users
	}
}

func (s *server) addRoom(r IRoom, isSys bool) {
	if isSys {
		s.sysData.roomsLock.Lock()
		defer s.sysData.roomsLock.Unlock()
		s.sysData.rooms[r.GetID()] = r.(*room)
	} else {
		s.userData.roomsLock.Lock()
		defer s.userData.roomsLock.Unlock()
		s.userData.rooms[r.GetID()] = r.(*room)
	}
}

func (s *server) addUser(u IUser, isSys bool) {
	if isSys {
		s.sysData.usersLock.Lock()
		defer s.sysData.usersLock.Unlock()
		s.sysData.users[u.GetID()] = u.(*user)
	} else {
		s.userData.usersLock.Lock()
		defer s.userData.usersLock.Unlock()
		s.userData.users[u.GetID()] = u.(*user)
	}
}

func (s *server) addSystemHandler() {
	s.addHandler("/join", systemJoin, true)
	s.addHandler("/report_status", systemReportStatus, true)
	s.addHandler("/get_low_load_server_addr", systemGetLowLoadServerAddr, true)
}
