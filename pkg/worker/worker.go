package worker

import (
	"cloud_gaming/pkg/emulator"
	"cloud_gaming/pkg/message"
	"cloud_gaming/pkg/pipeline"
	"cloud_gaming/pkg/storage"
	_webrtc "cloud_gaming/pkg/webrtc"
	_websocket "cloud_gaming/pkg/websocket"

	"encoding/json"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type (
	Worker struct {
		webrtcFactory   *_webrtc.Factory
		coordinatorConn *_websocket.Conn
		peerConn        *_webrtc.PeerConnection
		emulator        *emulator.Emulator
		videoPipe       *pipeline.VideoPipeline
		audioPipe       *pipeline.AudioPipeline
		storage         *storage.Storage
	}
)

func New() (*Worker, error) {
	var err error
	w := &Worker{
		emulator: emulator.New(),
		storage:  storage.New(),
	}

	w.videoPipe, err = pipeline.NewVideoPipeline(w.sendVideoFrame())
	if err != nil {
		return nil, err
	}
	w.audioPipe = pipeline.NewAudioPipeline(w.sendAudioPacket())
	return w, nil
}

func (w *Worker) Run() {
	w.initWebSocketConnToCoordinator()
	w.InitWebrtcFactory()

	go w.requestHandler()
}

func (w *Worker) initWebSocketConnToCoordinator() {
	u := url.URL{
		Scheme: "ws",
		Host:   "coordinator:9090",
		Path:   "/init/worker/ws",
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}

	w.coordinatorConn = _websocket.New(c)
}

func (w *Worker) requestHandler() {
	var err error

	conn := w.coordinatorConn
	defer conn.Close()

	w.peerConn, err = _webrtc.NewPeerConnection(conn, w.webrtcFactory)
	if err != nil {
		log.Println("create pc failed: ", err)
		return
	}

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			log.Println("worker web socket closed")
			break
		}

		msg := &message.RequestMsg{}
		if err := json.Unmarshal(data, msg); err != nil {
			log.Println("unmarshal error: ", err)
			w.sendError("unknown", "unmarshal request message failed")
			break
		}

		log.Println("message: ", msg)
		log.Println("message payload: ", string(msg.Payload))

		switch msg.Label {
		case message.MSG_WEBRTC_INIT:
			err = w.peerConn.AddInputChannel(
				w.handleKeyboardChannel(),
				w.handleMouseChannel(),
			)
			if err != nil {
				log.Println("create input channel failed: ", err)
				w.sendError(msg.Label, "create input channel failed")
			}

			localSD, err := w.peerConn.CreateOffer(nil)
			if err != nil {
				log.Println("localSD error: ", err)
				w.sendError(msg.Label, "create local session description failed")
			}
			log.Println("localSD: ", localSD)

			err = w.peerConn.SetLocalDescription(localSD)
			if err != nil {
				w.sendError(msg.Label, "set local description failed")
			}

			// err = w.peerConn.AddAVTrack()
			// if err != nil {
			// 	log.Println("create track failed: ", err)
			// 	w.sendError(msg.Label, "create video/audio track failed")
			// }

			payload, err := json.Marshal(localSD)
			if err != nil {
				w.sendError(msg.Label, "marshal local session description failed")
			}

			log.Println("payload: ", payload)
			res := &message.ResponseMsg{
				Label:   message.MSG_WEBRTC_OFFER,
				Payload: payload,
				Error:   nil,
			}

			w.coordinatorConn.WriteJSON(res)
		case message.MSG_WEBRTC_ANSWER:
			var remoteSD = &webrtc.SessionDescription{}
			err := json.Unmarshal(msg.Payload, remoteSD)
			if err != nil {
				w.sendError(msg.Label, "unmarshal session description offer failed")
			}
			w.peerConn.SetRemoteDescription(*remoteSD)

		case message.MSG_WEBRTC_ICE_CANDIDATE:
			log.Println("here  abc")
			var candidate = &webrtc.ICECandidateInit{}
			err := json.Unmarshal(msg.Payload, candidate)
			if err != nil {
				w.sendError(msg.Label, "unmarshal ice candidate failed")
			}
			w.peerConn.AddICECandidate(*candidate)

		case message.MSG_START_GAME:
			r := &StartGameRequest{}
			err = json.Unmarshal(msg.Payload, r)
			log.Println("start game: ", r)
			log.Println(err)
			if err != nil {
				w.sendError(msg.Label, "unmarshal game request failed")
			}

			err = w.startEmulator(r)
			if err != nil {
				w.sendError(msg.Label, "start emulator failed")
			}

		case message.MSG_STOP_GAME:
			w.stopEmulator()
		}

	}
}

func (w *Worker) sendError(label message.MsgType, text string) {
	resp := message.NewErrorMsg(label, text)
	w.coordinatorConn.WriteJSON(resp)
}

func (w *Worker) InitWebrtcFactory() {
	factory, err := _webrtc.NewFactory()
	if err != nil {
		log.Fatal(err)
	}

	w.webrtcFactory = factory
}
