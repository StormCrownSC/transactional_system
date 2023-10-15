package handler

import (
	"Service/internal/databases"
	"Service/internal/structures"
	"database/sql"
	"fmt"
	"net/http"
	"unicode"

	"github.com/gin-gonic/gin"
)

// CreateInvoice Create Invoice handler for creating an invoice
func CreateInvoice(c *gin.Context, db *sql.DB) {
	// Request and Parameter processing
	var request structures.TransactionRequest
	if err := c.BindJSON(&request); err != nil {
		fmt.Println(err, request)
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid request format"})
		return
	}

	// Calling the CreateInvoice function from the databases package
	if err := databases.CreateInvoice(db, request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	// Return a successful response
	c.JSON(http.StatusOK, gin.H{
		"Success": "The invoice has been created successfully and is awaiting processing",
	})
}

// WithdrawFunds handles the withdrawal of funds
func WithdrawFunds(c *gin.Context, db *sql.DB) {
	// Request and Parameter processing
	var request structures.TransactionRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid request format"})
		return
	}

	// Calling the WithdrawFunds function from the databases package
	if err := databases.WithdrawFunds(db, request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	// Return a successful response
	c.JSON(http.StatusOK, gin.H{
		"Success": "Funds have been withdrawn successfully",
	})
}

// GetClientBalance function for getting the current and frozen customer balance
func GetClientBalance(c *gin.Context, db *sql.DB) {
	// Extracting the request parameters
	clientAccountStr := c.Query("client_account")

	for _, char := range clientAccountStr {
		if !unicode.IsDigit(char) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid client account format"})
			return
		}
	}

	// Executing a query to the database to get the balance
	clientBalances, err := databases.GetClientBalances(db, clientAccountStr)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"Error": "The client's account was not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error when querying the database", "details": err.Error()})
		}
		return
	}

	// Returning a successful response with balances
	c.JSON(http.StatusOK, clientBalances)
}
