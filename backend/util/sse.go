package util

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type (
	NotificationEvent struct {
		EventName string
		Payload   interface{}
	}

	NotifierChan chan NotificationEvent

	Broker struct {
		Notifier       NotifierChan
		newClients     chan NotifierChan
		closingClients chan NotifierChan
		clients        map[NotifierChan]struct{}
	}
)

var broker *Broker

const patience time.Duration = time.Second * 1

func init() {
	broker = &Broker{
		Notifier:       make(NotifierChan, 1),
		newClients:     make(chan NotifierChan),
		closingClients: make(chan NotifierChan),
		clients:        make(map[NotifierChan]struct{}),
	}
	go broker.listen()
}

func HandleEvents(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	var eventName string
	topic := req.URL.Query()["t"]
	if len(topic) > 0 {
		eventName = topic[0]
	}
	if !Config.Production {
		log.Printf("Requested topic: %s", eventName)
	}
	header := rw.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("Access-Control-Allow-Origin", "*")

	messageChan := make(NotifierChan)
	broker.newClients <- messageChan

	defer func() {
		broker.closingClients <- messageChan
	}()

	notify := req.Context().Done()

	go func() {
		<-notify
		broker.closingClients <- messageChan
	}()

	for {
		event := <-messageChan
		if eventName == event.EventName || event.EventName == "" {
			fmt.Fprintf(rw, "data: %s\n\n", event.Payload)
			flusher.Flush()
		}
	}
}

func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:
			broker.clients[s] = struct{}{}
			if !Config.Production {
				log.Printf("Client added. %d registered clients", len(broker.clients))
			}
		case s := <-broker.closingClients:
			delete(broker.clients, s)
			if !Config.Production {
				log.Printf("Removed client. %d registered clients", len(broker.clients))
			}
		case event := <-broker.Notifier:
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- event:
				case <-time.After(patience):
					log.Printf("Skipping client %s", event.EventName)
				}
			}
		}
	}
}

func PublishTopic(topic string, data string) {
	broker.Notifier <- NotificationEvent{
		EventName: topic,
		Payload:   data,
	}
}
