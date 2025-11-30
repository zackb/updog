package main

import (
	"log"
	"net/http"
	"time"

	"github.com/zackb/updog/api"
	"github.com/zackb/updog/auth"
	"github.com/zackb/updog/db"
	"github.com/zackb/updog/enrichment"
	"github.com/zackb/updog/frontend"
	"github.com/zackb/updog/handler"
	"github.com/zackb/updog/job"
	"github.com/zackb/updog/serve"
	"github.com/zackb/updog/signal"
)

func main() {
	// initialize database
	store, err := db.NewDB()
	if err != nil {
		log.Fatal("Error initializing storage:", err)
	}

	enricher, err := enrichment.NewEnricher()
	if err != nil {
		log.Fatal("Error initializing enricher:", err)
	}

	// create auth service
	expHours := time.Duration(100) * time.Hour
	auth, err := auth.NewAuthService("jwks.json", expHours)

	if err != nil {
		log.Fatal("Error initializing auth service:", err)
	}

	// initialize api
	api := api.NewAPI(store, auth)

	// initialize frontend
	frontend, err := frontend.NewFrontend(auth, store)

	// create http server
	server := serve.NewHTTPServer(func(mux *http.ServeMux) {
		frontend.Routes(mux)
		mux.Handle("/view", handler.Handler(store, store, enricher, false))
		mux.Handle("/view.gif", handler.Handler(store, store, enricher, true))
		mux.Handle("/api/", api.Routes())
	})

	// create scheduler
	scheduler := job.NewScheduler()
	scheduler.AddDefaultJobs(store)
	scheduler.Start()

	sig := signal.Stop(func() {
		log.Println("Shutting down server...")
		if err := server.Close(); err != nil {
			log.Println("Error shutting down server:", err)
		} else {
			log.Println("Server shut down gracefully.")
		}
		store.Close()
	})

	go func() {
		log.Println("Starting server")
		server.ListenAndServe()
	}()
	<-sig
}
