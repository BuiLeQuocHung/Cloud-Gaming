package worker

import (
	"cloud_gaming/pkg/libretro"
	"unsafe"
)

func (w *Worker) environmentCallback(cmd uint32, data unsafe.Pointer) bool {
	switch cmd {
	case libretro.ENVIRONMENT_SET_PIXEL_FORMAT:
		w.videoPipe.SetPixelFormat(data)
		return true
	case libretro.ENVIRONMENT_SET_ROTATION:
		w.videoPipe.SetRotation(data)
		return true
	case libretro.EnvironmentGetLogInterface:
		w.emulator.BindLogCallback(data, w.emulator.LogCallback)
		return true
	case libretro.EnvironmentGetSystemDirectory:
		libretro.SetString(data, w.emulator.GetSystemDirectory())
		return true
	case libretro.EnvironmentGetVariableUpdate:
		return false
	case libretro.EnvironmentSetKeyboardCallback:
		return false
	}
	return false
}

func (w *Worker) videoRefreshCallback(data unsafe.Pointer, width int32, height int32, pitch int32) {
	arr := unsafe.Slice((*byte)(data), pitch*height)
	w.videoPipe.Process(arr, width, height, pitch)
}

func (w *Worker) audioSampleCallback(l int16, r int16) {
	w.audioPipe.Process([]int16{l, r}, 1)
}

func (w *Worker) audioSampleBatchCallback(buf unsafe.Pointer, frames int32) {
	arr := unsafe.Slice((*int16)(buf), frames*2)
	w.audioPipe.Process(arr, frames)
}
