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
	appName := uid.NewUID(defaultAppNameLen)

	defaultConfig = &Config{
		Handler: http.NewServeMux(),
		AppName: string(appName),
		Addr:    ":8080",
		Logger: log.NewLogger(&log.Config{
			CanPrint:    true,
			CanOutput:   true,
			Persist:     true,
			ServiceName: string(appName),
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
		l := log.NewStdOutLogger(c.AppName)
		return l
	}

	return c.Logger
}

func (c *Config) GetAppName() string {
	if c.AppName == "" {
		c.AppName = string(uid.NewUID(16))
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
