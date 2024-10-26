package webrtc

import (
	"cloud_gaming/pkg/log"
	"cloud_gaming/pkg/message"
	"encoding/json"

	_websocket "cloud_gaming/pkg/websocket"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"go.uber.org/zap"
)

type (
	PeerConnection struct {
		signalConn *_websocket.Conn
		*webrtc.PeerConnection

		vTrack *webrtc.TrackLocalStaticSample
		aTrack *webrtc.TrackLocalStaticSample
	}
)

func NewPeerConnection(signalConn *_websocket.Conn, factory *Factory,
	callbackWebRTCDisconnectedFunc func(),
) (*PeerConnection, error) {
	peerConn, err := factory.NewPeerConnection(
		webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	peerConn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		payload, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			log.Error("onicecandidate error", zap.Error(err))
			payload = nil
		}

		signalConn.WriteJSON(message.ResponseMsg{
			Label:   message.MSG_WEBRTC_ICE_CANDIDATE,
			Payload: payload,
			Error:   nil,
		})
	})

	peerConn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Debug("state change", zap.String("state", state.String()))

		if state == webrtc.PeerConnectionStateDisconnected {
			log.Debug("webrtc disconnected")
			callbackWebRTCDisconnectedFunc()
		}
	})

	pc := &PeerConnection{
		signalConn:     signalConn,
		PeerConnection: peerConn,
	}

	if err := pc.addAVTrack(); err != nil {
		return nil, err
	}
	return pc, nil
}

func (pc *PeerConnection) addAVTrack() error {
	var (
		videoTrack *webrtc.TrackLocalStaticSample
		audioTrack *webrtc.TrackLocalStaticSample
		err        error
	)

	videoTrack, err = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
		"video",
		"video",
	)
	if err != nil {
		log.Error("create video track: ", zap.Error(err))
		return err
	}
	pc.vTrack = videoTrack

	audioTrack, err = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio",
		"audio",
	)
	if err != nil {
		log.Error("create audio track: ", zap.Error(err))
		return err
	}
	pc.aTrack = audioTrack

	pc.AddTrack(videoTrack)
	pc.AddTrack(audioTrack)
	return nil
}

func (pc *PeerConnection) AddInputChannel(keyboardbCallback, mouseCallback func(msg webrtc.DataChannelMessage)) error {
	kbChannel, err := pc.CreateDataChannel("keyboard", nil)
	if err != nil {
		return err
	}

	mouseChannel, err := pc.CreateDataChannel("mouse", nil)
	if err != nil {
		return err
	}

	kbChannel.OnMessage(keyboardbCallback)
	mouseChannel.OnMessage(mouseCallback)
	return nil
}

func (pc *PeerConnection) SendVideoFrame(sample media.Sample) error {
	return pc.vTrack.WriteSample(sample)
}

func (pc *PeerConnection) SendAudioFrame(sample media.Sample) error {
	return pc.aTrack.WriteSample(sample)
}

func (pc *PeerConnection) Close() {
	pc.PeerConnection.Close()
}
