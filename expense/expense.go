package expense

import (
	"fmt"
	"net/http"

	"github.com/ElenaGrasovskaya/gobank/services"
	"github.com/ElenaGrasovskaya/gobank/storage"
	"github.com/ElenaGrasovskaya/gobank/types"
	"github.com/gin-gonic/gin"
)

type ExpenseHandlers interface {
	HandleGetAllExpense(*gin.Context)
	HandleGetExpenseForUser(*gin.Context)
	HandleCreateExpense(*gin.Context)
	HandleDeleteExpense(*gin.Context)
	HandleUpdateExpense(*gin.Context)
}

type StoreHandler struct {
	store storage.Storage
}

func NewExpenseHandler(store storage.Storage) *StoreHandler {
	return &StoreHandler{
		store: store,
	}
}

func (s *StoreHandler) HandleGetAllExpense(c *gin.Context) {
	stdCtx := c.Request.Context()
	expenses, err := s.store.GetAllExpense(stdCtx)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not load the data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expenses)
}

func (s *StoreHandler) HandleGetExpenseForUser(c *gin.Context) {
	stdCtx := c.Request.Context()
	userId, err := services.GetIdFromCookie(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	expenses, err := s.store.GetExpenseForUser(stdCtx, userId)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expenses)
}

func (s *StoreHandler) HandleCreateExpense(c *gin.Context) {
	createExpenseRequest := new(types.CreateExpenseRequest)
	fmt.Printf("WE GET %v", createExpenseRequest)
	stdCtx := c.Request.Context()
	if err := c.ShouldBindJSON(createExpenseRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Request %v \n", createExpenseRequest)

	userId, err := services.GetIdFromCookie(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to retrieve user ID from cookie"})
		return
	}

	expense, err := types.NewExpense(userId, createExpenseRequest.ExpenseName, createExpenseRequest.ExpensePurpose, createExpenseRequest.ExpenseCategory, createExpenseRequest.ExpenseValue, createExpenseRequest.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new expense"})
		return
	}

	fmt.Printf("Prepared new expense %v \n", expense)

	newExp, err := s.store.CreateExpense(stdCtx, expense)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store new expense"})
		return
	}

	c.JSON(http.StatusOK, newExp)
}

func (s *StoreHandler) HandleUpdateExpense(c *gin.Context) {
	updateExpenseRequest := new(types.UpdateExpenseRequest)
	stdCtx := c.Request.Context()
	if err := c.ShouldBindJSON(updateExpenseRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := services.GetId(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retrieve user ID from cookie"})
		return
	}

	userId, err := services.GetIdFromCookie(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to retrieve user ID from cookie"})
		return
	}
	expense, err := types.UpdatedExpense(id, userId, updateExpenseRequest.ExpenseName, updateExpenseRequest.ExpensePurpose, updateExpenseRequest.ExpenseCategory, updateExpenseRequest.ExpenseValue, updateExpenseRequest.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build an updated expense"})
		return
	}

	if err := s.store.UpdateExpense(stdCtx, id, expense); err != nil {
		fmt.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expense"})
		return
	}

	c.JSON(http.StatusOK, expense)
}

func (s *StoreHandler) HandleDeleteExpense(c *gin.Context) {
	stdCtx := c.Request.Context()
	id, err := services.GetId(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to get id from the request"})
		return
	}

	expense, err := s.store.GetExpenseById(stdCtx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to find the requested expense"})
		return
	}

	if err := s.store.DeleteExpense(stdCtx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete an expense"})
		return
	}
	c.JSON(http.StatusOK, map[string]int{"deleted": expense.ID})
}
