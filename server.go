package feng

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
)

type IServer interface {
	GetConfig() *serverConfig

	Start() error
	Stop() error
	AddHandler(string, any) error
}

type server struct {
	// 配置
	config *serverConfig
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

type middleware struct {
	route string
	fn    func(IContext, any) error
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
		config:      config,
		route:       make(map[string]any),
		middlewares: make([]*middleware, 0),
		responses:   make(map[string]*response),
		users:       make(map[string]*user),
		rooms:       make(map[string]*room),
	}
}

func (s *server) GetConfig() *serverConfig {
	return s.config
}

func (s *server) Start() error {
	r := gin.Default()

	r.GET("/", handle(s))

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
	s.routeLock.Lock()
	defer s.routeLock.Unlock()
	// 检查函数签名
	ft := reflect.TypeOf(fn)
	if ft.Kind() != reflect.Func {
		return errors.New("f must be func")
	}
	if ft.NumIn() != 2 {
		return errors.New("func must have 2 args")
	}
	if ft.In(0) != reflect.TypeOf((*IContext)(nil)).Elem() {
		return errors.New("first arg must be IContext")
	}
	// 保存
	s.route[route] = fn
	return nil
}

func (s *server) AddMiddleware(name string, fn func(IContext, any) error) {
	s.middlewaresLock.Lock()
	defer s.middlewaresLock.Unlock()
	s.middlewares = append(s.middlewares, &middleware{
		route: name,
		fn:    fn,
	})
}

func (s *server) getResponse(id string) (*response, bool) {
	s.responsesLock.RLock()
	defer s.responsesLock.RUnlock()
	res, ok := s.responses[id]
	return res, ok
}

func (s *server) deleteResponse(id string) {
	s.responsesLock.Lock()
	defer s.responsesLock.Unlock()
	delete(s.responses, id)
}

func (s *server) addResponse(id string, res *response) {
	s.responsesLock.Lock()
	defer s.responsesLock.Unlock()
	s.responses[id] = res
}

func (s *server) getRoute(route string) (any, bool) {
	s.routeLock.RLock()
	defer s.routeLock.RUnlock()
	handler, ok := s.route[route]
	return handler, ok
}
