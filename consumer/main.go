package main

import (
	"fmt"
	"log"
	"os"
	"time"

	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

func main() {
	<-time.After(10 * time.Second)

	URL := "nats://nats:4222"
	clusterID := "test-cluster"
	clientID := "consummer"

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Streaming Example Subscriber")}
	// Use UserCredentials

	// Connect to NATS
	nc, err := nats.Connect(URL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	fmt.Println("connected to nats")
	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}
	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", URL, clusterID, clientID)

	// Simple Async Subscriber
	sub, _ := sc.Subscribe("events", func(m *stan.Msg) {
		fmt.Printf("EVENT RECIEVED: %s\n", string(m.Data))
	})
	// Simple Async Subscriber
	subError, _ := sc.Subscribe("error", func(m *stan.Msg) {
		fmt.Printf("EVENT RECIEVED: %s\n", string(m.Data))
	})

	// Unsubscribe
	defer sub.Unsubscribe()
	defer subError.Unsubscribe()

	// Close connection
	defer sc.Close()

	signalChan := make(chan os.Signal, 1)
	<-signalChan
}
