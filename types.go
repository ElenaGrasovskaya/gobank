package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type CreateExpenseRequest struct {
	ExpenseName     string    `json:"expense_name"`
	ExpensePurpose  string    `json:"expense_purpose"`
	ExpenseCategory string    `json:"expense_category"`
	ExpenseValue    float32   `json:"expense_value"`
	CreatedAt       time.Time `json:"created_at"`
}

type UpdateExpenseRequest struct {
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

type LoginReasponce struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
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
	CreatedAt       time.Time `json:"updated_at"`
	UpdatedAt       time.Time `json:"created_at"`
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

func NewExpense(userId int, expenseName, expensePurpose, expenseCategory string, expenseValue float32, createdAt time.Time) (*Expense, error) {

	/* 	requestDate, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", createdAt)
	   	if err != nil {
	   		return nil, err
	   	} */
	return &Expense{
		ID:              0,
		UserId:          userId,
		ExpenseName:     expenseName,
		ExpensePurpose:  expensePurpose,
		ExpenseCategory: expenseCategory,
		ExpenseValue:    expenseValue,
		CreatedAt:       createdAt,
		UpdatedAt:       time.Now(),
	}, nil
}

func UpdatedExpense(id, userId int, expenseName, expensePurpose, expenseCategory string, expenseValue float32, createdAt time.Time) (*Expense, error) {
	/*
		requestDate, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", createdAt)
		if err != nil {
			return nil, err
		} */
	return &Expense{
		ID:              id,
		UserId:          userId,
		ExpenseName:     expenseName,
		ExpensePurpose:  expensePurpose,
		ExpenseCategory: expenseCategory,
		ExpenseValue:    expenseValue,
		CreatedAt:       createdAt,
		UpdatedAt:       time.Now(),
	}, nil
}
