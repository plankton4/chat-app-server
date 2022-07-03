package network

import "log"

type HubID uint32

const (
	HubGeneralChat    HubID = 1
	HubMemesChat      HubID = 2
	HubMoviesChat     HubID = 3
	HubVideoGamesChat HubID = 4
)

var testHubs = []HubID{
	HubGeneralChat,
	HubMemesChat,
	HubMoviesChat,
	HubVideoGamesChat,
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	hubID HubID
}

func SetupHubs() {
	for _, hubID := range testHubs {
		newHub := NewHub(hubID)
		ActiveHubs[hubID] = newHub
		go newHub.Run()
	}
}

func NewHub(id HubID) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		hubID:      id,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
			}
		case message := <-h.broadcast:
			log.Println("Broadcast, clients: ")

			for client := range h.clients {
				select {
				case client.send <- message:
					//
					log.Println("ClientID ", client.clientID)
				default:
					delete(h.clients, client)
				}
			}
		}
	}
}
