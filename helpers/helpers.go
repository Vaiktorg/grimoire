package helpers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	ANSIClearScreen = "|.[H.[2J.[3J|"
	ANSIClearLine   = "\033[2K\r"
)

func MakeTimestampStr() string { return time.Now().Format("Jan 02 2006 - 15:04:05 PM") }
func MakeTimestampNum() string { return time.Now().Format("20060102150405") }

func ConsoleCloser(appName string, execFunc func(), disposeFunc func()) {
	go execFunc()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	if disposeFunc != nil {
		disposeFunc()
	}

	fmt.Println(ANSIClearLine + "----------------------------------------\n\t" + appName + " Exited...")
	os.Exit(0)
}

func ServerCloser(server *http.Server) {
	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, os.ErrClosed) {
			fmt.Println(err)
		}
	}()
	fmt.Println("Listening on", server.Addr)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	err := server.Shutdown(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Shutting Down")
	os.Exit(0)
}
