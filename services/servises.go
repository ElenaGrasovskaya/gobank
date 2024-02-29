package services

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/ElenaGrasovskaya/gobank/storage"
	"github.com/ElenaGrasovskaya/gobank/types"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type ServiceHandlers interface {
	HandleHealth(*gin.Context)
	HandleRegister(*gin.Context)
	HandleLogout(*gin.Context)
	HandleDeleteExpense(*gin.Context)
	HandleUpdateExpense(*gin.Context)
}

type StoreHandler struct {
	store storage.Storage
}

func NewServiceHandler(store storage.Storage) *StoreHandler {
	return &StoreHandler{
		store: store,
	}
}

func (s *StoreHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, "Welcome to GoBank API")
}

func (s *StoreHandler) HandleLogin(c *gin.Context) {
	var req types.LoginRequest
	stdCtx := c.Request.Context()
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := s.store.GetAccountByEmail(stdCtx, req.Email)
	if err != nil {
		clearSession(c)
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}

	comparePass, err := encrPassword(req.Password, account.Password)
	if err != nil {
		clearSession(c)
		c.JSON(http.StatusResetContent, gin.H{"error": err.Error()})
		return
	}

	if comparePass {

		setSession(c, account.ID, req.Email)
		userResponse := types.LoginResponse{
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

func (s *StoreHandler) HandleRegister(c *gin.Context) {
	createAccountRequest := new(types.CreateAccountRequest)
	stdCtx := c.Request.Context()
	if err := c.ShouldBindJSON(createAccountRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := types.NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.Email, createAccountRequest.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	existAccount, err := s.store.GetAccountByEmail(stdCtx, createAccountRequest.Email)
	if err == nil && existAccount != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Account " + existAccount.Email + " already exists"})
		return
	}

	newAcc, err := s.store.CreateAccount(stdCtx, account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store new account"})
		return
	}

	c.JSON(http.StatusAccepted, newAcc)
}

func (s *StoreHandler) HandleLogout(c *gin.Context) {
	fmt.Printf("Logging out")
	clearSession(c)
	c.JSON(http.StatusOK, "User logged out")
}

func CreateJWT(account *types.Account) (string, error) {
	claims := &jwt.MapClaims{
		"id":    account.ID,
		"email": account.Email,
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
		fmt.Println("token created")
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie("token", tokenString, 30*24*60*60, "/", "", true, true)

	}
}

func clearSession(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", true, true)
}

func encrPassword(reqPassword string, dbPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(reqPassword))
	if err != nil {
		return false, err
	} else {
		fmt.Println("passwords match")
		return true, nil
	}

}

//UTILS  ********************************************************************************************

func GetId(c *gin.Context) (int, error) {
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

func GetIdFromCookie(c *gin.Context) (int, error) {
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

type ApiError struct {
	Error string `json:"error"`
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

func WithJWTAuthMiddleware(s storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("calling JWT auth middleware")
		stdCtx := c.Request.Context()

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

			account, err := s.GetAccountById(stdCtx, int(id))
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
