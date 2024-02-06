package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type TransferRequest struct {
	ToAccount int `json: "toAccount"`
	Ammount   int `json: "ammount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type CreateExpenseRequest struct {
	UserId          int     `json:"user_id"`
	ExpenseName     string  `json:"expense_name"`
	ExpensePurpose  string  `json:"expense_purpose"`
	ExpenseCategory string  `json:"expense_category"`
	ExpenseValue    float32 `json:"expense_value"`
}

type UpdateExpenseRequest struct {
	UserId          int       `json:"user_id"`
	ExpenseName     string    `json:"expense_name"`
	ExpensePurpose  string    `json:"expense_purpose"`
	ExpenseCategory string    `json:"expense_category"`
	ExpenseValue    float32   `json:"expense_value"`
	CreatedAt       time.Time `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Number    int64     `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}

type Expense struct {
	ID              int       `json:"id"`
	UserId          int       `json:"user_id"`
	ExpenseName     string    `json:"expense_name"`
	ExpensePurpose  string    `json:"expense_purpose"`
	ExpenseCategory string    `json:"expense_category"`
	ExpenseValue    float32   `json:"expense_value"`
	UpdatedAt       time.Time `json:"createdAt"`
	CreatedAt       time.Time `json:"updatedAt"`
}

func NewAccount(firstName, lastName, email, password string) (*Account, error) {
	encpw, er := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if er != nil {
		return nil, er
	}
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  string(encpw),
		Number:    int64(rand.Intn(1000000)),
		Balance:   0,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func NewExpense(userId int, expenseName, expensePurpose, expenseCategory string, expenseValue float32) (*Expense, error) {
	return &Expense{
		UserId:          userId,
		ExpenseName:     expenseName,
		ExpensePurpose:  expensePurpose,
		ExpenseCategory: expenseCategory,
		ExpenseValue:    expenseValue,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}, nil
}

func UpdatedExpense(userId int, expenseName, expensePurpose, expenseCategory string, expenseValue float32, createdAt time.Time) (*Expense, error) {
	return &Expense{
		UserId:          userId,
		ExpenseName:     expenseName,
		ExpensePurpose:  expensePurpose,
		ExpenseCategory: expenseCategory,
		ExpenseValue:    expenseValue,
		CreatedAt:       createdAt,
		UpdatedAt:       time.Now().UTC(),
	}, nil
}
