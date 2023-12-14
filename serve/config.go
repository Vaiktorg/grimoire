package serve

import (
	_ "embed"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/uid"
	"net/http"
	"os"
)

var defaultConfig *Config

const defaultAppNameLen = 16

func init() {
	appName := uid.NewUID(defaultAppNameLen).String()

	defaultConfig = &Config{
		Handler: http.NewServeMux(),
		AppName: appName,
		Addr:    ":8080",
		Logger: log.NewLogger(log.Config{
			CanPrint:    true,
			CanOutput:   true,
			Persist:     false,
			ServiceName: appName,
		}),
	}
}

type TLSConfig struct {
	CertPath string
	KeyPath  string
}

type Config struct {
	Handler http.Handler

	AppName string
	Addr    string

	Logger    log.ILogger
	TLSConfig *TLSConfig
}

func (c *Config) GetTLSConfig() *TLSConfig {
	if c.TLSConfig == nil {
		return &TLSConfig{
			CertPath: "cert/cert.pem",
			KeyPath:  "cert/private.key",
		}
	}

	return c.TLSConfig
}

func (c *Config) GetLoggerConfig() log.ILogger {
	if c.Logger == nil {
		return log.NewLogger(log.Config{
			CanOutput:   true,
			CanPrint:    true,
			Persist:     false,
			ServiceName: c.AppName,
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

	if env, ok := os.LookupEnv("ENV"); ok {
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
