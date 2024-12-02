package coordinator

import (
	"cloud_gaming/pkg/log"
	"cloud_gaming/pkg/message"
	"cloud_gaming/pkg/storage"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type (
	ConnectionType string
)

type (
	Coordinator struct {
		binding     *Binding
		freeWorkers chan *Connection

		storage *storage.Storage
	}

	Connection struct {
		id   string
		conn *websocket.Conn
	}
)

const (
	User   ConnectionType = "user"
	Worker ConnectionType = "worker"
)

func New() *Coordinator {
	return &Coordinator{
		freeWorkers: make(chan *Connection, 1000),
		binding:     NewBinding(),
		storage:     storage.New(),
	}
}

func (c *Coordinator) Run() {
	http.HandleFunc("/init/worker/ws", c.handleInitWebSocketWorker())
	http.HandleFunc("/init/user/ws", c.handleInitWebSocketUser())
	http.ListenAndServe(":9090", nil)
}

func (c *Coordinator) handleInitWebSocketWorker() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		log.Debug("worker opens coonection", zap.Error(err))

		if err != nil {
			return
		}

		workerConn := &Connection{
			id:   uuid.New().String(),
			conn: conn,
		}
		c.freeWorkers <- workerConn
		go c.workerRequestHandler(workerConn)
	}
}

func (c *Coordinator) handleInitWebSocketUser() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		userConn := &Connection{
			id:   uuid.New().String(),
			conn: conn,
		}

		if !c.bindUserAndWorker(userConn) {
			log.Error("cannot bind worker")
			conn.Close()
			return
		}

		payload, err := json.Marshal(c.getListGames())
		if err != nil {
			log.Error("cannot get  list game")
			return
		}

		// send list games to client
		userConn.conn.WriteJSON(message.ResponseMsg{
			Label:   message.MSG_COOR_HANDSHAKE,
			Payload: payload,
			Error:   nil,
		})

		log.Debug("Send game list to client", zap.Any("games", c.getListGames()))
		go c.userRequestHandler(userConn)
	}
}

func (c *Coordinator) bindUserAndWorker(userConn *Connection) bool {
	select {
	case workerConn := <-c.freeWorkers:
		c.binding.Bind(userConn, workerConn)
	default:
		return false
	}

	return true
}
