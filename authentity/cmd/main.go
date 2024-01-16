package main

import (
	"github.com/vaiktorg/grimoire/authentity/src"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/serve"
	"github.com/vaiktorg/grimoire/uid"
)

func main() {
	logger := log.NewLogger(&log.Config{
		CanPrint:    true,
		CanOutput:   true,
		Persist:     true,
		ServiceName: "Server",
	})

	server := serve.NewServer(&serve.Config{
		AppName: "Server",
		Logger:  logger,
		Handler: src.NewAuthentity(&src.Config{
			Issuer: "Authentity",
			GSpice: gwt.Spice{
				Salt:   []byte(uid.New()),
				Pepper: []byte(uid.New()),
			},
			Logger: logger.NewServiceLogger(&log.Config{
				CanPrint:    true,
				CanOutput:   true,
				Persist:     true,
				ServiceName: "Authentity",
			}),
		}),
	})

	server.ListenAndServe()
}
