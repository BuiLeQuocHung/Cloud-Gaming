package video

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

type (
	ProfileType int
)

const (
	BaseProfile ProfileType = C.FF_PROFILE_H264_BASELINE
	MainProfile ProfileType = C.FF_PROFILE_H264_MAIN // balance
	HighProfile ProfileType = C.FF_PROFILE_H264_HIGH
)
