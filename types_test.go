package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	acc, err := NewAccount("a", "b", "c@gmail", "111")
	assert.Nil(t, err)

	fmt.Printf("%v /n", acc)

}

func TestNewRequest(t *testing.T) {
	acc, err := NewExpense(1, "testa", "testb", "test", 100, time.Now())
	assert.Nil(t, err)

	fmt.Printf("%v /n", acc)
}

func TestHandleLogin(t *testing.T) {
	// Setup
	store := NewMockStore() // Assuming you have a function to mock your store
	server := NewAPIServer(":3000", store)
	router := gin.Default()
	router.POST("/login", server.handleLogin)

	// Execute

	userData := &LoginRequest{
		Email:    "elenagrasovskaya@gmail.com",
		Password: "11111",
	}

	body, err := json.Marshal(userData)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK; got %v", w.Code)
	}
	// Add more assertions as needed
}
