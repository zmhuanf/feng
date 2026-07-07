package server

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zmhuanf/feng/internal/core"
)

type systemJoinReq struct {
	URL  string `json:"url"`
	Sign string `json:"sign"`
}

type systemReportStatusReq struct {
	Load int `json:"load"`
}

func (s *Server) addSystemHandlers() {
	_ = s.systemData.router.Handle("/join", s.systemJoin)
	_ = s.systemData.router.Handle("/report_status", s.systemReportStatus)
	_ = s.systemData.router.Handle("/get_low_load_server_addr", s.systemGetLowLoadServerAddr)
}

func (s *Server) systemJoin(ctx core.ServerContext, req systemJoinReq) error {
	if !core.Verify(req.URL, s.config.NetworkSignKey, req.Sign) {
		return errors.New("invalid sign")
	}
	s.peersLock.Lock()
	defer s.peersLock.Unlock()
	s.peers[ctx.User().ID()] = &Status{URL: req.URL, Load: 0, ID: uuid.New().String(), ReportTime: time.Now()}
	return nil
}

func (s *Server) systemReportStatus(ctx core.ServerContext, req systemReportStatusReq) error {
	s.peersLock.Lock()
	defer s.peersLock.Unlock()
	status, ok := s.peers[ctx.User().ID()]
	if !ok {
		return errors.New("not joined")
	}
	status.Load = req.Load
	status.ReportTime = time.Now()
	return nil
}

func (s *Server) systemGetLowLoadServerAddr(core.ServerContext, bool) (string, error) {
	return "", nil
}
