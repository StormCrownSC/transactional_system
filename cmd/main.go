package main

import (
	"Service/internal/databases"
	"Service/internal/structures"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
	"log"
	"strconv"
	"sync"
)

func main() {
	log.Println("Start app")
	db, err := databases.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer wg.Done()
		r := gin.Default()
		// We specify the path to static files and the route for servicing static files
		r.Use(static.Serve("/", static.LocalFile("./assets", true)))
		r.Run(":8050")
	}()

	// Rsocket
	go func() {
		defer wg.Done()
		err := rsocket.Receive().OnStart(func() {
			log.Println("Server Started")
		}).Acceptor(func(_ context.Context, _ payload.SetupPayload, _ rsocket.CloseableRSocket) (rsocket.RSocket, error) {
			return rsocket.NewAbstractSocket(
				// Request-Response
				rsocket.RequestResponse(func(c payload.Payload) mono.Mono {
					amountStr := c.DataUTF8()
					if amountStr == "" {
						return mono.Error(fmt.Errorf("invalid request format"))
					}
					amount, err := strconv.ParseFloat(amountStr, 64)
					if err == nil {
						return mono.Error(fmt.Errorf("invalid request format"))
					}

					err = databases.WithdrawFunds(db, structures.TransactionRequest{Account: "12345678901234567890", Currency: "AED", Amount: amount})
					if err != nil {
						return mono.Error(err)
					}

					return mono.Just(payload.NewString("SUCCESS", ""))
				}),

				// Request-Stream
				rsocket.RequestStream(func(c payload.Payload) flux.Flux {
					balances, err := databases.GetClientBalances(db, c.DataUTF8())
					if err != nil {
						if err == sql.ErrNoRows {
							return flux.Error(fmt.Errorf("balances not found"))
						} else {
							return flux.Error(fmt.Errorf("with get balances"))
						}
					}

					// Return a successful response
					return flux.Create(func(_ context.Context, s flux.Sink) {
						for _, balance := range balances {
							data, err := json.Marshal(balance)
							if err != nil {
								s.Next(payload.New(nil, nil))
							}
							s.Next(payload.New(data, nil))
						}
						s.Complete()
					})
				}),

				// Request-Channel
				rsocket.RequestChannel(func(c flux.Flux) flux.Flux {
					clients := make(chan string)
					balances := make(chan bool)

					c.DoOnComplete(func() {
						close(clients)
					}).DoOnNext(func(msg payload.Payload) error {
						data := msg.DataUTF8()
						if err != nil {
							log.Fatalln(err)
							return nil
						}
						clients <- data
						return nil
					}).Subscribe(context.Background())

					go func() {
						for client := range clients {
							balanceStruct := databases.DeleteClients(db, client)
							balances <- balanceStruct
						}
						close(balances)
					}()
					return flux.Create(func(_ context.Context, s flux.Sink) {
						for balance := range balances {
							data, err := json.Marshal(balance)
							if err != nil {
								s.Next(payload.New(nil, nil))
							}
							s.Next(payload.New(data, nil))
						}
						s.Complete()
					})
				}),

				// Fire-and-forget
				rsocket.FireAndForget(func(c payload.Payload) {
					amountStr := c.DataUTF8()
					if amountStr != "" {
						amount, err := strconv.ParseFloat(amountStr, 64)
						if err == nil {
							databases.CreateInvoice(db, structures.TransactionRequest{Account: "12345678901234567890", Currency: "AED", Amount: amount})
						}
					}
				}),
			), nil
		}).Transport(rsocket.TCPServer().SetAddr(":8060").Build()).Serve(ctx)

		if err != nil {
			log.Fatalln(err)
		}
	}()
	wg.Wait()
	cancel()
}
