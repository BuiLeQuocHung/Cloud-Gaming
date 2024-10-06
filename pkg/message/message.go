package message

import "errors"

type (
	RequestMsg struct {
		Label   MsgType `json:"label"`
		Payload []byte  `json:"payload"`
	}

	ResponseMsg struct {
		Label   MsgType `json:"label"`
		Payload []byte  `json:"payload"`
		Error   error   `json:"error"`
	}

	MsgType string
)

const (
	MSG_COOR_HANDSHAKE MsgType = "msg_coor_handshake"
)

const (
	MSG_WEBRTC_INIT          MsgType = "msg_webrtc_init"
	MSG_WEBRTC_OFFER         MsgType = "msg_webrtc_offer"
	MSG_WEBRTC_ANSWER        MsgType = "msg_webrtc_answer"
	MSG_WEBRTC_ICE_CANDIDATE MsgType = "msg_webrtc_ice_candidate"
)

const (
	MSG_START_GAME MsgType = "msg_start_game"
	MSG_STOP_GAME  MsgType = "msg_stop_game"
)

func NewErrorMsg(label MsgType, text string) *ResponseMsg {
	return &ResponseMsg{
		Label: label,
		Error: errors.New(text),
	}
}
