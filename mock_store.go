package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/mock"

	_ "github.com/lib/pq"
)

// MockStore is a mock type for the Storage interface

func NewTestPostgresStore() (*PostgresStore, error) {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	host := os.Getenv("TEST_DB_HOST")
	port, _ := strconv.Atoi(os.Getenv("TEST_DB_PORT"))
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_NAME")

	// Connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)

	// Open database connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	// Check the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Return a new instance of your real store configured for the test database
	return &PostgresStore{db: db}, nil
}

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateAccount(acc *Account) (*Account, error) {
	args := m.Called(acc)
	return args.Get(0).(*Account), args.Error(1)
}
func (m *MockStore) DeleteAccount(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockStore) UpdateAccount(acc *Account) error {
	args := m.Called(acc)
	return args.Error(1)
}

func (m *MockStore) GetAccounts() (accounts []*Account, err error) {
	args := m.Called()
	return args.Get(0).([]*Account), args.Error(1)
}

func (m *MockStore) GetAccountById(id int) (*Account, error) {
	args := m.Called(id)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockStore) GetAccountByEmail(email string) (*Account, error) {

	args := m.Called(email)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockStore) CreateExpense(exp *Expense) (*Expense, error) {
	args := m.Called(exp)
	return args.Get(0).(*Expense), args.Error(1)
}

func (m *MockStore) UpdateExpense(id int, exp *Expense) error {
	args := m.Called(exp)
	return args.Error(0)
}

func (m *MockStore) DeleteExpense(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStore) GetExpenseForUser(id int) ([]*Expense, error) {
	args := m.Called(id)
	return args.Get(0).([]*Expense), args.Error(1)
}

func (m *MockStore) GetAllExpense() ([]*Expense, error) {
	args := m.Called()
	return args.Get(0).([]*Expense), args.Error(1)
}

// NewMockStore creates a new instance of MockStore
func NewMockStore() *MockStore {
	return &MockStore{}
}
