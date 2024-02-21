package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bytes"

	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	return router
}
func setupTestServer() (*APIServer, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}
	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize the store: %v", err)
	}
	server := NewAPIServer(":3000", store)
	return server, router
}

func TestHandleLogin(t *testing.T) {
	server, router := setupTestServer()

	email := "elenagrasovskaya@gmail.com"
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
}

func TestHandleRegister(t *testing.T) {
	server, router := setupTestServer()
	router.POST("/register", server.handleCreateAccount)

	email := "elenagrasovskaya@gmail.com"

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

	if w.Code == http.StatusAccepted {
		t.Errorf("Expected Account testing@gmail.com already exists; got %v", w.Code)
	}

	if w.Code == http.StatusConflict {
		t.Errorf("Expected Account testing@gmail.com already exists; got %v", w.Code)
	}
}

func TestHandleLogout(t *testing.T) {
	server, router := setupTestServer()
	//Case 1 logout without a cookie
	router.POST("/logout", server.handleLogout)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/logout", nil)

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected 200; got %v", w.Code)

	//Case 2 wrong URL
	var testString string = "test!"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", fmt.Sprintf("/logout/%v", testString), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected not found got %v", w.Code)
}

func TestHandleGetAccount(t *testing.T) {
	server, router := setupTestServer()
	router.GET("/accounts", server.handleGetAccount)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/accounts", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected Status.OK and accounts data; got %v", w.Code)
	}
}

