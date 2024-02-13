package main

import (
	"github.com/stretchr/testify/mock"
)

// MockStore is a mock type for the Storage interface
type MockStore struct {
	mock.Mock
}

// Define methods for each interface method
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
