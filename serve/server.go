package serve

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/vaiktorg/grimoire/log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	mu  sync.Mutex
	mux http.Handler

	appName string
	config  *Config
	server  http.Server

	Logger log.ILogger
}

func NewServer(config *Config) *Server {
	return &Server{
		appName: config.GetAppName(),
		mux:     config.GetHandler(),
		Logger:  config.GetLoggerConfig(),
		config:  config,
	}
}

func (s *Server) ListenAndServe() {
	defer s.Logger.Close()

	s.server = http.Server{
		Addr:    s.config.GetAddr(),
		Handler: s.mux,
	}

	// Start the server in a goroutine
	go func() {
		s.Logger.TRACE("Listening " + s.appName + " on " + s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Logger.ERROR(errors.New("ListenAndServe error: " + err.Error()))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// Block until a signal is received
	<-quit
	s.Logger.TRACE("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline.
	if err := s.server.Shutdown(ctx); err != nil {
		s.Logger.FATAL("ListenAndServe shutdown error: " + err.Error())
	}
}
func (s *Server) ListenAndServeTLS() {
	if s.config.TLSConfig == nil ||
		s.config.TLSConfig.CertPath == "" ||
		s.config.TLSConfig.KeyPath == "" {
		panic("validate tls server config file paths")
	}

	// Load your certificate and private key
	cert, err := tls.LoadX509KeyPair(s.config.TLSConfig.CertPath, s.config.TLSConfig.KeyPath)
	if err != nil {
		s.Logger.FATAL(err.Error())
	}

	// Define the TLS config in a Server
	s.server = http.Server{
		Addr:    s.config.GetAddr(),
		Handler: s.mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}
	// Close logger
	defer s.Logger.Close()

	// Start the server in a goroutine
	go func() {
		if err = s.server.ListenAndServeTLS("", ""); err != nil || !errors.Is(err, http.ErrServerClosed) {
			s.Logger.FATAL("ListenAndServe error: " + err.Error())
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// Block until a signal is received
	<-quit
	s.Logger.TRACE("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline.
	if err = s.server.Shutdown(ctx); err != nil {
		s.Logger.FATAL("ListenAndServe Shutdown: " + err.Error())
	}

	// Bye bye!
	s.Logger.TRACE(s.config.AppName + " exiting...")
}

func (s *Server) RegisterPaths(muxReg func(*http.ServeMux)) {
	if mux, ok := s.mux.(*http.ServeMux); ok {
		muxReg(mux)
	} else {
		s.Logger.TRACE("cannot register paths, server handler is not mux")
	}
}