func TestHandleGetAccountById(t *testing.T) {
	server, router := setupTestServer()

	router.GET("/account/:id", server.handleGetAccountById)

	testID := 7
	mockAccount := &Account{
		ID:        7,
		FirstName: "Test",
		LastName:  "Testovich",
		Email:     "testing@gmail.com",
		Password:  "",
		Status:    "Active",
		Number:    112302,
		Balance:   0,
		CreatedAt: time.Now(),
	}

	// Test 1: Valid request

	req, _ := http.NewRequest("GET", fmt.Sprintf("/account/%d", testID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

	var response ResponceAccount
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(t, err, "Expected no error unmarshalling response")
	assert.Equal(t, mockAccount.ID, response.ID, "Expected account ID to match mock account")
	assert.Equal(t, mockAccount.Email, response.Email, "Expected account email to match mock account")

	// Test 2: Invalid request
	testInvalidId := "test"
	req, _ = http.NewRequest("GET", fmt.Sprintf("/account/%v", testInvalidId), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")

	//Test 3: Invalid request starting with numbers

	testInvalidId = "75test"
	req, _ = http.NewRequest("GET", fmt.Sprintf("/account/%v", testInvalidId), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")

	//Test 4: Invalid request starting with numbers

	testInvalidId = ""
	req, _ = http.NewRequest("GET", fmt.Sprintf("/account/%v", testInvalidId), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")

	//Test 5: Account doesn't exist

	testID = 0
	req, _ = http.NewRequest("GET", fmt.Sprintf("/account/%d", testID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected Bad request, got %v", w.Code)

}

func TestHandleDeleteAccount(t *testing.T) {
	server, router := setupTestServer()

	router.DELETE("/account/:id", server.handleDeleteAccount)

	server.store.RestoreAccount(7)

	tests := []struct {
		description  string
		accountID    string
		expectedCode int
		expectedBody string
	}{
		{"Delete existing account", "7", http.StatusOK, "{\"deleted\":7}"},
		{"Delete non-existing account", "0", http.StatusBadRequest, ""},
		{"Invalid account ID", "abc", http.StatusBadRequest, ""},
		{"Delete already deleted account", "7", http.StatusBadRequest, "{\"This accout was already deleted\": 7}"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/account/"+test.accountID, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedCode, w.Code)

			if test.expectedBody != "" {
				assert.JSONEq(t, test.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHandleGetAllExpense(t *testing.T) {
	server, router := setupTestServer()
	router.GET("/expenses", server.handleGetAllExpense)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/expenses", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected Status.OK and expenses data; got %v", w.Code)
	}
}

func TestHandleGetExpenseForUser(t *testing.T) {
	server, router := setupTestServer()

	router.GET("/expense", server.handleGetExpenseForUser)
	// Test 1: Not authorized request

	req, _ := http.NewRequest("GET", "/expense", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected status code 401")

	// Test 2: Valid request
	cookie := &http.Cookie{
		Name:  "token",
		Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RpbmdAZ21haWwuY29tIiwiaWQiOjd9.L96v-PecYaVZ7vjeZ3uSbGcQXhfOGHCOFeJj0rCWH0w",
	}
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 and expense data for the account")

	var response []Expense
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Expected no error unmarshalling response")

	// Test 3: Invalid request
	testInvalidId := "test"
	req, _ = http.NewRequest("GET", fmt.Sprintf("/expense/%v", testInvalidId), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")
}

func TestHandleCreateExpense(t *testing.T) {
	server, router := setupTestServer()

	router.POST("/expense", server.handleCreateExpense)

	expenseData := &CreateExpenseRequest{
		ExpenseName:     "test",
		ExpensePurpose:  "test",
		ExpenseCategory: "test",
		ExpenseValue:    100,
		CreatedAt:       time.Now(),
	}

	body, err := json.Marshal(expenseData)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}
	// Test 1: Not authorized request

	req, _ := http.NewRequest("POST", "/expense", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected status code 401")

	// Test 2: Valid request
	cookie := &http.Cookie{
		Name:  "token",
		Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RpbmdAZ21haWwuY29tIiwiaWQiOjd9.L96v-PecYaVZ7vjeZ3uSbGcQXhfOGHCOFeJj0rCWH0w",
	}
	req, _ = http.NewRequest("POST", "/expense", bytes.NewBuffer(body))
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 and new expense data")

	// Test 3: Invalid request
	testInvalidId := "test"
	req, _ = http.NewRequest("GET", fmt.Sprintf("/expense%v", testInvalidId), bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")
}

/* func TestHandleUpdateExpense(t *testing.T) {
	server, router := setupTestServer()

	router.POST("/expense", server.handleUpdateExpense)

	newExpenseData := &Expense{
		UserId:          7,
		ExpenseName:     "test",
		ExpensePurpose:  "test",
		ExpenseCategory: "test",
		ExpenseValue:    100,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	newExp, err := server.store.CreateExpense(newExpenseData)

	editId := newExp.ID

	updateExpenseData := &UpdateExpenseRequest{
		ExpenseName:     "edited",
		ExpensePurpose:  "edited",
		ExpenseCategory: "edited",
		ExpenseValue:    200,
		CreatedAt:       time.Now(),
	}

	body, err := json.Marshal(updateExpenseData)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}
	// Test 1: Not authorized request

	req, _ := http.NewRequest("POST", fmt.Sprintf("/expense/%d", editId), bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected status code 401")

	// Test 2: Valid request
	cookie := &http.Cookie{
		Name:  "token",
		Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RpbmdAZ21haWwuY29tIiwiaWQiOjd9.L96v-PecYaVZ7vjeZ3uSbGcQXhfOGHCOFeJj0rCWH0w",
	}
	req, _ = http.NewRequest("POST", fmt.Sprintf("/expense/%d", editId), bytes.NewBuffer(body))
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 and new expense data")

	// Test 3: Invalid request
	testInvalidId := "test"
	req, _ = http.NewRequest("POST", fmt.Sprintf("/expense/%v", testInvalidId), bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")
} */

func TestHandleDeleteExpense(t *testing.T) {
	server, router := setupTestServer()

	router.DELETE("/expense/:id", server.handleDeleteExpense)

	testExpense, err := NewExpense(7, "test", "test", "test", 100, time.Now())
	assert.NoError(t, err, "Expected no error creating new expense")

	newExpense, newErr := server.store.CreateExpense(testExpense)
	assert.NoError(t, newErr, "Expected no error creating new expense")

	tests := []struct {
		description  string
		expenseID    string
		expectedCode int
		expectedBody string
	}{
		{"Delete existing expense", fmt.Sprintf("%d", newExpense.ID), http.StatusOK, fmt.Sprintf("{\"deleted\":%d}", newExpense.ID)},
		{"Delete non-existing expense", fmt.Sprintf("%d", newExpense.ID), http.StatusNotFound, ""},
		{"Invalid account ID", "abc", http.StatusNotFound, ""},
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RpbmdAZ21haWwuY29tIiwiaWQiOjd9.L96v-PecYaVZ7vjeZ3uSbGcQXhfOGHCOFeJj0rCWH0w",
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/expense/"+test.expenseID, nil)
			req.AddCookie(cookie)
			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedCode, w.Code)

			if test.expectedBody != "" {
				assert.JSONEq(t, test.expectedBody, w.Body.String())
			}
		})
	}
}
