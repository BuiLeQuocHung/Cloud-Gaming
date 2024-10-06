package webrtc

import (
	"net"

	"github.com/pion/webrtc/v3"
)

type (
	Factory struct {
		*webrtc.API
	}
)

func NewFactory() (*Factory, error) {
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 9000})
	if err != nil {
		return nil, err
	}

	s := &webrtc.SettingEngine{}
	s.SetICEUDPMux(webrtc.NewICEUDPMux(nil, udpConn))

	m := &webrtc.MediaEngine{}
	m.RegisterDefaultCodecs()

	return &Factory{
		API: webrtc.NewAPI(webrtc.WithSettingEngine(*s), webrtc.WithMediaEngine(m)),
	}, nil
}
