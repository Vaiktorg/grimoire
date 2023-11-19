package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/log/viewer"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	logger := log.NewLogger(log.Config{
		ServiceName: "Example",
		CanOutput:   true,
	})

	Listen(&http.Server{
		Addr: ":8080",
		Handler: viewer.NewLogViewer(viewer.Config{
			Logger:       logger,
			TemplatePath: "C:/Users/User/go/src/grimoire/log/viewer/templates",
		})}, logger.Close)
}

func Listen(s *http.Server, clean func()) {
	go func() {
		err := s.ListenAndServe()
		if errors.Is(err, os.ErrClosed) {
			fmt.Println(err)
		}
	}()

	fmt.Println("Listening on", s.Addr)

	// Wait for s to clean.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c

	// Closing func
	if clean != nil {
		clean()
	}

	err := s.Shutdown(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(0)
}
