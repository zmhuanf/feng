package feng

import (
	"strings"

	"github.com/gorilla/websocket"
)

func clientHandle(c *client) {
MAINFOR:
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		if msgType != websocket.TextMessage {
			continue
		}
		// 解析请求
		var req message
		err = c.config.Codec.Unmarshal(msg, &req)
		if err != nil {
			continue
		}
		// 回复的消息
		if req.Type == messageTypePushBack {
			continue
		}
		ctx := newClientContext(c)
		if req.Type == messageTypeRequestBack {
			resp, ok := c.getResponse(req.ID)
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
		resType := messageTypeRequestBack
		if req.Type == messageTypePush {
			resType = messageTypePushBack
		}
		// 中间件处理
		c.middlewaresLock.Lock()
		for _, middleware := range c.middlewares {
			if !strings.HasPrefix(req.Route, middleware.route) {
				continue
			}
			_, err = call(middleware.fn, ctx, req.Data)
			if err != nil {
				c.config.Logger.Error("call middleware func failed", "err", err)
				err = c.send(&message{
					ID:      req.ID,
					Type:    resType,
					Data:    err.Error(),
					Success: false,
				})
				if err != nil {
					c.config.Logger.Error("send middleware error failed", "err", err)
				}
				continue MAINFOR
			}
		}
		c.middlewaresLock.Unlock()
		// 路由处理
		fn, ok := c.getRoute(req.Route)
		if !ok {
			err = c.send(&message{
				ID:      req.ID,
				Type:    resType,
				Data:    "route not found",
				Success: false,
			})
			if err != nil {
				c.config.Logger.Error("send route not found failed", "err", err)
			}
			continue
		}
		data, err := call(fn, ctx, req.Data)
		if err != nil {
			err = c.send(&message{
				ID:      req.ID,
				Type:    resType,
				Data:    err.Error(),
				Success: false,
			})
			if err != nil {
				c.config.Logger.Error("send route error failed", "err", err)
			}
			continue
		}
		err = c.send(&message{
			ID:      req.ID,
			Type:    resType,
			Data:    data,
			Success: true,
		})
		if err != nil {
			c.config.Logger.Error("send route success failed", "err", err)
		}
	}
}
