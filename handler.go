package feng

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type requestType int

const (
	requestTypeRequest requestType = iota
	requestTypePush
	requestTypeRequestBack
	requestTypePushBack
)

type request struct {
	Route   string      `json:"route"`
	ID      string      `json:"id"`
	Type    requestType `json:"type"`
	Data    string      `json:"data"`
	Success bool        `json:"success"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handle(s *server, isSys bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		defer conn.Close()

		// 创建上下文
		u := &user{
			id:     uuid.New().String(),
			server: s,
			conn:   conn,
			isSys:  isSys,
		}
		r := &room{
			id:    uuid.New().String(),
			users: map[string]IUser{u.id: u},
		}
		u.room = r
		ctx := newContext(r, u, s)
		s.addUser(u, isSys)
		s.addRoom(r, isSys)

		// 主消息循环
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				s.config.Logger.Error("read message failed", "err", err)
				break
			}
			if msgType != websocket.TextMessage {
				continue
			}
			// 解析请求
			var req request
			err = s.config.Codec.Unmarshal(msg, &req)
			if err != nil {
				s.config.Logger.Error("unmarshal request failed", "err", err)
				continue
			}
			switch req.Type {
			case requestTypePushBack:
				err := handlePushBack(s, &req)
				if err != nil {
					s.config.Logger.Error("handle push back failed", "err", err)
				}
			case requestTypeRequestBack:
				err := handleRequestBack(ctx, s, &req, isSys)
				if err != nil {
					s.config.Logger.Error("handle request back failed", "err", err)
				}
			case requestTypePush, requestTypeRequest:
				err := handlePushOrRequest(ctx, s, &req, u, isSys)
				if err != nil {
					s.config.Logger.Error("handle push or request failed", "err", err)
				}
			default:
				s.config.Logger.Error("unknown request type", "type", req.Type)
			}
		}
	}
}

func handlePushBack(s *server, req *request) error {
	if !req.Success {
		return fmt.Errorf("push back failed, id: %s, data: %s", req.ID, req.Data)
	}
	return nil
}

func handleRequestBack(ctx IContext, s *server, req *request, isSys bool) error {
	resp, ok := s.getResponse(req.ID, isSys)
	if !ok {
		return fmt.Errorf("response not found, id: %s", req.ID)
	}
	// 删除响应防止内存泄漏
	close(resp.ch)
	s.deleteResponse(req.ID, isSys)
	// 处理响应
	if !req.Success {
		resp.ch <- chanData{
			Success: false,
			Data:    req.Data,
		}
		return nil
	}
	_, err := call(resp.fn, ctx, req.Data)
	if err != nil {
		resp.ch <- chanData{
			Success: false,
			Data:    err.Error(),
		}
		return nil
	}
	resp.ch <- chanData{
		Success: true,
	}
	return nil
}

func handlePushOrRequest(ctx IContext, s *server, req *request, u *user, isSys bool) error {
	resType := requestTypeRequestBack
	if req.Type == requestTypePush {
		resType = requestTypePushBack
	}
	// 中间件处理
	mutex := &s.userData.middlewaresLock
	middlewares := s.userData.middlewares
	if isSys {
		mutex = &s.sysData.middlewaresLock
		middlewares = s.sysData.middlewares
	}
	mutex.Lock()
	for _, middleware := range middlewares {
		if !strings.HasPrefix(req.Route, middleware.route) {
			continue
		}
		_, err := call(middleware.fn, ctx, req.Data)
		if err != nil {
			s.config.Logger.Error("call middleware func failed", "err", err)
			err = u.send(&request{
				ID:      req.ID,
				Type:    resType,
				Data:    fmt.Sprintf("middleware error: %s", err.Error()),
				Success: false,
			})
			if err != nil {
				s.config.Logger.Error("send middleware error failed", "err", err)
			}
			return nil
		}
	}
	mutex.Unlock()
	// 路由处理
	fn, ok := s.getRoute(req.Route, isSys)
	if !ok {
		s.config.Logger.Error("route not found", "route", req.Route)
		err := u.send(&request{
			ID:      req.ID,
			Type:    resType,
			Data:    "route not found",
			Success: false,
		})
		if err != nil {
			s.config.Logger.Error("send route not found failed", "err", err)
		}
		return nil
	}
	data, err := call(fn, ctx, req.Data)
	if err != nil {
		err = u.send(&request{
			ID:      req.ID,
			Type:    resType,
			Data:    fmt.Sprintf("route error: %s", err.Error()),
			Success: false,
		})
		if err != nil {
			s.config.Logger.Error("send route error failed", "err", err)
		}
		return nil
	}
	err = u.send(&request{
		ID:      req.ID,
		Type:    resType,
		Data:    data,
		Success: true,
	})
	if err != nil {
		s.config.Logger.Error("send route success failed", "err", err)
	}
	return nil
}
