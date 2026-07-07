package core

import "time"

type ServerConfig struct {
	// 监听地址。
	Addr string
	// 监听端口。
	Port int
	// 序列化方式。
	Codec Codec
	// 日志记录器。
	Logger Logger
	// 证书文件路径。
	CertFile string
	// 私钥文件路径。
	KeyFile string
	// 全局请求超时时间。
	Timeout time.Duration
	// 要加入的服务器网络地址。
	JoinNetwork string
	// 服务器网络签名密钥。
	NetworkSignKey string
	// 心跳上报间隔。
	ReportInterval time.Duration
	// 超时节点清理间隔。
	RemoveInterval time.Duration
	// 每页房间/用户数量。
	PageSize int
}

func NewDefaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:           "0.0.0.0",
		Port:           22100,
		Codec:          NewJSONCodec(),
		Logger:         NewSlogLogger(),
		Timeout:        5 * time.Minute,
		NetworkSignKey: GenerateRandomKey(64),
		ReportInterval: time.Minute,
		RemoveInterval: 10 * time.Second,
		PageSize:       10,
	}
}

type Mode int

const (
	ModeClient Mode = iota
	ModeServer
)

type ClientConfig struct {
	// 服务器地址。
	Addr string
	// 服务器端口。
	Port int
	// 序列化方式。
	Codec Codec
	// 日志记录器。
	Logger Logger
	// 全局请求超时时间。
	Timeout time.Duration
	// 是否启用 TLS。
	EnableTLS bool
	// 是否直接连接游戏链路。
	DirectConnect bool
	// 内部连接模式。
	Mode Mode
}

func NewDefaultClientConfig() ClientConfig {
	return ClientConfig{
		Addr:          "127.0.0.1",
		Port:          22100,
		Codec:         NewJSONCodec(),
		Logger:        NewSlogLogger(),
		Timeout:       5 * time.Minute,
		DirectConnect: true,
		Mode:          ModeClient,
	}
}

func NormalizeServerConfig(config ServerConfig) ServerConfig {
	defaults := NewDefaultServerConfig()
	if config.Addr == "" {
		config.Addr = defaults.Addr
	}
	if config.Port == 0 {
		config.Port = defaults.Port
	}
	if config.Codec == nil {
		config.Codec = defaults.Codec
	}
	if config.Logger == nil {
		config.Logger = defaults.Logger
	}
	if config.Timeout == 0 {
		config.Timeout = defaults.Timeout
	}
	if config.NetworkSignKey == "" {
		config.NetworkSignKey = defaults.NetworkSignKey
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = defaults.ReportInterval
	}
	if config.RemoveInterval == 0 {
		config.RemoveInterval = defaults.RemoveInterval
	}
	if config.PageSize <= 0 {
		config.PageSize = defaults.PageSize
	}
	return config
}

func NormalizeClientConfig(config ClientConfig) ClientConfig {
	defaults := NewDefaultClientConfig()
	if config.Addr == "" {
		config.Addr = defaults.Addr
	}
	if config.Port == 0 {
		config.Port = defaults.Port
	}
	if config.Codec == nil {
		config.Codec = defaults.Codec
	}
	if config.Logger == nil {
		config.Logger = defaults.Logger
	}
	if config.Timeout == 0 {
		config.Timeout = defaults.Timeout
	}
	return config
}
