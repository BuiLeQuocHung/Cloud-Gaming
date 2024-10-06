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
		newConn     chan *Connection
		binding     *Binding
		freeWorkers []*Connection

		storage *storage.Storage
	}

	Connection struct {
		id    string
		_type ConnectionType
		conn  *websocket.Conn
	}
)

const (
	User   ConnectionType = "user"
	Worker ConnectionType = "worker"
)

func New() *Coordinator {
	return &Coordinator{
		freeWorkers: []*Connection{},
		binding:     NewBinding(),
		newConn:     make(chan *Connection, 10),
		storage:     storage.New(),
	}
}

func (c *Coordinator) Run() {
	go c.listenForNewWebSocketConn()

	http.HandleFunc("/init/worker/ws", c.handleInitWebSocketWorker())
	http.HandleFunc("/init/user/ws", c.handleInitWebSocketUser())
	http.ListenAndServe(":9090", nil)
}

func (c *Coordinator) listenForNewWebSocketConn() {
	log.Debug("listening on websocket")

	for {
		select {
		case conn := <-c.newConn:
			switch conn._type {
			case User:
				go c.userRequestHandler(conn)
			case Worker:
				go c.workerRequestHandler(conn)
			}

		default:
			continue
		}
	}
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

		id := uuid.New()
		newConn := &Connection{
			id:    id.String(),
			_type: Worker,
			conn:  conn,
		}

		c.newConn <- newConn
		c.freeWorkers = append(c.freeWorkers, newConn)
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

		id := uuid.New()
		newConn := &Connection{
			id:    id.String(),
			_type: User,
			conn:  conn,
		}

		if !c.bindUserAndWorker(newConn) {
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
		newConn.conn.WriteJSON(message.ResponseMsg{
			Label:   message.MSG_COOR_HANDSHAKE,
			Payload: payload,
			Error:   nil,
		})

		log.Debug("Send game list to client", zap.Any("games", c.getListGames()))
		c.newConn <- newConn
	}
}
