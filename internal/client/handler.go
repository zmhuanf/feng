package client

import (
	"context"
	"fmt"

	"github.com/zmhuanf/feng/internal/core"
	"github.com/zmhuanf/feng/internal/pending"
	"github.com/zmhuanf/feng/internal/protocol"
	"github.com/zmhuanf/feng/internal/router"
)

func (c *Client) readLoop(ctx context.Context, ch *channel) {
	clientCtx := core.NewClientContext(c)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := ch.conn.Read()
		if err != nil {
			c.config.Logger.Error("read message failed", "err", err)
			return
		}
		if err := c.dispatch(clientCtx, ch, msg); err != nil {
			c.config.Logger.Error("dispatch message failed", "err", err)
		}
	}
}

func (c *Client) dispatch(ctx core.ClientContext, ch *channel, msg *protocol.Message) error {
	switch msg.Type {
	case protocol.MessageTypePushBack:
		return nil
	case protocol.MessageTypeRequestBack:
		return c.handleRequestBack(ctx, ch.pending, msg)
	case protocol.MessageTypePush, protocol.MessageTypeRequest:
		return c.handleIncoming(ctx, ch, msg)
	default:
		return fmt.Errorf("unknown message type: %d", msg.Type)
	}
}

func (c *Client) handleRequestBack(ctx core.ClientContext, store *pending.Store, msg *protocol.Message) error {
	req, ok := store.Get(msg.ID)
	if !ok {
		return nil
	}
	if !msg.Success {
		store.Resolve(msg.ID, pending.Result{Success: false, Data: msg.Data})
		return nil
	}
	if _, err := router.Call(req.Callback, ctx, msg.Data, c.config.Codec); err != nil {
		store.Resolve(msg.ID, pending.Result{Success: false, Data: err.Error()})
		return nil
	}
	store.Resolve(msg.ID, pending.Result{Success: true})
	return nil
}

func (c *Client) handleIncoming(ctx core.ClientContext, ch *channel, msg *protocol.Message) error {
	responseType := protocol.MessageTypeRequestBack
	if msg.Type == protocol.MessageTypePush {
		responseType = protocol.MessageTypePushBack
	}
	for _, middleware := range ch.router.Middlewares(msg.Route) {
		if _, err := router.Call(middleware.Fn, ctx, msg.Data, c.config.Codec); err != nil {
			return ch.conn.Send(&protocol.Message{ID: msg.ID, Type: responseType, Data: err.Error(), Success: false})
		}
	}
	fn, ok := ch.router.Handler(msg.Route)
	if !ok {
		return ch.conn.Send(&protocol.Message{ID: msg.ID, Type: responseType, Data: "route not found", Success: false})
	}
	result, err := router.Call(fn, ctx, msg.Data, c.config.Codec)
	if err != nil {
		return ch.conn.Send(&protocol.Message{ID: msg.ID, Type: responseType, Data: err.Error(), Success: false})
	}
	return ch.conn.Send(&protocol.Message{ID: msg.ID, Type: responseType, Data: result, Success: true})
}
