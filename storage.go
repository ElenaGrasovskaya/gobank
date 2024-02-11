package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) (*Account, error)
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
	GetAccountByEmail(string) (*Account, error)

	CreateExpense(*Expense) (*Expense, error)
	UpdateExpense(int, *Expense) error
	DeleteExpense(int) error
	GetExpenseForUser(int) ([]*Expense, error)
	GetAllExpense() ([]*Expense, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	fmt.Println("Init DB gobank")

	/* 	appEnv := os.Getenv("APP_ENV")
	   	fmt.Println(appEnv)
	   	if appEnv == "development" {
	   		// Load .env file in development environment
	   		err := godotenv.Load()
	   		if err != nil {
	   			log.Fatal("Error loading .env file")
	   		}
	   	} */

	// Load .env file in development environment
	err := godotenv.Load()
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

	// Check the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	fmt.Println("Successfully connected!")

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.CreateTables()
}

func (s *PostgresStore) CreateTables() error {
	query := `
		create table if not exists account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		email varchar(50),
		password varchar(200),
		number serial,
		balance int,
		created_at timestamp
	);
		create table if not exists expense (
		id serial primary key,
		user_id int,
		expense_name varchar(50),
		expense_purpose varchar(50),
		expense_category varchar(50),
		expense_value float,
		created_at timestamp,
		updated_at timestamp,
		FOREIGN KEY (user_id) REFERENCES account(id)
		);`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) (*Account, error) {
	query := `
    INSERT INTO account
    (first_name, last_name, email, password, number, balance, created_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id;`

	var id int
	err := s.db.QueryRow(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Email,
		acc.Password,
		acc.Number,
		acc.Balance,
		acc.CreatedAt,
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	acc.ID = id
	return acc, nil
}

func (s *PostgresStore) CreateExpense(exp *Expense) (*Expense, error) {

	query := `
    INSERT INTO expense
    (user_id, expense_name, expense_purpose, expense_category, expense_value, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id; `

	var id int
	err := s.db.QueryRow(
		query,
		exp.UserId,
		exp.ExpenseName,
		exp.ExpensePurpose,
		exp.ExpenseCategory,
		exp.ExpenseValue,
		exp.UpdatedAt,
		exp.CreatedAt,
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	exp.ID = id

	return exp, nil
}

func (s *PostgresStore) UpdateAccount(*Account) error {

	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	_, err := s.db.Query("delete from account where id = $1", id)

	return err
}

func (s *PostgresStore) DeleteExpense(id int) error {
	_, err := s.db.Query("delete from expense where id = $1", id)

	return err
}

func (s *PostgresStore) UpdateExpense(id int, newExp *Expense) error {
	fmt.Printf("New data for Expense %v", newExp)
	query := `
    UPDATE expense
    SET expense_name = $1, expense_purpose = $2, expense_category = $3, expense_value = $4, created_at = $5, updated_at = $6
    WHERE id = $7
    `
	result, err := s.db.Exec(
		query,
		newExp.ExpenseName,
		newExp.ExpensePurpose,
		newExp.ExpenseCategory,
		newExp.ExpenseValue,
		newExp.UpdatedAt,
		newExp.CreatedAt,
		id,
	)

	if err != nil {
		return err
	}

	// Optional: Check how many rows were affected
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("No rows affected")
	}

	return nil
}

func (s *PostgresStore) GetAccountById(id int) (*Account, error) {

	rows, err := s.db.Query("select id, first_name, last_name, email, password, number, balance, created_at from account where id=$1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return ScanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) GetAccountByEmail(email string) (*Account, error) {
	rows, err := s.db.Query("select id, first_name, last_name, email, password, number, balance, created_at from account where email=$1", email)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return ScanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %s not found", email)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select id, first_name, last_name, email, password, number, balance, created_at from account")
	if err != nil {
		return nil, err
	}
	accounts := []*Account{}
	for rows.Next() {
		account, err := ScanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func ScanIntoAccount(r *sql.Rows) (*Account, error) {
	account := new(Account)
	err := r.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Email,
		&account.Password,
		&account.Number,
		&account.Balance,
		&account.CreatedAt)

	return account, err
}

func ScanIntoExpense(r *sql.Rows) (*Expense, error) {
	expense := new(Expense)
	err := r.Scan(
		&expense.ID,
		&expense.UserId,
		&expense.ExpenseName,
		&expense.ExpensePurpose,
		&expense.ExpenseCategory,
		&expense.ExpenseValue,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)

	return expense, err
}

func (s *PostgresStore) GetExpenseForUser(id int) ([]*Expense, error) {
	rows, err := s.db.Query("select id, user_id, expense_name, expense_purpose, expense_category, expense_value, created_at, created_at from expense where user_id=$1", id)
	if err != nil {
		return nil, err
	}
	expenses := []*Expense{}
	for rows.Next() {
		expense, err := ScanIntoExpense(rows)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

func (s *PostgresStore) GetAllExpense() ([]*Expense, error) {
	rows, err := s.db.Query("select id, user_id, expense_name, expense_purpose, expense_category, expense_value, created_at, updated_at from expense")
	if err != nil {
		return nil, err
	}
	expenses := []*Expense{}
	for rows.Next() {
		expense, err := ScanIntoExpense(rows)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}
