package main

import (
	"github.com/jgkawell/reddit-api-demo/controller"
	"github.com/jgkawell/reddit-api-demo/handler"

	"github.com/steady-bytes/draft/pkg/chassis"
	"github.com/steady-bytes/draft/pkg/loggers/zerolog"
)

func main() {

	var (
		logger = zerolog.New()
		ctrl   = controller.NewController(logger)
		hnd    = handler.NewHandler(logger, ctrl)
	)

	chassis.New(logger).
		WithRPCHandler(hnd).
		WithRunner(ctrl.Start).
		Start()
}
