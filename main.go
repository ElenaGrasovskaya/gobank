package main

import (
	"fmt"
	"log"
)

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
		fmt.Println("store")
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}
	server := NewAPIServer(":3000", store)
	server.Run()

	/* func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}

	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}

	r := gin.Default()


	if err := r.Run(":3000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}

	r.Run(":3000") */
}
