package main

import (
	"github.com/zackb/updog/db"
	"log"
)

func main() {
	// initialize database
	store, err := db.NewDB()
	if err != nil {
		log.Fatal("Error initializing storage:", err)
	}
	store.Close()
}
