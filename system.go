package feng

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type systemJoinReq struct {
	// 地址
	Url string `json:"url"`
	// 签名
	Sign string `json:"sign"`
}

// 加入网络
func systemJoin(ctx IServerContext, req systemJoinReq) error {
	server := ctx.GetServer().(*server)
	// 验证签名
	if !verify(req.Url, server.config.NetworkSignKey, req.Sign) {
		return errors.New("invalid sign")
	}
	// 加入其他服务器状态
	server.otherStatusLock.Lock()
	defer server.otherStatusLock.Unlock()
	server.otherStatus[ctx.GetUser().GetID()] = &serverStatus{
		Url:  req.Url,
		Load: 0,
		ID:   uuid.New().String(),
	}
	return nil
}

type systemReportStatusReq struct {
	// 负载
	Load int `json:"load"`
}

// 上报状态
func systemReportStatus(ctx IServerContext, req systemReportStatusReq) error {
	server := ctx.GetServer().(*server)
	// 验证
	status, ok := server.otherStatus[ctx.GetUser().GetID()]
	if !ok {
		return errors.New("not joined")
	}
	// 更新状态
	status.Load = req.Load
	status.ReportTime = time.Now()
	return nil
}

// 获取低负载服务器地址
func systemGetLowLoadServerAddr(ctx IServerContext, _ string) (string, error) {
	// TODO
	return "", nil
}
