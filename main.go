package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ElenaGrasovskaya/gobank/account"
	"github.com/ElenaGrasovskaya/gobank/expense"
	"github.com/ElenaGrasovskaya/gobank/services"
	"github.com/ElenaGrasovskaya/gobank/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	store, err := storage.NewPostgresStore()

	if err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}

	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}

	e := expense.NewExpenseHandler(store)
	a := account.NewAccountHandler(store)
	s := services.NewServiceHandler(store)
	r := gin.Default()
	r.Use(services.CorsMiddleware())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Not Found"})
	})

	authMiddleware := services.WithJWTAuthMiddleware(store)

	r.GET("/", s.HandleHealth)
	r.POST("/login", s.HandleLogin)
	r.POST("/register", s.HandleRegister)
	r.POST("/logout", s.HandleLogout)

	authGroup := r.Group("/")
	authGroup.Use(authMiddleware)
	{
		authGroup.POST("/expense", e.HandleCreateExpense)
		authGroup.POST("/expense/:id", e.HandleUpdateExpense)
		authGroup.GET("/expense", e.HandleGetExpenseForUser)
		authGroup.DELETE("/expense/:id", e.HandleDeleteExpense)
		authGroup.GET("/expenses", e.HandleGetAllExpense)

		authGroup.GET("/accounts", a.HandleGetAccount)
		authGroup.POST("/account", a.HandleCreateAccount)
		authGroup.DELETE("/account/:id", a.HandleDeleteAccount)
		authGroup.GET("/account/:id", a.HandleGetAccountById)
	}
	fmt.Println("JSON API server is running on port: 3000")

	if err := r.Run(":3000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}

	r.Run(":3000")
}
