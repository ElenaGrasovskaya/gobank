package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	// Defining routes
	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin)).Methods(http.MethodPost)
	router.HandleFunc("/expense/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleUpdateExpense), s.store)).Methods(http.MethodPost)
	router.HandleFunc("/expense", withJWTAuth(makeHTTPHandleFunc(s.handleCreateExpense), s.store)).Methods(http.MethodPost)
	router.HandleFunc("/expense", withJWTAuth(makeHTTPHandleFunc(s.handleGetExpenseForUser), s.store)).Methods(http.MethodGet)
	router.HandleFunc("/expense/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleDeleteExpense), s.store)).Methods(http.MethodDelete)
	router.HandleFunc("/expenses", makeHTTPHandleFunc(s.handleGetAllExpense)).Methods(http.MethodGet)
	router.HandleFunc("/register", makeHTTPHandleFunc(s.handleRegister)).Methods(http.MethodPost)
	router.HandleFunc("/logout", makeHTTPHandleFunc(s.handleLogout)).Methods(http.MethodPost)
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleGetAccount)).Methods(http.MethodGet)
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleCreateAccount)).Methods(http.MethodPost)
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleDeleteAccount)).Methods(http.MethodDelete)
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountById), s.store)).Methods(http.MethodGet)

	// Apply the CORS middleware to the router
	corsRouter := corsMiddleware(router)

	fmt.Println("JSON API server is running on port: " + s.listenAddr)
	http.ListenAndServe(s.listenAddr, corsRouter)
}

// ACCOUNT MANAGING *********************************************************

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {

	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getId(r)
		if err != nil {
			return err
		}
		account, err := s.store.GetAccountById(id)
		if err != nil {
			return err
		}
		fmt.Println(id)
		return WriteJSON(w, http.StatusOK, account)

	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountRequest := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountRequest); err != nil {
		return err
	}

	account, err := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.Email, createAccountRequest.Password)
	if err != nil {
		return err
	}

	newAcc, err := s.store.CreateAccount(account)

	if err != nil {
		return err
	}

	tokenString, err := createJWT(newAcc)
	if err != nil {
		return err
	}

	fmt.Printf("JWT token: %v", tokenString)
	return WriteJSON(w, http.StatusOK, newAcc)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

//EXPENCE SERVICES ****************************************************************************

func (s *APIServer) handleGetAllExpense(w http.ResponseWriter, r *http.Request) error {

	expenses, err := s.store.GetAllExpense()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, expenses)
}

func (s *APIServer) handleGetExpenseForUser(w http.ResponseWriter, r *http.Request) error {
	userId, err := getIdFromCookie(w, r)
	if err != nil {
		return err
	}
	expenses, err := s.store.GetExpenseForUser(userId)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, expenses)
}

func (s *APIServer) handleCreateExpense(w http.ResponseWriter, r *http.Request) error {
	createExpenseRequest := new(CreateExpenseRequest)
	if err := json.NewDecoder(r.Body).Decode(createExpenseRequest); err != nil {
		return err
	}

	userId, err := getIdFromCookie(w, r)
	if err != nil {
		return err
	}

	expense, err := NewExpense(userId, createExpenseRequest.ExpenseName, createExpenseRequest.ExpensePurpose, createExpenseRequest.ExpenseCategory, createExpenseRequest.ExpenseValue)
	if err != nil {
		return err
	}

	newExp, err := s.store.CreateExpense(expense)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, newExp)
}

func (s *APIServer) handleUpdateExpense(w http.ResponseWriter, r *http.Request) error {
	updateExpenseRequest := new(UpdateExpenseRequest)

	if err := json.NewDecoder(r.Body).Decode(updateExpenseRequest); err != nil {
		return err
	}

	id, err := getId(r)
	if err != nil {
		return err
	}

	expense, err := UpdatedExpense(updateExpenseRequest.UserId, updateExpenseRequest.ExpenseName, updateExpenseRequest.ExpensePurpose, updateExpenseRequest.ExpenseCategory, updateExpenseRequest.ExpenseValue, updateExpenseRequest.CreatedAt)
	if err != nil {
		return err
	}

	if err := s.store.UpdateExpense(id, expense); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, expense)
}

