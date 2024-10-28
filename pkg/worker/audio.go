package worker

import (
	"cloud_gaming/pkg/pipeline/audio"
	"time"

	"github.com/pion/webrtc/v3/pkg/media"
)

func (w *Worker) sendAudioPacket(audioPacket *audio.AudioPacket) {
	w.peerConn.SendAudioFrame(media.Sample{
		Data:     audioPacket.Buffer,
		Duration: time.Duration(audioPacket.Duration) * time.Millisecond,
		Metadata: map[string]interface{}{
			"Codec":  audioPacket.Codec,
			"Format": audioPacket.Format,
		},
	})
}
