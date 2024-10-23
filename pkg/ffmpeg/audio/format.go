package audio

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

type (
	AudioFormat int
)

const (
	PCM AudioFormat = C.AV_SAMPLE_FMT_S16
)
