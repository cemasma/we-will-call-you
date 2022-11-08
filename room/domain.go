package room

import "we-will-call-you/communication"

type Domain interface {
	ProcessCommand(message communication.Message)
	AddConnection(roomId string, connection *communication.Connection)
	RemoveConnection(roomId string, connection *communication.Connection)
	Initialize(domainBroadcast chan communication.Message)
	DeleteGame(roomId string)
}
