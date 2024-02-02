package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50TnVtYmVyIjo2Mjk4OCwiZXhwaXJlc0F0IjoxNTAwMH0.7pUeYeHFKe30HlfIQXKPnFZIEZ2uBmPjK20dyr_wrWU
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

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}
	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}

	fmt.Printf("JWT token: %v", tokenString)
	return WriteJSON(w, http.StatusOK, account)
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

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)

	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()
	return WriteJSON(w, http.StatusOK, transferReq)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	account, err := s.store.GetAccountByEmail(req.Email)
	if err != nil {
		clearSession(w)
		return WriteJSON(w, http.StatusExpectationFailed, req)
	}

	compare, err := EncrPassword(req.Password, account.Password)

	if err != nil {
		clearSession(w)
		return WriteJSON(w, http.StatusResetContent, err)
	}

	if compare {
		setSession(req.Email, w)

		return WriteJSON(w, http.StatusAccepted, req)
	}

	return WriteJSON(w, http.StatusBadRequest, compare)
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
		if err := s.store.CreateAccount(account); err != nil {
			return err
		}
		return WriteJSON(w, http.StatusAccepted, account)
	} else {
		return WriteJSON(w, http.StatusConflict, existAccount)
	}

}

func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) error {

	clearSession(w)
	return WriteJSON(w, http.StatusLocked, "User Logged Out")
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	fmt.Printf("This is the method %s", r.Method)
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	// Define your routes
	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin)).Methods(http.MethodPost)
	router.HandleFunc("/register", makeHTTPHandleFunc(s.handleRegister)).Methods(http.MethodPost)
	router.HandleFunc("/logout", makeHTTPHandleFunc(s.handleLogout)).Methods(http.MethodPost)
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountById), s.store)).Methods(http.MethodPost)
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer))

	// Apply the CORS middleware to the router
	corsRouter := corsMiddleware(router)

	fmt.Println("JSON API server is running on port: " + s.listenAddr)
	http.ListenAndServe(s.listenAddr, corsRouter)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

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

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}
	secrtet := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secrtet))
}

func PermissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "Invalid Token "})
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")

		token, err := validateJWT(tokenString)
		if err != nil {
			PermissionDenied(w)
			return
		}

		if !token.Valid {
			PermissionDenied(w)
			return
		}

		userId, err := getId(r)
		if err != nil {
			PermissionDenied(w)
			return
		}
		account, err := s.GetAccountById(userId)
		if err != nil {
			PermissionDenied(w)
			return
		}
		fmt.Println(account)
		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int64(claims["accountNumber"].(float64)) {
			PermissionDenied(w)
			return
		}
		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

type ApiError struct {
	Error string `json: "error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}

	return id, nil
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
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
