package pipeline

import (
	"cloud_gaming/pkg/encoder"
	"cloud_gaming/pkg/format"
	"cloud_gaming/pkg/libretro"
	"cloud_gaming/pkg/log"
	"sync"

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
		mu  sync.Mutex
	}

	AudioPacket struct {
		Buffer   []byte             `json:"buffer"`
		Format   format.AudioFormat `json:"format"`
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

func (a *AudioPipeline) init() {
	opusEncoder, err := encoder.NewOpusEncoder(a.sampleRate, a.channel)
	if err != nil {
		log.Error("opus init failed", zap.Error(err))
		return
	}
	a.enc = opusEncoder
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

	a.init()
}

func (a *AudioPipeline) Process(data []int16, frames int32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.enc == nil {
		log.Error("encoder is nil")
		return
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

			a.SendAudioPacket(&AudioPacket{
				Buffer:   buf,
				Format:   format.PCM,
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

func (a *AudioPipeline) SendAudioPacket(packet *AudioPacket) {
	a.sendAudioPacket(packet)
	a.offset = 0
}

func (a *AudioPipeline) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.buffer = nil
	a.maxLen = 0
	a.offset = 0
	a.sampleRate = 0

	a.enc.Close()
	a.enc = nil

	return nil
}
