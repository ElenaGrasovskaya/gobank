package tests

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"bytes"

	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ElenaGrasovskaya/gobank/expense"
	"github.com/ElenaGrasovskaya/gobank/router"
	"github.com/ElenaGrasovskaya/gobank/services"
	"github.com/ElenaGrasovskaya/gobank/storage"
	"github.com/ElenaGrasovskaya/gobank/types"
)

func NewTestPostgresStore() (*storage.PostgresStore, error) {
	err := godotenv.Load("../.env")

	fmt.Println("load env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	port, enverr := strconv.Atoi(os.Getenv("DB_PORT"))
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if enverr != nil {
		fmt.Printf("%v", enverr)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &storage.PostgresStore{
		Db: db,
	}, nil
}

func InitializeTestServer() *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	store, err := NewTestPostgresStore()
	if err != nil {
		log.Fatalf("Failed to initialize the test store: %v", err)
	}
	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize the test store: %v", err)
	}

	r := router.SetupRouter(store)
	return r
}

func TestHandleLogin(t *testing.T) {
	router := InitializeTestServer()

	email := "elenagrasovskaya@gmail.com"

	userData := &types.LoginRequest{
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
	router := InitializeTestServer()

	email := "testing@gmail.com"

	userData := &types.CreateAccountRequest{

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
}

func TestHandleLogout(t *testing.T) {
	router := InitializeTestServer()
	//Case 1 logout without a cookie
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
	router := InitializeTestServer()
	mockAccount := &types.Account{
		ID:        7,
		FirstName: "Test",
		LastName:  "Testovich",
		Email:     "testing@gmail.com",
		Password:  "$2a$10$q/cjukk2QtKtTdcaype0UOgPydr5MRcQm9wmbpfvyDksUuuv2gomu",
		Status:    "Active",
		Number:    112302,
		Balance:   0,
		CreatedAt: time.Now(),
	}

	tokenString, err := services.CreateJWT(mockAccount)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: tokenString,
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/accounts", nil)
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected Status.OK and accounts data; got %v", w.Code)
	}
}

func TestHandleGetAccountById(t *testing.T) {
	router := InitializeTestServer()

	testID := 7
	mockAccount := &types.Account{
		ID:        7,
		FirstName: "Test",
		LastName:  "Testovich",
		Email:     "testing@gmail.com",
		Password:  "$2a$10$q/cjukk2QtKtTdcaype0UOgPydr5MRcQm9wmbpfvyDksUuuv2gomu",
		Status:    "Active",
		Number:    112302,
		Balance:   0,
		CreatedAt: time.Now(),
	}

	tokenString, err := services.CreateJWT(mockAccount)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: tokenString,
	}

	// Test 1: Valid request

	req, _ := http.NewRequest("GET", fmt.Sprintf("/account/%d", testID), nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

	var response types.ResponceAccount
	err = json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(t, err, "Expected no error unmarshalling response")
	assert.Equal(t, mockAccount.ID, response.ID, "Expected account ID to match mock account")
	assert.Equal(t, mockAccount.Email, response.Email, "Expected account email to match mock account")

	// Test 2: Invalid request
	testInvalidId := "test"
	req, _ = http.NewRequest("GET", fmt.Sprintf("/account/%v", testInvalidId), nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")

	//Test 3: Invalid request starting with numbers

	testInvalidId = "75test"
	req, _ = http.NewRequest("GET", fmt.Sprintf("/account/%v", testInvalidId), nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")

	//Test 4: Invalid request starting with numbers

	testInvalidId = "152"
	req, _ = http.NewRequest("GET", fmt.Sprintf("/account/%v", testInvalidId), nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")

	//Test 5: Account doesn't exist

	testID = 0
	req, _ = http.NewRequest("GET", fmt.Sprintf("/account/%d", testID), nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400, got %v", w.Code)
}

func TestHandleDeleteAccount(t *testing.T) {
	router := InitializeTestServer()

	testID := "7"
	mockAccount := &types.Account{
		ID:        7,
		FirstName: "Test",
		LastName:  "Testovich",
		Email:     "testing@gmail.com",
		Password:  "$2a$10$q/cjukk2QtKtTdcaype0UOgPydr5MRcQm9wmbpfvyDksUuuv2gomu",
		Status:    "Active",
		Number:    112302,
		Balance:   0,
		CreatedAt: time.Now(),
	}

	tokenString, err := services.CreateJWT(mockAccount)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: tokenString,
	}

	tests := []struct {
		description  string
		accountID    string
		expectedCode int
		expectedBody string
	}{
		//{"Delete existing account", testID, http.StatusOK, "{\"deleted\":7}"},
		{"Delete non-existing account", "0", http.StatusBadRequest, ""},
		{"Invalid account ID", "abc", http.StatusBadRequest, ""},
		{"Delete already deleted account", testID, http.StatusBadRequest, "{\"This accout was already deleted\": 7}"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/account/"+test.accountID, nil)
			req.AddCookie(cookie)
			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedCode, w.Code)

			if test.expectedBody != "" {
				assert.JSONEq(t, test.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHandleGetAllExpense(t *testing.T) {
	router := InitializeTestServer()
	mockAccount := &types.Account{
		ID:        7,
		FirstName: "Test",
		LastName:  "Testovich",
		Email:     "testing@gmail.com",
		Password:  "$2a$10$q/cjukk2QtKtTdcaype0UOgPydr5MRcQm9wmbpfvyDksUuuv2gomu",
		Status:    "Active",
		Number:    112302,
		Balance:   0,
		CreatedAt: time.Now(),
	}

	tokenString, err := services.CreateJWT(mockAccount)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: tokenString,
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/expenses", nil)
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected Status.OK and expenses data; got %v", w.Code)
	}
}

func TestHandleGetExpenseForUser(t *testing.T) {
	router := InitializeTestServer()

	mockAccount := &types.Account{
		ID:        7,
		FirstName: "Test",
		LastName:  "Testovich",
		Email:     "testing@gmail.com",
		Password:  "$2a$10$q/cjukk2QtKtTdcaype0UOgPydr5MRcQm9wmbpfvyDksUuuv2gomu",
		Status:    "Active",
		Number:    112302,
		Balance:   0,
		CreatedAt: time.Now(),
	}

	tokenString, err := services.CreateJWT(mockAccount)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: tokenString,
	}

	// Test 1: Not authorized request

	req, _ := http.NewRequest("GET", "/expense", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code, "Expected status code 403")

	// Test 2: Valid request

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/expense", nil)
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 and expense data for the account")

	var response []types.Expense
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Expected no error unmarshalling response")

	// Test 3: Invalid request
	testInvalidId := "test"
	req, _ = http.NewRequest("GET", fmt.Sprintf("/expense/%v", testInvalidId), nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")
}

func TestHandleCreateExpense(t *testing.T) {
	router := InitializeTestServer()

	expenseData := &types.CreateExpenseRequest{
		ExpenseName:     "test",
		ExpensePurpose:  "test",
		ExpenseCategory: "test",
		ExpenseValue:    100,
		CreatedAt:       time.Now(),
	}

	mockAccount := &types.Account{
		ID:        7,
		FirstName: "Test",
		LastName:  "Testovich",
		Email:     "testing@gmail.com",
		Password:  "$2a$10$q/cjukk2QtKtTdcaype0UOgPydr5MRcQm9wmbpfvyDksUuuv2gomu",
		Status:    "Active",
		Number:    112302,
		Balance:   0,
		CreatedAt: time.Now(),
	}

	tokenString, err := services.CreateJWT(mockAccount)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: tokenString,
	}

	body, err := json.Marshal(expenseData)
	if err != nil {
		t.Fatalf("Failed to marshal request data: %v", err)
	}
	// Test 1: Not authorized request

	req, _ := http.NewRequest("POST", "/expense", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code, "Expected status code 403")

	// Test 2: Valid request

	req, _ = http.NewRequest("POST", "/expense", bytes.NewBuffer(body))
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 and new expense data")

	// Test 3: Invalid request
	testInvalidId := "test"
	req, _ = http.NewRequest("GET", fmt.Sprintf("/expense%v", testInvalidId), bytes.NewBuffer(body))
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")
}

func TestHandleUpdateExpense(t *testing.T) {
	router := InitializeTestServer()
	newExpenseData := &types.Expense{
		UserId:          7,
		ExpenseName:     "test",
		ExpensePurpose:  "test",
		ExpenseCategory: "test",
		ExpenseValue:    100,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	e := expense.NewExpenseHandler(store)

	newExp, err := e.CreateExpense(newExpenseData)

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
}

/* func TestHandleDeleteExpense(t *testing.T) {
	router := InitializeTestServer()

	//router.DELETE("/expense/:id", server.handleDeleteExpense)

	testExpense, err := types.NewExpense(7, "test", "test", "test", 100, time.Now())
	assert.NoError(t, err, "Expected no error creating new expense")

	//newExpense, newErr := server.store.CreateExpense(testExpense)
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
} */
