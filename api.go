package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
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

// ACCOUNT MANAGING *********************************************************

func (s *APIServer) handleGetAccount(c *gin.Context) {

	accounts, err := s.store.GetAccounts()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not load the data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)

}

func (s *APIServer) handleGetAccountById(c *gin.Context) {

	if c.Request.Method == "GET" {
		id, err := getId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Could not load the data from request": err.Error()})
			return
		}
		account, err := s.store.GetAccountById(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Could not load the data from request": err.Error()})
			return
		}
		fmt.Println(id)
		c.JSON(http.StatusOK, account)
	}
	if c.Request.Method == "DELETE" {
		s.handleDeleteAccount(c)
	}
	c.JSON(http.StatusMethodNotAllowed, gin.H{"Method not allowed": "Error"})
}

func (s *APIServer) handleCreateAccount(c *gin.Context) {
	createAccountRequest := new(CreateAccountRequest)
	if err := c.ShouldBindJSON(createAccountRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.Email, createAccountRequest.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not load the data from request": err.Error()})
		return
	}

	newAcc, err := s.store.CreateAccount(account)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not create an account": err.Error()})
		return
	}

	tokenString, err := createJWT(newAcc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not create token": err.Error()})
		return
	}

	fmt.Printf("JWT token: %v", tokenString)
	c.JSON(http.StatusOK, newAcc)
}

func (s *APIServer) handleDeleteAccount(c *gin.Context) {
	id, err := getId(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not delete an account": err.Error()})
		return
	}

	if err := s.store.DeleteAccount(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not delete an account": err.Error()})
	}

	c.JSON(http.StatusOK, map[string]int{"deleted": id})
}

//EXPENCE SERVICES ****************************************************************************

func (s *APIServer) handleGetAllExpense(c *gin.Context) {

	expenses, err := s.store.GetAllExpense()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not load the data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expenses)
}

func (s *APIServer) handleGetExpenseForUser(c *gin.Context) {
	userId, err := getIdFromCookie(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	expenses, err := s.store.GetExpenseForUser(userId)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expenses)
}

func (s *APIServer) handleCreateExpense(c *gin.Context) {
	createExpenseRequest := new(CreateExpenseRequest)
	if err := c.ShouldBindJSON(createExpenseRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Request %v \n", createExpenseRequest)

	userId, err := getIdFromCookie(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to retrieve user ID from cookie"})
		return
	}

	expense, err := NewExpense(userId, createExpenseRequest.ExpenseName, createExpenseRequest.ExpensePurpose, createExpenseRequest.ExpenseCategory, createExpenseRequest.ExpenseValue, createExpenseRequest.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new expense"})
		return
	}

	fmt.Printf("Prepared new expense %v \n", expense)

	newExp, err := s.store.CreateExpense(expense)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store new expense"})
		return
	}

	c.JSON(http.StatusOK, newExp)
}

func (s *APIServer) handleUpdateExpense(c *gin.Context) {
	updateExpenseRequest := new(UpdateExpenseRequest)

	if err := c.ShouldBindJSON(updateExpenseRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := getId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to retrieve user ID from cookie"})
		return
	}

	userId, err := getIdFromCookie(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to retrieve user ID from cookie"})
		return
	}
	expense, err := UpdatedExpense(id, userId, updateExpenseRequest.ExpenseName, updateExpenseRequest.ExpensePurpose, updateExpenseRequest.ExpenseCategory, updateExpenseRequest.ExpenseValue, updateExpenseRequest.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build an updated expense"})
		return
	}

	if err := s.store.UpdateExpense(id, expense); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expense"})
		return
	}

	c.JSON(http.StatusOK, expense)
}

func (s *APIServer) handleDeleteExpense(c *gin.Context) {
	id, err := getId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get id from the request"})
		return
	}

	if err := s.store.DeleteExpense(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete an expense"})
		return
	}
	c.JSON(http.StatusOK, map[string]int{"deleted": id})
}

//AUTHORIZATION SERVICES**********************************************************

