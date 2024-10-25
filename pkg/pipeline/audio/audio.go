package audio

import (
	"cloud_gaming/pkg/encoder"
	"cloud_gaming/pkg/ffmpeg/audio"
	"cloud_gaming/pkg/libretro"
	"cloud_gaming/pkg/log"

	"go.uber.org/zap"
)

type (
	AudioPipeline struct {
		buffer []int16
		maxLen int
		offset int

		channel    int // channel is always 2 for libretro
		sampleRate int

		sendAudioPacket SendAudioPacketFunc

		enc encoder.IAudioEncoder
	}

	AudioPacket struct {
		Buffer   []byte             `json:"buffer"`
		Format   audio.AudioFormat  `json:"format"`
		Codec    encoder.AudioCodec `json:"codec"`
		Duration float64            `json:"duration"` // in milliseconds
	}

	SendAudioPacketFunc func(*AudioPacket)
)

func NewAudioPipeline(sendAudioPacket SendAudioPacketFunc) *AudioPipeline {
	return &AudioPipeline{
		offset:          0,
		channel:         2,
		sendAudioPacket: sendAudioPacket,
	}
}

func (a *AudioPipeline) createEncoder() error {
	var err error
	if a.enc, err = encoder.NewOpusEncoder(a.sampleRate, a.channel); err != nil {
		return err
	}
	return nil
}

func (a *AudioPipeline) SetSystemAudioInfo(systemAVInfo *libretro.SystemAVInfo) {
	sampleRate := systemAVInfo.Timing.SampleRate
	if sampleRate != 48000 {
		sampleRate = 48000
	}

	maxLen := int16(sampleRate * 10 / 1000 * float64(a.channel))
	buffer := make([]int16, maxLen)

	a.buffer = buffer
	a.maxLen = int(maxLen)
	a.sampleRate = int(sampleRate)
}

func (a *AudioPipeline) Process(data []int16, frames int32) {
	if a.enc == nil {
		if err := a.createEncoder(); err != nil {
			log.Error("encoder is nil", zap.Error(err))
			return
		}
	}

	dataOffset := 0
	for dataOffset < len(data) {
		writtenLen := min(a.maxLen-a.offset, len(data)-dataOffset)
		a.write(data, dataOffset, writtenLen)
		dataOffset += writtenLen

		if a.offset == a.maxLen {
			buf, err := a.enc.Encode(a.buffer)
			if err != nil {
				log.Error("audio encoding error: ", zap.Error(err))
				return
			}

			a._sendAudioPacket(&AudioPacket{
				Buffer:   buf,
				Format:   audio.PCM,
				Codec:    encoder.OPUS,
				Duration: 10,
			})
		}

	}
}

func (a *AudioPipeline) write(data []int16, from int, length int) {
	for offset := 0; offset < length; offset++ {
		a.buffer[a.offset+offset] = data[from+offset]
	}

	a.offset += length
}

func (a *AudioPipeline) _sendAudioPacket(packet *AudioPacket) {
	a.sendAudioPacket(packet)
	a.offset = 0
}

func (a *AudioPipeline) Close() error {
	return a.enc.Close()
}
