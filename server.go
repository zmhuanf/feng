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
	AddHandler(route string, callback any) error
	AddMiddleware(route string, middleware any) error

	GetRoom(id string) (IRoom, error)
	GetAllRooms() []IRoom
	GetRoomByPage(page int) []IRoom
	GetUser(id string) (IUser, error)
	GetAllUsers() []IUser
	GetUsersByPage(page int) []IUser
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
	userData *serverData
	// 系统数据
	sysData *serverData
	// 服务器状态
	status serverStatus
	// 状态锁
	statusLock sync.RWMutex
	// 其他服务器状态
	otherStatus map[string]*serverStatus
	// 其他服务器状态锁
	otherStatusLock sync.RWMutex
}

type chanData struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

func NewServer(config *serverConfig) IServer {
	return &server{
		config:   config,
		userData: newServerData(config),
		sysData:  newServerData(config),
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
	err := checkFuncType(fn, true)
	if err != nil {
		return err
	}

	lock := &s.userData.routeLock
	routeMap := s.userData.route
	if isSys {
		lock = &s.sysData.routeLock
		routeMap = s.sysData.route
	}

	lock.Lock()
	defer lock.Unlock()
	routeMap[route] = fn
	return nil
}

func (s *server) AddMiddleware(name string, fn any) error {
	return s.addMiddleware(name, fn, false)
}

func (s *server) addMiddleware(route string, fn any, isSys bool) error {
	// 检查函数签名
	err := checkFuncType(fn, true)
	if err != nil {
		return err
	}

	lock := &s.userData.middlewaresLock
	middlewares := &s.userData.middlewares
	if isSys {
		lock = &s.sysData.middlewaresLock
		middlewares = &s.sysData.middlewares
	}

	lock.Lock()
	defer lock.Unlock()
	*middlewares = append(*middlewares, &middleware{
		route: route,
		fn:    fn,
	})
	return nil
}

func (s *server) getResponse(id string, isSys bool) (*response, bool) {
	lock := &s.userData.responsesLock
	responses := s.userData.responses
	if isSys {
		lock = &s.sysData.responsesLock
		responses = s.sysData.responses
	}

	lock.RLock()
	defer lock.RUnlock()
	res, ok := responses[id]
	return res, ok
}

func (s *server) deleteResponse(id string, isSys bool) {
	lock := &s.userData.responsesLock
	responses := s.userData.responses
	if isSys {
		lock = &s.sysData.responsesLock
		responses = s.sysData.responses
	}

	lock.Lock()
	defer lock.Unlock()
	delete(responses, id)
}

func (s *server) addResponse(id string, res *response, isSys bool) {
	lock := &s.userData.responsesLock
	responses := s.userData.responses
	if isSys {
		lock = &s.sysData.responsesLock
		responses = s.sysData.responses
	}

	lock.Lock()
	defer lock.Unlock()
	responses[id] = res
}

func (s *server) getRoute(route string, isSys bool) (any, bool) {
	lock := &s.userData.routeLock
	routeMap := s.userData.route
	if isSys {
		lock = &s.sysData.routeLock
		routeMap = s.sysData.route
	}

	lock.RLock()
	defer lock.RUnlock()
	handler, ok := routeMap[route]
	return handler, ok
}

func (s *server) GetRoom(id string) (IRoom, error) {
	return s.getRoom(id, false)
}

func (s *server) getRoom(id string, isSys bool) (IRoom, error) {
	manage := s.userData.roomManage
	if isSys {
		manage = s.sysData.roomManage
	}

	return manage.getRoom(id)
}

func (s *server) GetAllRooms() []IRoom {
	return s.getAllRooms(false)
}

func (s *server) getAllRooms(isSys bool) []IRoom {
	manage := s.userData.roomManage
	if isSys {
		manage = s.sysData.roomManage
	}

	return manage.getAllRooms()
}

func (s *server) GetRoomByPage(page int) []IRoom {
	return s.getRoomsByPage(page, false)
}

func (s *server) getRoomsByPage(page int, isSys bool) []IRoom {
	manage := s.userData.roomManage
	if isSys {
		manage = s.sysData.roomManage
	}

	return manage.getRoomsByPage(page)
}

func (s *server) GetUser(id string) (IUser, error) {
	return s.getUser(id, false)
}

func (s *server) getUser(id string, isSys bool) (IUser, error) {
	manage := s.userData.userManage
	if isSys {
		manage = s.sysData.userManage
	}

	return manage.getUser(id)
}

func (s *server) GetAllUsers() []IUser {
	return s.getAllUsers(false)
}

func (s *server) getAllUsers(isSys bool) []IUser {
	manage := s.userData.userManage
	if isSys {
		manage = s.sysData.userManage
	}

	return manage.getAllUsers()
}

func (s *server) GetUsersByPage(page int) []IUser {
	return s.getUsersByPage(page, false)
}

func (s *server) getUsersByPage(page int, isSys bool) []IUser {
	manage := s.userData.userManage
	if isSys {
		manage = s.sysData.userManage
	}

	return manage.getUsersByPage(page)
}

func (s *server) addRoom(r *room, isSys bool) error {
	manage := s.userData.roomManage
	if isSys {
		manage = s.sysData.roomManage
	}

	return manage.addRoom(r)
}

func (s *server) removeRoom(id string, isSys bool) error {
	manage := s.userData.roomManage
	if isSys {
		manage = s.sysData.roomManage
	}

	return manage.removeRoom(id)
}

func (s *server) addUser(u *user, isSys bool) {
	manage := s.userData.userManage
	if isSys {
		manage = s.sysData.userManage
	}

	manage.addUser(u)
}

func (s *server) removeUser(id string, isSys bool) {
	manage := s.userData.userManage
	if isSys {
		manage = s.sysData.userManage
	}

	manage.removeUser(id)
}

func (s *server) addSystemHandler() {
	s.addHandler("/join", systemJoin, true)
	s.addHandler("/report_status", systemReportStatus, true)
	s.addHandler("/get_low_load_server_addr", systemGetLowLoadServerAddr, true)
}
