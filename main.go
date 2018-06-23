package main

import (
	"context"
	"net/http"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/healthcheck/api"
	"github.com/gophish/healthcheck/config"
	"github.com/gophish/healthcheck/db"
)

func main() {
	err := config.LoadConfig("./config.json")
	if err != nil {
		panic(err)
	}
	err = db.Setup()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mailer.Mailer.Start(ctx)

	router := api.NewAPIRouter()
	log.Info("API Server started on :3000")
	http.ListenAndServe(":3000", router)
}
