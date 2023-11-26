package main

import (
	"github.com/vaiktorg/grimoire/dashboard"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/serve"
)

func main() {
	logger := log.NewLogger(log.Config{
		CanPrint:    true,
		CanOutput:   true,
		ServiceName: "VKTRG Diagnostics",
	})

	server := serve.NewServer(&serve.Config{
		AppName: "VKTRG Diagnostics",
		Logger:  logger,
		Handler: dashboard.NewDashboard(dashboard.Config{
			Logger: logger,
		}),
	})

	server.ListenAndServe()
}
