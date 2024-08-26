package socket

import (
	"log"
	"strings"
)

type Hub struct {
	clients    map[string]*Client
	message    chan PrivateMessage
	broadcast  chan BroadcastMessge
	register   chan *Client
	unregister chan string
}

func NewHub() *Hub {
	return &Hub{
		message:    make(chan PrivateMessage),
		broadcast:  make(chan BroadcastMessge),
		register:   make(chan *Client),
		unregister: make(chan string),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			id := client.GetId()
			log.Printf("Registering %s\n", id)
			h.clients[id] = client
		case client := <-h.unregister:
			log.Printf("Unregistering client: %s\n", client)
			if c, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(c.send)
			}
		case msg := <-h.message:
			id := strings.Join([]string{msg.ToUser, msg.SessionId}, ":")
			info, ok := h.clients[id]
			if ok {
				data, err := msg.ToJson()
				if err != nil {
					log.Println(err)
					continue
				}
				select {
				case info.send <- data:
				default:
					close(info.send)
					delete(h.clients, id)
				}
			}
		case msg := <-h.broadcast:
			for id, info := range h.clients {
				if info.SessionId != msg.SessionId {
					continue
				}

				data, err := msg.ToJson()
				if err != nil {
					log.Println(err)
					continue
				}

				select {
				case info.send <- data:
				default:
					close(info.send)
					delete(h.clients, id)
				}
			}
		}
	}
}
