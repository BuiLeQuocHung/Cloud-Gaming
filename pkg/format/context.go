package format

/*
#cgo pkg-config: libswscale
#include <libswscale/swscale.h>
*/
import "C"

type (
	CtxKey struct {
		from_width  int
		from_height int
		from        VideoFormat

		width  int
		height int
		to     VideoFormat

		scalingAlgo int
	}

	Context struct {
		ctxMap map[CtxKey]*C.struct_SwsContext
	}
)

var (
	FmtCtx *Context = NewFormatCtx()
)

func NewFormatCtx() *Context {
	return &Context{
		ctxMap: make(map[CtxKey]*C.struct_SwsContext),
	}
}

func (c *Context) Get(k CtxKey) *C.struct_SwsContext {
	if ctx, ok := c.ctxMap[k]; ok {
		return ctx
	}

	c.ctxMap[k] = C.sws_getContext(
		C.int(k.from_width), C.int(k.from_height), int32(k.from),
		C.int(k.width), C.int(k.height), int32(k.to),
		C.int(k.scalingAlgo), nil, nil, nil)

	return c.ctxMap[k]
}

func (c *Context) Reset() {
	for _, v := range c.ctxMap {
		C.sws_freeContext(v)
	}

	c.ctxMap = make(map[CtxKey]*C.struct_SwsContext)
}
