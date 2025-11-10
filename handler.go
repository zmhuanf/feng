package feng

import (
	"errors"
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

type requestTmp struct {
	Route   string      `json:"route"`
	ID      string      `json:"id"`
	Type    requestType `json:"type"`
	Data    any         `json:"data"`
	Success bool        `json:"success"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getMain(server *server) func(c *gin.Context) {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		defer conn.Close()

		// 创建上下文
		ctx := newContext()
		u := &user{
			id:     uuid.New().String(),
			server: server,
			conn:   conn,
		}
		r := &room{
			id:    uuid.New().String(),
			users: map[string]IUser{u.id: u},
		}
		u.room = r
		ctx.Set("user", u)
		ctx.Set("room", r)
		ctx.Set("server", server)

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
			var req requestTmp
			err = server.config.Codec.Unmarshal(msg, &req)
			if err != nil {
				continue
			}
			// 回复的消息
			if req.Type == requestTypePushBack {
				continue
			}
			if req.Type == requestTypeRequestBack {
				resp, ok := server.getResponse(req.ID)
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
						Data:    err,
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
			server.middlewaresLock.Lock()
			for _, middleware := range server.middlewares {
				if !strings.HasPrefix(req.Route, middleware.route) {
					continue
				}
				_, err = call(middleware.fn, ctx, req.Data)
				if err != nil {
					server.config.Logger.Error("call middleware func failed", "err", err)
					err = u.send(&requestTmp{
						ID:      req.ID,
						Type:    resType,
						Data:    err,
						Success: false,
					})
					if err != nil {
						server.config.Logger.Error("send middleware error failed", "err", err)
					}
					continue MAINFOR
				}
			}
			server.middlewaresLock.Unlock()
			// 路由处理
			fn, ok := server.getRoute(req.Route)
			if !ok {
				err = u.send(&requestTmp{
					ID:      req.ID,
					Type:    resType,
					Data:    errors.New("route not found"),
					Success: false,
				})
				if err != nil {
					server.config.Logger.Error("send route not found failed", "err", err)
				}
				continue
			}
			data, err := call(fn, ctx, req.Data)
			if err != nil {
				err = u.send(&requestTmp{
					ID:      req.ID,
					Type:    resType,
					Data:    err,
					Success: false,
				})
				if err != nil {
					server.config.Logger.Error("send route error failed", "err", err)
				}
				continue
			}
			err = u.send(&requestTmp{
				ID:      req.ID,
				Type:    resType,
				Data:    data,
				Success: true,
			})
			if err != nil {
				server.config.Logger.Error("send route success failed", "err", err)
			}
		}
	}
}
