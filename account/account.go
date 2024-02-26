package account

import (
	"fmt"
	"net/http"

	"github.com/ElenaGrasovskaya/gobank/services"
	"github.com/ElenaGrasovskaya/gobank/storage"
	"github.com/ElenaGrasovskaya/gobank/types"
	"github.com/gin-gonic/gin"
)

type AccountHandlers interface {
	HandleGetAccount(*gin.Context)
	HandleGetAccountById(*gin.Context)
	HandleCreateAccount(*gin.Context)
	HandleDeleteAccount(*gin.Context)
}

type StoreHandler struct {
	store storage.Storage
}

func NewAccountHandler(store storage.Storage) *StoreHandler {
	return &StoreHandler{
		store: store,
	}
}

func (s *StoreHandler) HandleGetAccount(c *gin.Context) {
	stdCtx := c.Request.Context()
	accounts, err := s.store.GetAccounts(stdCtx)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not load the data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)

}

func (s *StoreHandler) HandleGetAccountById(c *gin.Context) {
	stdCtx := c.Request.Context()
	id, err := services.GetId(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not load the data from request": err.Error()})
		return
	}

	account, err := s.store.GetAccountById(stdCtx, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not load the data from request": err.Error()})
		return
	}

	var responceAccount = types.ResponceAccount{
		ID:        account.ID,
		FirstName: account.FirstName,
		LastName:  account.LastName,
		Email:     account.Email,
		Balance:   account.Balance,
		CreatedAt: account.CreatedAt,
	}

	fmt.Println(id)
	c.JSON(http.StatusOK, responceAccount)
}

func (s *StoreHandler) HandleCreateAccount(c *gin.Context) {
	stdCtx := c.Request.Context()
	createAccountRequest := new(types.CreateAccountRequest)
	if err := c.ShouldBindJSON(createAccountRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := types.NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.Email, createAccountRequest.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not load the data from request": err.Error()})
		return
	}

	newAcc, err := s.store.CreateAccount(stdCtx, account)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not create an account": err.Error()})
		return
	}

	tokenString, err := services.CreateJWT(newAcc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not create token": err.Error()})
		return
	}

	fmt.Printf("JWT token: %v", tokenString)
	c.JSON(http.StatusOK, newAcc)
}

func (s *StoreHandler) HandleDeleteAccount(c *gin.Context) {
	id, err := services.GetId(c)
	stdCtx := c.Request.Context()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Invalid account number": err.Error()})
		return
	}

	account, err := s.store.GetAccountById(stdCtx, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not delete an account": err.Error()})
		return
	}

	if account.Status == "Deleted" {
		c.JSON(http.StatusBadRequest, gin.H{"This accout was already deleted": account.ID})
		return
	}
	fmt.Printf("After check %v", account.Status)
	if err := s.store.DeleteAccount(stdCtx, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not delete an account": err.Error()})
	}

	c.JSON(http.StatusOK, map[string]int{"deleted": id})
}
