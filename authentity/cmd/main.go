package main

import (
	"github.com/vaiktorg/grimoire/authentity/internal"
	"github.com/vaiktorg/grimoire/authentity/src"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/serve"
	"gorm.io/driver/sqlite"
)

func main() {
	dial := sqlite.Open("testdb.sqlite")
	AppName := "AuthentityServer"

	logger := log.NewLogger(log.Config{
		CanPrint:    true,
		CanOutput:   true,
		Persist:     false,
		ServiceName: AppName,
	})

	authentityHandler := src.NewAuthentity(AppName, logger.NewServiceLogger(log.Config{
		CanPrint:    true,
		CanOutput:   true,
		Persist:     false,
		ServiceName: "Authentity",
	}), dial)

	server := serve.NewServer(&serve.Config{
		Handler: internal.SecurityMiddleware(authentityHandler),
		AppName: AppName,
		Logger:  logger,
	})

	server.ListenAndServe()
}
