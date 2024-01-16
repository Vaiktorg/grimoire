package main

import (
	"context"
	"github.com/vaiktorg/grimoire/dashboard"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/serve"
)

func main() {
	logger := log.NewSimLogger("Dashboard")

	dashboardHandler, err := dashboard.NewDashboard(&dashboard.Config{
		Logger: logger.NewServiceLogger(&log.Config{
			CanPrint:    true,
			CanOutput:   true,
			Persist:     true,
			ServiceName: "Dashboard",
		}),
		Context:           context.Background(),
		TemplateDirectory: "templates",
		StaticDirectory:   "static",
	})
	if err != nil {
		panic(err)
	}

	server := serve.NewServer(&serve.Config{
		Handler: dashboardHandler,
		AppName: "Dashboard",
		Logger:  logger,
	})

	server.ListenAndServe()
}
