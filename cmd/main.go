package main

import (
	"Service/internal/databases"
	"Service/internal/handler"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	log.Println("Start app")

	db, err := databases.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	pool, err := databases.ConnectPgxDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.Default() // Creating a Gin Router

	// We specify the path to static files and the route for servicing static files
	r.Use(static.Serve("/", static.LocalFile("./assets", true)))
	// Passing db to route handlers via closures (closure)
	r.POST("/invoice", func(c *gin.Context) {
		handler.CreateInvoice(c, db)
	})
	r.POST("/withdraw", func(c *gin.Context) {
		handler.WithdrawFunds(c, db)
	})
	r.GET("/balance", func(c *gin.Context) {
		handler.GetClientBalance(c, pool)
	})

	r.Run(":8050")
}
