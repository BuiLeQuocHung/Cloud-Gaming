package webrtc

import (
	"cloud_gaming/pkg/message"
	"encoding/json"
	"log"

	_websocket "cloud_gaming/pkg/websocket"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type (
	PeerConnection struct {
		signalConn *_websocket.Conn
		*webrtc.PeerConnection

		vTrack *webrtc.TrackLocalStaticSample
		aTrack *webrtc.TrackLocalStaticSample
	}
)

func NewPeerConnection(signalConn *_websocket.Conn, factory *Factory) (*PeerConnection, error) {
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

		log.Println("worker send ice candidate")

		payload, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			log.Println(" onicecandidate error: %w", err)
			payload = nil
		}

		signalConn.WriteJSON(message.ResponseMsg{
			Label:   message.MSG_WEBRTC_ICE_CANDIDATE,
			Payload: payload,
			Error:   nil,
		})
	})

	peerConn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Println("state change: ", state.String())
	})

	pc := &PeerConnection{
		signalConn:     signalConn,
		PeerConnection: peerConn,
	}

	pc.addAVTrack()

	return pc, nil
}

func (pc *PeerConnection) addAVTrack() error {
	var (
		videoTrack *webrtc.TrackLocalStaticSample
		audioTrack *webrtc.TrackLocalStaticSample
		err        error
	)

	videoTrack, err = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP9},
		"video",
		"video",
	)
	if err != nil {
		log.Println("create video track: ", err)
		return err
	}
	pc.vTrack = videoTrack

	audioTrack, err = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio",
		"audio",
	)
	if err != nil {
		log.Println("create audio track: ", err)
		return err
	}
	pc.aTrack = audioTrack

	_, err = pc.AddTrack(videoTrack)
	log.Println("video track: ", err)
	_, err = pc.AddTrack(audioTrack)
	log.Println("audio track: ", err)
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
