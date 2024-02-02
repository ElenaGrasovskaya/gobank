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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json: "password"`
	Number    int64     `json:"number"`
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
		Password:  string(encpw),
		Number:    int64(rand.Intn(1000000)),
		Balance:   0,
		CreatedAt: time.Now().UTC(),
	}, nil

}
