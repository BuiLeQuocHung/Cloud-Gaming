package utils

/*
#cgo pkg-config: libavutil
#include <libavutil/error.h>
*/
import "C"
import "errors"

func CErrorToString(errnum int) error {
	if errnum == 0 {
		return nil
	}

	var errBuf [128]C.char
	C.av_strerror(C.int(errnum), &errBuf[0], C.size_t(len(errBuf)))
	return errors.New(C.GoString(&errBuf[0]))
}
