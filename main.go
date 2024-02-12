package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

/* func main() {
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
} */

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}

	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}
	s := NewAPIServer(":3000", store)

	r := gin.Default()
	r.Use(CorsMiddleware())

	authMiddleware := WithJWTAuthMiddleware(store)

	r.POST("/login", s.handleLogin)

	r.POST("/logout", s.handleLogout)
	r.POST("/register", s.handleRegister)

	authGroup := r.Group("/")
	authGroup.Use(authMiddleware)
	{
		authGroup.POST("/expense", s.handleCreateExpense)
		authGroup.POST("/expense/:id", s.handleUpdateExpense)
		authGroup.GET("/expense", s.handleGetExpenseForUser)
		authGroup.DELETE("/expense/:id", s.handleDeleteExpense)
		authGroup.GET("/expenses", s.handleGetAllExpense)

		authGroup.GET("/account", s.handleGetAccount)
		authGroup.POST("/account", s.handleCreateAccount)
		authGroup.DELETE("/account", s.handleDeleteAccount)
		authGroup.GET("/account/:id", s.handleGetAccountById)
	}
	fmt.Println("JSON API server is running on port: 3000")

	if err := r.Run(":3000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}

	r.Run(":3000")
}
