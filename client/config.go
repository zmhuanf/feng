package client

import (
	"time"

	"github.com/zmhuanf/feng"
)

type Config struct {
	// 服务器地址
	Addr string
	// 服务器端口
	Port int
	// 序列化方式
	Codec feng.ICodec
	// 日志记录器
	Logger feng.Logger
	// 全局超时时间
	Timeout time.Duration
	// 启用TLS
	EnableTLS bool
}

func NewDefaultClientConfig() *Config {
	return &Config{
		Addr:    "127.0.0.1",
		Port:    22100,
		Codec:   feng.NewJsonCodec(),
		Logger:  feng.NewSlogLogger(),
		Timeout: 5 * time.Minute,
	}
}
