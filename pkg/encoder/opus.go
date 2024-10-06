package encoder

import (
	"fmt"

	"gopkg.in/hraban/opus.v2"
)

type (
	OpusEncoder struct {
		sampleRate int
		encoder    *opus.Encoder
		channel    int
	}
)

func NewOpusEncoder(sampleRate, channel int) (IAudioEncoder, error) {
	encoder, err := opus.NewEncoder(sampleRate, channel, opus.AppRestrictedLowdelay)
	if err != nil {
		return nil, fmt.Errorf("crete audio encoder error: %w", err)
	}

	encoder.SetDTX(true)
	encoder.SetInBandFEC(true)
	encoder.SetBitrate(96000)
	encoder.SetMaxBandwidth(opus.Fullband)

	return &OpusEncoder{
		encoder:    encoder,
		sampleRate: sampleRate,
		channel:    channel,
	}, nil
}

func (e *OpusEncoder) Encode(pcm []int16) ([]byte, error) {
	buffer := make([]byte, 1024)
	n, err := e.encoder.Encode(pcm, buffer)
	if err != nil {
		return nil, fmt.Errorf("encode error: %w", err)
	}

	return buffer[:n], nil
}

func (e *OpusEncoder) Close() error {
	e.encoder = nil
	return nil
}
