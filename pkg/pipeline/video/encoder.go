package video

import (
	"cloud_gaming/pkg/encoder"
	"cloud_gaming/pkg/ffmpeg/video"
	"errors"
)

type (
	Encoder struct {
		encoder.IVideoEncoder
	}
)

func NewEncoder(codec video.VideoCodec, width, height int, pixFormat video.PixelFormat, fps float64) (*Encoder, error) {
	e := &Encoder{}

	enc, err := e.createEncoder(codec, width, height, pixFormat, fps)
	if err != nil {
		return nil, err
	}

	return &Encoder{
		IVideoEncoder: enc,
	}, nil
}

func (e *Encoder) createEncoder(codec video.VideoCodec, width, height int,
	pixFormat video.PixelFormat, fps float64) (encoder.IVideoEncoder, error) {

	switch codec {
	case video.H264:
		return e.createH264Encoder(width, height, pixFormat, fps)
	case video.VP9:
		return e.createVP9Encoder(width, height, pixFormat, fps)
	default:
		return nil, errors.New("codec not supported")
	}
}

func (e *Encoder) createH264Encoder(width, height int,
	pixFormat video.PixelFormat, fps float64) (encoder.IVideoEncoder, error) {

	var err error
	var enc encoder.IVideoEncoder

	enc, err = encoder.NewH264Encoder(width, height, int(fps), pixFormat)
	if err == nil {
		return enc, nil
	}

	return nil, err
}

func (e *Encoder) createVP9Encoder(width, height int,
	pixFormat video.PixelFormat, fps float64) (encoder.IVideoEncoder, error) {

	var err error
	var enc encoder.IVideoEncoder

	enc, err = encoder.NewVP9Encoder(width, height, int(fps), pixFormat)
	if err == nil {
		return enc, nil
	}
	return nil, err
}
