package coordinator

import (
	"log"

	"github.com/gorilla/websocket"
)

func (c *Coordinator) workerRequestHandler(connection *Connection) {
	senderId := connection.id
	conn := connection.conn

	log.Println("listening to worker")

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			log.Println("worker web socket closed", err)
			break
		}

		pair := c.binding.GetPair(senderId)
		if pair == nil {
			conn.Close()
			break
		}

		receiverConn := pair.user.conn
		receiverConn.WriteMessage(websocket.TextMessage, data)
	}
}
