package room

import (
	"time"
	"we-will-call-you/communication"
	"we-will-call-you/game"
)

const (
	pongWait               = 60 * time.Second
	pingPeriod             = (pongWait * 9) / 10
	maxMessageSize         = 512
	connectionLimitPerRoom = 20
)

var H = Hub{
	Broadcast:       make(chan communication.Message),
	Register:        make(chan Subscription),
	Unregister:      make(chan Subscription),
	Rooms:           make(map[string]map[*communication.Connection]bool),
	Domain:          game.C,
	DomainBroadcast: make(chan communication.Message),
}

type Hub struct {
	Rooms           map[string]map[*communication.Connection]bool
	Broadcast       chan communication.Message
	Register        chan Subscription
	Unregister      chan Subscription
	Domain          Domain
	DomainBroadcast chan communication.Message
}

func (h *Hub) Run() {
	h.Domain.Initialize(h.DomainBroadcast)

	for {
		select {
		case s := <-h.Register:
			connections := h.Rooms[s.Room]
			if connections == nil {
				connections = make(map[*communication.Connection]bool)
				h.Rooms[s.Room] = connections
			}

			if len(h.Rooms[s.Room]) < connectionLimitPerRoom {
				h.Rooms[s.Room][s.Conn] = true
				h.Domain.AddConnection(s.Room, s.Conn)
			}
		case s := <-h.Unregister:
			connections := h.Rooms[s.Room]
			if connections != nil {
				if _, ok := connections[s.Conn]; ok {
					delete(connections, s.Conn)
					close(s.Conn.Send)
					h.Domain.RemoveConnection(s.Room, s.Conn)
					if len(connections) == 0 {
						delete(h.Rooms, s.Room)
						h.Domain.DeleteGame(s.Room)
					}
				}
			}
		case m := <-h.Broadcast:
			h.Domain.ProcessCommand(m)
		case m := <-h.DomainBroadcast:
			if m.GetConnection() != nil && m.GetRoomId() != "" {
				h.sendDataToConnection(m.GetRoomId(), m.GetConnection().GetId(), m.GetData())
			} else {
				h.sendDataToConnections(h.Rooms[m.GetRoomId()], m.GetRoomId(), m.GetData())
			}
		}
	}
}

func (h *Hub) sendDataToConnection(roomId, playerId string, output []byte) {
	for c := range h.Rooms[roomId] {
		if c.Id == playerId {
			c.Send <- output
			break
		}
	}
}

func (h *Hub) sendDataToConnections(connections map[*communication.Connection]bool, roomId string, output []byte) {
	for c := range connections {
		select {
		case c.Send <- output:
		default:
			close(c.Send)
			delete(connections, c)
			if len(connections) == 0 {
				delete(h.Rooms, roomId)
				h.Domain.DeleteGame(roomId)
			}
		}
	}
}
