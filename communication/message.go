package communication

type Message struct {
	connection *Connection
	data       []byte
	room       string
}

func NewMessage(c *Connection, data []byte, roomId string) Message {
	return Message{
		connection: c,
		data:       data,
		room:       roomId,
	}
}

func (m Message) GetConnection() *Connection {
	return m.connection
}

func (m Message) GetRoomId() string {
	return m.room
}

func (m Message) GetData() []byte {
	return m.data
}
