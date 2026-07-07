package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zmhuanf/feng/internal/core"
	"github.com/zmhuanf/feng/internal/session"
)

type Status struct {
	URL        string    `json:"url"`
	Load       int       `json:"load"`
	ID         string    `json:"id"`
	ReportTime time.Time `json:"reportTime"`
}

type Server struct {
	gin         *gin.Engine
	config      core.ServerConfig
	userData    *channelData
	systemData  *channelData
	status      Status
	statusLock  sync.RWMutex
	peers       map[string]*Status
	peersLock   sync.RWMutex
	httpServer  *http.Server
	serverMutex sync.Mutex
}

func New(config core.ServerConfig) core.Server {
	config = core.NormalizeServerConfig(config)
	s := &Server{
		config:     config,
		userData:   newChannelData(config),
		systemData: newChannelData(config),
		status: Status{
			URL:        config.Addr,
			Load:       0,
			ID:         uuid.New().String(),
			ReportTime: time.Now(),
		},
		peers: make(map[string]*Status),
	}
	s.addSystemHandlers()
	return s
}

func (s *Server) Config() *core.ServerConfig { return &s.config }

func (s *Server) Gin() *gin.Engine {
	if s.gin == nil {
		s.gin = gin.Default()
	}
	return s.gin
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	s.serverMutex.Lock()
	if s.httpServer != nil {
		s.serverMutex.Unlock()
		return errors.New("server already started")
	}
	engine := s.Gin()
	engine.GET("/game", s.handleWebsocket(false))
	engine.GET("/system", s.handleWebsocket(true))
	s.httpServer = &http.Server{Addr: fmt.Sprintf("%s:%d", s.config.Addr, s.config.Port), Handler: engine}
	server := s.httpServer
	s.serverMutex.Unlock()

	errCh := make(chan error, 1)
	go func() {
		if s.config.CertFile != "" && s.config.KeyFile != "" {
			errCh <- server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
			return
		}
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return s.Stop(context.Background())
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.serverMutex.Lock()
	server := s.httpServer
	s.httpServer = nil
	s.serverMutex.Unlock()
	if server == nil {
		return nil
	}
	return server.Shutdown(ctx)
}

func (s *Server) Handle(route string, handler any) error {
	return s.userData.router.Handle(route, handler)
}

func (s *Server) Use(route string, middleware any) error {
	return s.userData.router.Use(route, middleware)
}

func (s *Server) Room(id string) (core.Room, error) { return s.userData.rooms.Room(id) }

func (s *Server) Rooms() []core.Room { return s.userData.rooms.Rooms() }

func (s *Server) RoomsByPage(page int) []core.Room { return s.userData.rooms.RoomsByPage(page) }

func (s *Server) User(id string) (core.User, error) { return s.userData.users.User(id) }

func (s *Server) Users() []core.User { return s.userData.users.Users() }

func (s *Server) UsersByPage(page int) []core.User { return s.userData.users.UsersByPage(page) }

func (s *Server) channel(isSystem bool) *channelData {
	if isSystem {
		return s.systemData
	}
	return s.userData
}

func (s *Server) addUser(user *session.User, isSystem bool) {
	_ = s.channel(isSystem).users.Add(user)
}

func (s *Server) removeUser(id string, isSystem bool) {
	_ = s.channel(isSystem).users.Remove(id)
}
