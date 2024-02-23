package main

import (
	"fmt"
	"log"

	"github.com/ElenaGrasovskaya/gobank/router"
	"github.com/ElenaGrasovskaya/gobank/storage"
)

func main() {
	store, err := storage.NewPostgresStore()
	if err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}

	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}

	r := router.SetupRouter(store)
	fmt.Println("JSON API server is running on port: 3000")

	if err := r.Run(":3000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
