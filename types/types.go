package types

import (
	"math/rand"
	"time"

	"github.com/uptrace/bun"
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

type LoginResponse struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type ResponceAccount struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
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
		Status:    "Active",
		Password:  string(encpw),
		Number:    int64(rand.Intn(1000000)),
		Balance:   0,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func NewExpense(userId int, expenseName, expensePurpose, expenseCategory string, expenseValue float32, createdAt time.Time) (*Expense, error) {

	/* 	requestDate, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", createdAt.Local().String())
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

	/* 	requestDate, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", createdAt.Local().String())
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

type Account struct {
	bun.BaseModel `bun:"table:account,alias:a"  json:"-"`
	ID            int       `bun:"id,pk,autoincrement" json:"id"`
	FirstName     string    `bun:"first_name" json:"first_name"`
	LastName      string    `bun:"last_name" json:"last_name"`
	Email         string    `bun:"email,unique" json:"email"`
	Password      string    `bun:"password" json:"password"`
	Status        string    `bun:"status" json:"status"`
	Number        int64     `bun:"number" json:"number"`
	Balance       int64     `bun:"balance" json:"balance"`
	CreatedAt     time.Time `bun:"created_at" json:"createdAt"`
}

type Expense struct {
	bun.BaseModel   `bun:"table:expense,alias:e" json:"-"`
	ID              int       `bun:"id,pk,autoincrement" json:"id"`
	UserId          int       `bun:"user_id" json:"user_id"`
	ExpenseName     string    `bun:"expense_name" json:"expense_name"`
	ExpensePurpose  string    `bun:"expense_purpose" json:"expense_purpose"`
	ExpenseCategory string    `bun:"expense_category" json:"expense_category"`
	ExpenseValue    float32   `bun:"expense_value" json:"expense_value"`
	CreatedAt       time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt       time.Time `bun:"updated_at" json:"updated_at"`
	Account         *Account  `bun:"rel:belongs-to,join:user_id=id" json:"-"`
}
