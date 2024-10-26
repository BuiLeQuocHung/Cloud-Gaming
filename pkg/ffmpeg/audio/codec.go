package audio

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

type (
	AudioCodec int
)

const (
	NoCodec AudioCodec = C.AV_CODEC_ID_NONE
	OPUS    AudioCodec = C.AV_CODEC_ID_OPUS
)
