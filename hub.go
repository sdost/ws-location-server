package main

import (
	"encoding/json"
	"log"
)

// Location object which holds latitude and longitude
type Location struct {
	Longitude float64 `json:"lat"`
	Latitude  float64 `json:"lon"`
	Altitude  float64 `json:"alt"`
	Heading   float64 `json:"heading"`
}

// Message message is the object passed over the web socket
type Message struct {
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Message  string   `json:"message"`
	Position Location `json:"position"`
}

// Hub hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			log.Printf("Client connected: %+v\n", client)
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Printf("Client disconnected: %+v\n", client)
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			log.Printf("Client broadcast message: %+v\n", message)
			bytes, err := json.Marshal(message)
			if err != nil {
				log.Fatal(err)
			}
			for client := range h.clients {
				select {
				case client.send <- bytes:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
