package main

import (
	"context"
	"errors"
	"fmt"
	log2 "github.com/vaiktorg/grimoire/log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	logger, err := log2.NewLogger()
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for {
			switch rand.Intn(3) {
			case 0:
				logger.TRACE("this is a TraceTest")
			case 1:
				logger.WARN("this is a warning")
			case 2:
				logger.ERROR(errors.New("this is an error"))
			}

			time.Sleep(1 * time.Second)
		}
	}()

	Server(&http.Server{
		Addr:    ":8080",
		Handler: log2.NewLogViewer("TestApp", logger),
	}, func() {
		logger.Close()
	})
}

func Server(server *http.Server, close func()) {
	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, os.ErrClosed) {
			fmt.Println(err)
		}
	}()

	fmt.Println("Listening on", server.Addr)

	// Wait for server to close.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c

	// Closing func
	close()

	err := server.Shutdown(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(0)
}
