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
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
	GetAccountByEmail(string) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	fmt.Println("Init DB gobank")

	appEnv := os.Getenv("APP_ENV")
	fmt.Println(appEnv)
	/* 	if appEnv == "development" {
	   		// Load .env file in development environment
	   		err := godotenv.Load()
	   		if err != nil {
	   			log.Fatal("Error loading .env file")
	   		}
	   	}
	*/

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
	return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {
	query := `create table if not exists account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		email varchar(50),
		password varchar(200),
		number serial,
		balance int,
		created_at timestamp
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {

	query := `
	insert into account
	(first_name, last_name, email, password, number, balance, created_at)
	values ($1, $2, $3, $4, $5, $6, $7)
	`
	resp, err := s.db.Query(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Email,
		acc.Password,
		acc.Number,
		acc.Balance,
		acc.CreatedAt,
	)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)
	return nil
}

func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	_, err := s.db.Query("delete from account where id = $1", id)

	return err
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