func (s *APIServer) handleDeleteExpense(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}

	if err := s.store.DeleteExpense(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

//AUTHORIZATION SERVICES**********************************************************

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	account, err := s.store.GetAccountByEmail(req.Email)
	if err != nil {
		clearSession(w)
		return WriteJSON(w, http.StatusExpectationFailed, err)
	}

	comparePass, err := EncrPassword(req.Password, account.Password)

	if err != nil {
		clearSession(w)
		return WriteJSON(w, http.StatusResetContent, err)
	}

	if comparePass {
		account, err := s.store.GetAccountByEmail(req.Email)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, err)
		}
		setSession(account.ID, req.Email, w)

		return WriteJSON(w, http.StatusOK, "User logged in")
	}

	return WriteJSON(w, http.StatusBadRequest, comparePass)
}

func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) error {
	createAccountRequest := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountRequest); err != nil {
		return err
	}

	account, err := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.Email, createAccountRequest.Password)
	if err != nil {
		return err
	}

	existAccount, err := s.store.GetAccountByEmail(createAccountRequest.Email)
	if err != nil {
		newAcc, err := s.store.CreateAccount(account)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusAccepted, newAcc)
	} else {
		return WriteJSON(w, http.StatusConflict, existAccount)
	}

}

func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) error {

	clearSession(w)
	return WriteJSON(w, http.StatusLocked, "User Logged Out")
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}
	secrtet := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secrtet))
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")
		cookies, err := r.Cookie("token")

		if err != nil {
			fmt.Println(err)
		}

		token, err := validateJWT(cookies.Value)
		if err != nil {
			PermissionDenied(w)
			return
		}

		if !token.Valid {
			PermissionDenied(w)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			id := claims["id"].(float64)
			email := claims["email"].(string)
			account, err := s.GetAccountById(int(id))

			if err != nil {
				PermissionDenied(w)
				return
			}
			if account.Email != email {
				PermissionDenied(w)
				return
			}
			fmt.Printf("%s %v", email, id)
			handlerFunc(w, r)

		} else {
			fmt.Errorf("invalid token")
		}

	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

func setSession(userId int, userEmail string, response http.ResponseWriter) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userId,
		"email": userEmail,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		fmt.Println("failed to create token")
		return
	} else {

		cookie := &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Path:     "/",
			MaxAge:   1600,
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		}
		fmt.Println("token created")
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

func EncrPassword(reqPassword string, dbPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(reqPassword))
	if err != nil {
		return false, err
	} else {
		fmt.Println("passwords match")
		return true, nil
	}

}

//UTILS  ********************************************************************************************

func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}

	return id, nil
}

func getIdFromCookie(w http.ResponseWriter, r *http.Request) (int, error) {
	cookies, err := r.Cookie("token")

	if err != nil {
		fmt.Println(err)
	}

	token, err := validateJWT(cookies.Value)
	if err != nil {
		PermissionDenied(w)
		return 0, err
	}

	if !token.Valid {
		PermissionDenied(w)
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["id"].(float64)
		return int(id), nil

	} else {
		PermissionDenied(w)
		return 0, err
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func PermissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "Invalid Token "})
}

type ApiError struct {
	Error string `json: "error"`
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func corsMiddleware(next http.Handler) http.Handler {

	var allowedOrigins = map[string]bool{
		"http://localhost:5173":              true,
		"https://gobank-frontend.vercel.app": true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		// Set CORS headers

		// Check if it's a preflight request
		if r.Method == "OPTIONS" {
			// Send response for preflight request
			w.WriteHeader(http.StatusOK)
			return
		}

		// Serve the next handler
		next.ServeHTTP(w, r)
	})
}
