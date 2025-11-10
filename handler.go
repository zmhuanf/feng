package feng

import (
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
	requestTypeSystem
)

type request struct {
	Route   string      `json:"route"`
	ID      string      `json:"id"`
	Type    requestType `json:"type"`
	Data    []byte      `json:"data"`
	Success bool        `json:"success"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handle(s *server) func(c *gin.Context) {
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
		}
		r := &room{
			id:    uuid.New().String(),
			users: map[string]IUser{u.id: u},
		}
		u.room = r
		ctx := newContext(r, u, s)

		// 主消息循环
	MAINFOR:
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if msgType != websocket.TextMessage {
				continue
			}
			// 解析请求
			var req request
			err = s.config.Codec.Unmarshal(msg, &req)
			if err != nil {
				continue
			}
			// 回复的消息
			if req.Type == requestTypePushBack {
				continue
			}
			if req.Type == requestTypeRequestBack {
				resp, ok := s.getResponse(req.ID)
				if !ok {
					continue
				}
				if !req.Success {
					resp.ch <- chanData{
						Success: false,
						Data:    req.Data,
					}
					continue
				}
				_, err = call(resp.fn, ctx, req.Data)
				if err != nil {
					resp.ch <- chanData{
						Success: false,
						Data:    err.Error(),
					}
					continue
				}
				resp.ch <- chanData{
					Success: true,
				}
				continue
			}

			// 请求的消息
			resType := requestTypeRequestBack
			if req.Type == requestTypePush {
				resType = requestTypePushBack
			}
			// 中间件处理
			s.middlewaresLock.Lock()
			for _, middleware := range s.middlewares {
				if !strings.HasPrefix(req.Route, middleware.route) {
					continue
				}
				_, err = call(middleware.fn, ctx, req.Data)
				if err != nil {
					s.config.Logger.Error("call middleware func failed", "err", err)
					err = u.send(&request{
						ID:      req.ID,
						Type:    resType,
						Data:    []byte(err.Error()),
						Success: false,
					})
					if err != nil {
						s.config.Logger.Error("send middleware error failed", "err", err)
					}
					continue MAINFOR
				}
			}
			s.middlewaresLock.Unlock()
			// 路由处理
			fn, ok := s.getRoute(req.Route)
			if !ok {
				err = u.send(&request{
					ID:      req.ID,
					Type:    resType,
					Data:    []byte("route not found"),
					Success: false,
				})
				if err != nil {
					s.config.Logger.Error("send route not found failed", "err", err)
				}
				continue
			}
			data, err := call(fn, ctx, req.Data)
			if err != nil {
				err = u.send(&request{
					ID:      req.ID,
					Type:    resType,
					Data:    []byte(err.Error()),
					Success: false,
				})
				if err != nil {
					s.config.Logger.Error("send route error failed", "err", err)
				}
				continue
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
		}
	}
}
