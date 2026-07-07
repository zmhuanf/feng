package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zmhuanf/feng/internal/core"
	"github.com/zmhuanf/feng/internal/pending"
	"github.com/zmhuanf/feng/internal/protocol"
	"github.com/zmhuanf/feng/internal/router"
	"github.com/zmhuanf/feng/internal/session"
	"github.com/zmhuanf/feng/internal/transport"
)

func (s *Server) handleWebsocket(isSystem bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		conn, err := transport.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ws := transport.NewConn(conn, s.config.Codec)
		defer ws.Close()

		data := s.channel(isSystem)
		serverCtx := core.NewServerContext(s, ctx)
		room := data.rooms.CreateRoom()
		user := session.NewUser(s, serverCtx, data.rooms, data.pending, ws)
		serverCtx.Bind(room, user)
		_ = room.AddUser(user)
		s.addUser(user, isSystem)
		defer s.removeUser(user.ID(), isSystem)

		for {
			msg, err := ws.Read()
			if err != nil {
				s.config.Logger.Error("read message failed", "err", err)
				return
			}
			if err := s.dispatch(serverCtx, ws, data, msg); err != nil {
				s.config.Logger.Error("dispatch message failed", "err", err)
			}
		}
	}
}

func (s *Server) dispatch(ctx core.ServerContext, sender *transport.Conn, data *channelData, msg *protocol.Message) error {
	switch msg.Type {
	case protocol.MessageTypePushBack:
		if !msg.Success {
			return fmt.Errorf("push back failed, id: %s, data: %s", msg.ID, msg.Data)
		}
		return nil
	case protocol.MessageTypeRequestBack:
		return s.handleRequestBack(ctx, data.pending, msg)
	case protocol.MessageTypePush, protocol.MessageTypeRequest:
		return s.handleIncoming(ctx, sender, data.router, msg)
	default:
		return fmt.Errorf("unknown message type: %d", msg.Type)
	}
}

func (s *Server) handleRequestBack(ctx core.ServerContext, store *pending.Store, msg *protocol.Message) error {
	req, ok := store.Get(msg.ID)
	if !ok {
		return fmt.Errorf("response not found, id: %s", msg.ID)
	}
	if !msg.Success {
		store.Resolve(msg.ID, pending.Result{Success: false, Data: msg.Data})
		return nil
	}
	if _, err := router.Call(req.Callback, ctx, msg.Data, s.config.Codec); err != nil {
		store.Resolve(msg.ID, pending.Result{Success: false, Data: err.Error()})
		return nil
	}
	store.Resolve(msg.ID, pending.Result{Success: true})
	return nil
}

func (s *Server) handleIncoming(ctx core.ServerContext, sender *transport.Conn, route *router.Router, msg *protocol.Message) error {
	responseType := protocol.MessageTypeRequestBack
	if msg.Type == protocol.MessageTypePush {
		responseType = protocol.MessageTypePushBack
	}
	for _, middleware := range route.Middlewares(msg.Route) {
		if _, err := router.Call(middleware.Fn, ctx, msg.Data, s.config.Codec); err != nil {
			return sender.Send(&protocol.Message{ID: msg.ID, Type: responseType, Data: "middleware error: " + err.Error(), Success: false})
		}
	}
	fn, ok := route.Handler(msg.Route)
	if !ok {
		return sender.Send(&protocol.Message{ID: msg.ID, Type: responseType, Data: "route not found", Success: false})
	}
	result, err := router.Call(fn, ctx, msg.Data, s.config.Codec)
	if err != nil {
		return sender.Send(&protocol.Message{ID: msg.ID, Type: responseType, Data: "route error: " + err.Error(), Success: false})
	}
	return sender.Send(&protocol.Message{ID: msg.ID, Type: responseType, Data: result, Success: true})
}
