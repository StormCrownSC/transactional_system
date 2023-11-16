package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
)

func init() {
	go main()
}

func connect(t *testing.T) (rsocket.Client, error) {
	// Connect to server
	return rsocket.Connect().
		Transport(rsocket.TCPClient().
			SetHostAndPort("localhost", 8060).
			Build()).
		Start(context.Background())
}

func TestRequestResponse(t *testing.T) {
	client, err := connect(t)
	if err != nil {
		panic("error with connect")
	}
	defer client.Close()
	amountStr := "100.50" // Replace with your mock amount

	// Create a payload with the mock amount
	requestPayload := payload.NewString(amountStr, "")

	// Call the function being tested (RequestResponse)
	response := client.RequestResponse(requestPayload)

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
	client, err := connect(t)
	if err != nil {
		t.Error(err)
	}
	defer client.Close()
	clientID := "123"

	requestPayload := payload.NewString(clientID, "")

	response := client.RequestStream(requestPayload)

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
	client, err := connect(t)
	if err != nil {
		t.Error(err)
	}
	defer client.Close()
	mockClients := []string{"client1", "client2"} // Replace with your mock client data
	mockFlux := createMockFlux(mockClients)

	// Call the function being tested (RequestChannel)
	response := client.RequestChannel(mockFlux)

	var receivedPayloads []payload.Payload

	response.
		DoOnNext(func(input payload.Payload) error {
			receivedPayloads = append(receivedPayloads, input)
			return nil
		}).
		Subscribe(context.Background())
}

func TestFireAndForget(t *testing.T) {
	client, err := connect(t)
	if err != nil {
		t.Error(err)
	}
	defer client.Close()

	mockAmountStr := "100.50"
	payloadToSend := payload.NewString(mockAmountStr, "")

	client.FireAndForget(payloadToSend)

}