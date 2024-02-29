package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/ElenaGrasovskaya/gobank/types"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Storage interface {
	CreateAccount(context.Context, *types.Account) (*types.Account, error)
	DeleteAccount(context.Context, int) error
	RestoreAccount(context.Context, int) error
	UpdateAccount(context.Context, *types.Account) error
	GetAccounts(context.Context) ([]*types.Account, error)
	GetAccountById(context.Context, int) (*types.Account, error)
	GetAccountByEmail(context.Context, string) (*types.Account, error)

	CreateExpense(context.Context, *types.Expense) (*types.Expense, error)
	UpdateExpense(context.Context, int, *types.Expense) error
	DeleteExpense(context.Context, int) error
	GetExpenseForUser(context.Context, int) ([]*types.Expense, error)
	GetExpenseById(context.Context, int) (*types.Expense, error)
	GetAllExpense(context.Context) ([]*types.Expense, error)
}

type PostgresStore struct {
	Db *bun.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	fmt.Println("Init DB gobank")

	appEnv := os.Getenv("APP_ENV")
	fmt.Println(appEnv)
	if appEnv == "development" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	// Load .env file in development environment
	/* 	err := godotenv.Load()
	   	if err != nil {
	   		log.Fatal("Error loading .env file")
	   	}
	*/
	host := os.Getenv("DB_HOST")
	port, enverr := strconv.Atoi(os.Getenv("DB_PORT"))
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if enverr != nil {
		fmt.Printf("%v", enverr)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require",
		user, password, host, port, dbname)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	fmt.Println("Successfully connected to the database with Bun!")

	return &PostgresStore{Db: db}, nil
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
		status varchar(50),
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
	_, err := s.Db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(ctx context.Context, acc *types.Account) (*types.Account, error) {
	_, err := s.Db.NewInsert().Model(acc).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStore) CreateExpense(ctx context.Context, exp *types.Expense) (*types.Expense, error) {
	_, err := s.Db.NewInsert().Model(exp).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

func (s *PostgresStore) UpdateAccount(context.Context, *types.Account) error {

	return nil
}

func (s *PostgresStore) DeleteAccount(ctx context.Context, id int) error {
	newStatus := "Deleted"
	_, err := s.Db.NewUpdate().
		Model((*types.Account)(nil)).
		Set("status = ?", newStatus).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (s *PostgresStore) RestoreAccount(ctx context.Context, id int) error {
	newStatus := "Active"
	_, err := s.Db.NewUpdate().
		Model((*types.Account)(nil)).
		Set("status = ?", newStatus).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (s *PostgresStore) DeleteExpense(ctx context.Context, id int) error {

	expense := &types.Expense{ID: id}
	_, err := s.Db.NewDelete().Model(expense).Where("? = ?", bun.Ident("id"), id).Exec(ctx)

	return err
}

func (s *PostgresStore) UpdateExpense(ctx context.Context, id int, newExp *types.Expense) error {
	newExp.ID = id

	res, err := s.Db.NewUpdate().
		Model(newExp).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (s *PostgresStore) GetAccountById(ctx context.Context, id int) (*types.Account, error) {
	if id == 0 {
		return nil, fmt.Errorf("account %d not found", id)
	}

	account := new(types.Account)

	err := s.Db.NewSelect().Model(account).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account %d not found", id)
		}
		return nil, err
	}

	return account, nil
}

func (s *PostgresStore) GetExpenseById(ctx context.Context, id int) (*types.Expense, error) {
	if id == 0 {
		return nil, fmt.Errorf("expense %d not found", id)
	}

	expense := new(types.Expense)

	err := s.Db.NewSelect().Model(expense).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("expense %d not found", id)
		}
		return nil, err
	}

	return expense, nil
}

func (s *PostgresStore) GetAccountByEmail(ctx context.Context, email string) (*types.Account, error) {
	if email == "" {
		return nil, fmt.Errorf("account %v not found", email)
	}

	account := new(types.Account)

	err := s.Db.NewSelect().Model(account).Where("email = ?", email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account for %v not found", email)
		}
		return nil, err
	}

	return account, nil
}

func (s *PostgresStore) GetAccounts(ctx context.Context) ([]*types.Account, error) {
	var accounts []*types.Account
	err := s.Db.NewSelect().Model(&accounts).Order("id ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func ScanIntoAccount(r *sql.Rows) (*types.Account, error) {
	account := new(types.Account)
	err := r.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Email,
		&account.Password,
		&account.Status,
		&account.Number,
		&account.Balance,
		&account.CreatedAt)

	return account, err
}

func ScanIntoExpense(r *sql.Rows) (*types.Expense, error) {
	expense := new(types.Expense)
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

func (s *PostgresStore) GetExpenseForUser(ctx context.Context, id int) ([]*types.Expense, error) {
	var expenses []*types.Expense
	err := s.Db.NewSelect().Model(&expenses).Where("user_id = ?", id).Order("id ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return expenses, nil
}

func (s *PostgresStore) GetAllExpense(ctx context.Context) ([]*types.Expense, error) {
	var expenses []*types.Expense
	err := s.Db.NewSelect().Model(&expenses).Order("id ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return expenses, nil
}