func (s *APIServer) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := s.store.GetAccountByEmail(req.Email)
	if err != nil {
		clearSession(c)
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}

	comparePass, err := EncrPassword(req.Password, account.Password)
	if err != nil {
		clearSession(c)
		c.JSON(http.StatusResetContent, gin.H{"error": err.Error()})
		return
	}

	if comparePass {

		setSession(c, account.ID, req.Email)
		userResponse := LoginResponse{
			ID:        account.ID,
			FirstName: account.FirstName,
			LastName:  account.LastName,
			Email:     account.Email,
		}

		c.JSON(http.StatusOK, userResponse)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials"})
	}
}

func (s *APIServer) handleRegister(c *gin.Context) {
	createAccountRequest := new(CreateAccountRequest)
	if err := c.ShouldBindJSON(createAccountRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.Email, createAccountRequest.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	existAccount, err := s.store.GetAccountByEmail(createAccountRequest.Email)
	if err == nil && existAccount != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Account " + existAccount.Email + " already exists"})
		return
	}

	newAcc, err := s.store.CreateAccount(account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store new account"})
		return
	}

	c.JSON(http.StatusAccepted, newAcc)
}

func (s *APIServer) handleLogout(c *gin.Context) {
	fmt.Printf("Logging out")
	clearSession(c)
	c.JSON(http.StatusOK, "User logged out")
	return
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

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}
func setSession(c *gin.Context, userId int, userEmail string) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userId,
		"email": userEmail,
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session token"})
		return
	} else {

		/* cookie := &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Path:     "/",
			MaxAge:   30 * 24 * 60 * 60,
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		} */
		fmt.Println("token created")
		c.SetCookie("token", tokenString, 30*24*60*60, "/", "", false, true)
	}
}

func clearSession(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
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

func getId(c *gin.Context) (int, error) {
	idStr, boolErr := c.Params.Get("id")
	if !boolErr {
		c.JSON(http.StatusExpectationFailed, gin.H{"no params": "Error"})
		return 0, nil
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}

	return id, nil
}

func getIdFromCookie(c *gin.Context) (int, error) {
	cookie, err := c.Cookie("token")
	if err != nil {
		fmt.Println(err)
		return 0, fmt.Errorf("failed to retrieve cookie: %v", err)
	}
	token, err := validateJWT(cookie)
	if err != nil {

		return 0, fmt.Errorf("permission denied: %v", err)
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"].(float64)
		if !ok {
			return 0, fmt.Errorf("error parsing user ID from token")
		}
		return int(id), nil
	} else {
		return 0, fmt.Errorf("permission denied")
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
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var allowedOrigins = map[string]bool{
			"http://localhost:5173":                      true,
			"http://localhost:5173/expense":              true,
			"https://gobank-frontend.vercel.app":         true,
			"https://gobank-frontend.vercel.app/expense": true,
		}

		origin := c.GetHeader("Origin")
		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			c.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		// Set CORS headers

		// Check if it's a preflight request
		if c.Request.Method == "OPTIONS" {
			// Send response for preflight request
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func WithJWTAuthMiddleware(s Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("calling JWT auth middleware")
		// Using Gin's context to get cookie
		cookie, err := c.Cookie("token")
		if err != nil {
			fmt.Println(err)
			permissionDenied(c)
			return
		}

		token, err := validateJWT(cookie)
		if err != nil {
			permissionDenied(c)
			return
		}

		if !token.Valid {
			permissionDenied(c)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			id := claims["id"].(float64)
			email := claims["email"].(string)

			account, err := s.GetAccountById(int(id))
			if err != nil {
				permissionDenied(c)
				return
			}

			if account.Email != email {
				permissionDenied(c)
				return
			}

			fmt.Printf("%s %v", email, id)
			// If authentication is successful, proceed with the request
			c.Next()
		} else {
			permissionDenied(c)
		}
	}
}

// A helper function to handle permission denied response
func permissionDenied(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
	c.Abort() // Prevent calling any subsequent handlers
}
