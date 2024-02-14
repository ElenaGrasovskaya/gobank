package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"

	"github.com/gin-gonic/gin"
)

/* func TestHandleLogin(t *testing.T) {
	// Setup
	store := NewMockStore()
	server := NewAPIServer(":3000", store)
	router := gin.Default()
	router.POST("/login", server.handleLogin)

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
} */

func TestHandleLogin(t *testing.T) {

	store := NewMockStore()

	email := "elenagrasovskaya@gmail.com"
	mockAccount := &Account{
		ID:        1,
		FirstName: "Elena",
		LastName:  "Grasovskaya",
		Email:     email,
		Password:  "$2a$10$tp.LGA7nUpiAf4bgW9quDeg02QSo7JiXBup7NXCq7FTAojDxW4.iC", // Use an appropriate hashed password for comparison
	}
	store.On("GetAccountByEmail", email).Return(mockAccount, nil)

	server := NewAPIServer(":3000", store)
	router := gin.Default()
	router.POST("/login", server.handleLogin)

	userData := &LoginRequest{
		Email:    email,
		Password: "11111",
	}

	body, err := json.Marshal(userData)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK; got %v", w.Code)
	}
	// Add more assertions as needed
}

func TestHandleRegister(t *testing.T) {
	// Setup
	store := NewMockStore()

	// Setting up the expectation
	email := "testing@gmail.com"
	mockAccount := &Account{
		ID:        7,
		FirstName: "Test",
		LastName:  "Testovich",
		Email:     email,
		Password:  "$2a$10$q/cjukk2QtKtTdcaype0UOgPydr5MRcQm9wmbpfvyDksUuuv2gomu", // Use an appropriate hashed password for comparison
	}
	store.On("GetAccountByEmail", email).Return(mockAccount, nil)

	server := NewAPIServer(":3000", store)
	router := gin.Default()
	router.POST("/register", server.handleCreateAccount)

	userData := &CreateAccountRequest{

		FirstName: "Test",
		LastName:  "Testovich",
		Email:     email,
		Password:  "test"}

	body, err := json.Marshal(userData)
	if err != nil {
		t.Fatalf("Failed to marshal register data: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	// Assert
	if w.Code == http.StatusAccepted {
		t.Errorf("Expected Account testing@gmail.com already exists; got %v", w.Code)
	}
	// Add more assertions as needed
}
