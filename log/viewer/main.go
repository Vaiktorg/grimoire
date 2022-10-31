package main

import (
	"context"
	"fmt"
	"github.com/vaiktorg/grimoire/log/log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	logger, err := log.NewLogger("TestLogger1")
	if err != nil {
		fmt.Println(err)
		return
	}

	logger.TRACE("Start Server")

	Server(&http.Server{
		Addr:    ":8080",
		Handler: logger.Handler,
	}, func() {
		logger.Close()
	})
}

func Server(server *http.Server, close func()) {
	go func() {
		err := server.ListenAndServe()
		if err == os.ErrClosed {
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

	fmt.Println("Shutting Down")
	os.Exit(0)
}
