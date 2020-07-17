package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	socketio "github.com/googollee/go-socket.io"
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

var eventMsg = `{"type": "%s", "msg": "%s"}`

func main() {
	<-time.After(10 * time.Second)

	URL := "nats://nats:4222"
	clusterID := "test-cluster"
	clientID := "gateway"

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

	// Close connection
	defer sc.Close()

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		connString := "connected:" + s.ID()
		sc.Publish("events", []byte(connString))
		return nil
	})
	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		noticeString := "notice:" + msg
		sc.Publish("events", []byte(fmt.Sprintf(eventMsg, "notice", noticeString)))
		s.Emit("reply", "have "+msg)
	})
	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		sc.Publish("events", []byte(fmt.Sprintf(eventMsg, "chat", msg)))
		return "recv " + msg
	})
	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		sc.Publish("events", []byte(fmt.Sprintf(eventMsg, "conn", "bye "+last)))
		s.Close()
		return last
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		sc.Publish("error", []byte(fmt.Sprintf(eventMsg, "error", e.Error())))
	})
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		sc.Publish("events", []byte(fmt.Sprintf(eventMsg, "conn", "disconnected "+reason)))
	})
	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
