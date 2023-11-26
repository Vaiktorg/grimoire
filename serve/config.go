package serve

import (
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/uid"
	"net/http"
	"os"
	"sync"
)

type TLSConfig struct {
	CertPath string
	KeyPath  string
}

type Config struct {
	once sync.Once

	AppName   string
	Addr      string
	TLSConfig *TLSConfig
	Handler   http.Handler
	Logger    log.ILogger
}

func (c *Config) GetTLSConfig() *TLSConfig {
	if c.TLSConfig == nil {
		return &TLSConfig{
			CertPath: "",
			KeyPath:  "",
		}
	}

	return c.TLSConfig
}

func (c *Config) GetLoggerConfig() log.ILogger {
	if c.Logger == nil {
		return log.NewLogger(log.Config{
			CanOutput:   true,
			CanPrint:    true,
			ServiceName: c.GetAppName(),
		})
	}

	return c.Logger
}

func (c *Config) GetAppName() string {
	if c.AppName == "" {
		c.AppName = uid.NewUID(16).String()
	}

	return c.AppName
}

func (c *Config) GetAddr() string {
	if c.TLSConfig != nil {
		return c.Addr + ":443"
	}

	if env, ok := os.LookupEnv("ENVIRONMENT"); ok {
		switch env {
		case "PROD":
			return c.Addr + ":80"
		}
	}

	return c.Addr + ":8080"
}

func (c *Config) GetHandler() http.Handler {
	if c.Handler == nil {
		return http.NewServeMux()
	}

	return c.Handler
}
