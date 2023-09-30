package main

import (
	"Service/internal/databases"
	"Service/internal/handler"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Start app")

	db, err := databases.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.Default() // Creating a Gin Router

	// Passing db to route handlers via closures (closure)
	r.POST("/invoice", func(c *gin.Context) {
		handler.CreateInvoice(c, db)
	})
	r.POST("/withdraw", func(c *gin.Context) {
		handler.WithdrawFunds(c, db)
	})
	r.GET("/balance", func(c *gin.Context) {
		handler.GetClientBalance(c, db)
	})

	r.Run(":8050")
}
