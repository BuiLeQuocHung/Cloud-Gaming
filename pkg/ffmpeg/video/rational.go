package video

/*
#cgo pkg-config: libavutil
#include <libavutil/rational.h>
*/
import "C"

type (
	Rational = C.AVRational
)

func NewRational(num, den int) *Rational {
	return &Rational{
		num: C.int(num),
		den: C.int(den),
	}
}
