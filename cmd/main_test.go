package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
)

func connect(t *testing.T) rsocket.Client {
	// Connect to server
	cli, err := rsocket.Connect().
		Transport(rsocket.TCPClient().
			SetHostAndPort("localhost", 12000).
			Build()).
		Start(context.Background())
	if err != nil {
		t.Error(err)
	}
	return cli
}

func TestRequestResponse(t *testing.T) {
	cli := connect(t)
	defer cli.Close()
	amountStr := "100.50" // Replace with your mock amount

	// Create a payload with the mock amount
	requestPayload := payload.NewString(amountStr, "")

	// Call the function being tested (RequestResponse)
	response := cli.RequestResponse(requestPayload)

	response.
		DoOnSuccess(func(input payload.Payload) error {
			return nil
		}).
		DoOnError(func(e error) {
			fmt.Printf("Возникла ошибка (%s)", e)
		}).
		Subscribe(context.Background())
}

func TestRequestStream(t *testing.T) {
	cli := connect(t)
	defer cli.Close()
	clientID := "123"

	requestPayload := payload.NewString(clientID, "")

	response := cli.RequestStream(requestPayload)

	var receivedPayloads []payload.Payload

	response.
		DoOnNext(func(input payload.Payload) error {
			receivedPayloads = append(receivedPayloads, input)
			return nil
		}).
		Subscribe(context.Background())
}

func createMockFlux(mockData []string) flux.Flux {
	return flux.Create(func(ctx context.Context, sink flux.Sink) {
		for _, data := range mockData {
			sink.Next(payload.NewString(data, ""))
		}
		sink.Complete()
	})
}

func TestRequestChannel(t *testing.T) {
	cli := connect(t)
	defer cli.Close()
	mockClients := []string{"client1", "client2"} // Replace with your mock client data
	mockFlux := createMockFlux(mockClients)

	// Call the function being tested (RequestChannel)
	response := cli.RequestChannel(mockFlux)

	var receivedPayloads []payload.Payload

	response.
		DoOnNext(func(input payload.Payload) error {
			receivedPayloads = append(receivedPayloads, input)
			return nil
		}).
		Subscribe(context.Background())
}

func TestFireAndForget(t *testing.T) {
	cli := connect(t)
	defer cli.Close()

	mockAmountStr := "100.50"
	payloadToSend := payload.NewString(mockAmountStr, "")

	cli.FireAndForget(payloadToSend)

}
