package main

import (
	"log"
	"net/http"

	"github.com/zackb/updog/db"
	"github.com/zackb/updog/enrichment"
	"github.com/zackb/updog/handler"
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

	// create http server
	server := serve.NewHTTPServer(func(mux *http.ServeMux) {
		mux.Handle("/view", handler.Handler(store, store, enricher))
	})

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
