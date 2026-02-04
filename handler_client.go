package feng

import (
	"strings"

	"github.com/gorilla/websocket"
)

func clientHandle(c *client, isSys bool) {
	ctx := c.ctx
	conn := c.conn
	if isSys {
		ctx = c.ctxSys
		conn = c.connSys
	}
	ctxClient := newClientContext(c)
MAINFOR:
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msgType, msg, err := conn.ReadMessage()
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
		// 推送的回复，忽略
		if req.Type == messageTypePushBack {
			continue
		}
		// 请求的回复
		if req.Type == messageTypeRequestBack {
			resp, ok := c.getResponse(req.ID, isSys)
			if !ok {
				continue
			}
			c.deleteResponse(req.ID, isSys)
			if !req.Success {
				resp.ch <- chanData{
					Success: false,
					Data:    req.Data,
				}
				continue
			}
			_, err = call(resp.fn, ctxClient, req.Data)
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
			close(resp.ch)
			continue
		}
		// 请求的消息
		resType := messageTypeRequestBack
		if req.Type == messageTypePush {
			resType = messageTypePushBack
		}
		// 中间件处理
		lock := &c.middlewaresLock
		if isSys {
			lock = &c.middlewaresSysLock
		}
		lock.Lock()
		middlewares := c.middlewares
		if isSys {
			middlewares = c.middlewaresSys
		}
		for _, middleware := range middlewares {
			if !strings.HasPrefix(req.Route, middleware.route) {
				continue
			}
			_, err = call(middleware.fn, ctxClient, req.Data)
			if err != nil {
				c.config.Logger.Error("call middleware func failed", "err", err)
				err = c.send(&message{
					ID:      req.ID,
					Type:    resType,
					Data:    err.Error(),
					Success: false,
				}, isSys)
				if err != nil {
					c.config.Logger.Error("send middleware error failed", "err", err)
				}
				lock.Unlock()
				continue MAINFOR
			}
		}
		lock.Unlock()
		// 路由处理
		fn, ok := c.getRoute(req.Route, isSys)
		if !ok {
			err = c.send(&message{
				ID:      req.ID,
				Type:    resType,
				Data:    "route not found",
				Success: false,
			}, isSys)
			if err != nil {
				c.config.Logger.Error("send route not found failed", "err", err)
			}
			continue
		}
		data, err := call(fn, ctxClient, req.Data)
		if err != nil {
			err = c.send(&message{
				ID:      req.ID,
				Type:    resType,
				Data:    err.Error(),
				Success: false,
			}, isSys)
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
		}, isSys)
		if err != nil {
			c.config.Logger.Error("send route success failed", "err", err)
		}
	}
}
