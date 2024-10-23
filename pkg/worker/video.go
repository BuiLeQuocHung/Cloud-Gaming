package worker

import (
	"cloud_gaming/pkg/pipeline/video"
	"time"

	"github.com/pion/webrtc/v3/pkg/media"
)

func (w *Worker) sendVideoFrame() video.SendVideoFrameFunc {
	return func(vidFrame *video.VideoFrame) {
		w.peerConn.SendVideoFrame(media.Sample{
			Data:     vidFrame.Data,
			Duration: time.Duration(vidFrame.Duration) * time.Millisecond,
			Metadata: map[string]interface{}{
				"Codec":  vidFrame.Codec,
				"Format": vidFrame.Format,
				"Width":  vidFrame.Width,
				"Height": vidFrame.Height,
			},
		})
	}
}
