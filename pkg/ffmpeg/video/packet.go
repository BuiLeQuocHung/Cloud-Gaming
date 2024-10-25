package video

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"
import "unsafe"

type (
	Packet = C.struct_AVPacket
)

func NewPacket() *Packet {
	return C.av_packet_alloc()
}

func (p *Packet) Close() {
	C.av_packet_free(&p)
}

func (p *Packet) GetData() unsafe.Pointer {
	return unsafe.Pointer(p.data)
}

func (p *Packet) GetSize() int {
	return int(p.size)
}
