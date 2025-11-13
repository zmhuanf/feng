package feng

import "time"

type serverConfig struct {
	// 监听地址
	Addr string
	// 监听端口
	Port int
	// 序列化方式
	Codec ICodec
	// 日志记录器
	Logger Logger
	// 证书文件路径
	CertFile string
	// 私钥文件路径
	KeyFile string
	// 全局超时时间
	Timeout time.Duration
	// 加入的网络
	JoinNetwork string
	// 网络签名密钥
	NetworkSignKey string
	// 心跳上报间隔
	ReportInterval time.Duration
	// 超时移除间隔
	RemoveInterval time.Duration
}

func NewDefaultServerConfig() *serverConfig {
	return &serverConfig{
		Addr:           "0.0.0.0",
		Port:           22100,
		Codec:          NewJsonCodec(),
		Logger:         NewSlogLogger(),
		CertFile:       "",
		KeyFile:        "",
		Timeout:        5 * time.Minute,
		JoinNetwork:    "",
		NetworkSignKey: generateRandomKey(64),
		ReportInterval: time.Minute,
		RemoveInterval: 10 * time.Second,
	}
}
