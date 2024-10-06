package emulator

import (
	"cloud_gaming/pkg/libretro"
	"errors"
	"log"
	"os"
	"unsafe"
)

/*
#include <stdlib.h>
*/
import "C"

const (
	MAX_PLAYERS = 2
)

type (
	Emulator struct {
		core    *libretro.Core
		state   EmulatorState
		players [MAX_PLAYERS]Player

		systemDir string
	}

	EmulatorState int
)

const (
	Ready EmulatorState = iota
	Running
	Deinitializing
	LastState // used to count number of states
)

func New() *Emulator {
	return &Emulator{
		core:    nil,
		state:   Ready,
		players: [MAX_PLAYERS]Player{},

		systemDir: "./libretro/system",
	}
}

func KbToRetroPad(btnID uint) (uint, bool) {
	_map := map[uint]uint{
		13:  uint(libretro.DeviceIDJoypadStart),  // Enter
		304: uint(libretro.DeviceIDJoypadSelect), // LShift
		303: uint(libretro.DeviceIDJoypadSelect), // RShift
		120: uint(libretro.DeviceIDJoypadA),      // X
		122: uint(libretro.DeviceIDJoypadB),      // Z
		119: uint(libretro.DeviceIDJoypadR),      // W
		113: uint(libretro.DeviceIDJoypadL),      // Q
		273: uint(libretro.DeviceIDJoypadUp),     // UP
		274: uint(libretro.DeviceIDJoypadDown),   // DOWN
		276: uint(libretro.DeviceIDJoypadLeft),   // LEFT
		275: uint(libretro.DeviceIDJoypadRight),  // RIGHT
	}

	if retroID, ok := _map[btnID]; ok {
		return retroID, true
	}
	return 0, false
}

func (e *Emulator) LoadCore(
	sofile string,
	environmentCallback libretro.EnvironmentFunc,
	videoRefreshCallback libretro.VideoRefreshFunc,
	audioSampleCallback libretro.AudioSampleFunc,
	audioSampleBatchCallback libretro.AudioSampleBatchFunc,
) error {
	core, err := libretro.Load(sofile)
	if err != nil {
		return err
	}

	e.core = core
	e.core.SetEnvironment(environmentCallback)
	e.core.SetVideoRefresh(videoRefreshCallback)
	e.core.SetAudioSample(audioSampleCallback)
	e.core.SetAudioSampleBatch(audioSampleBatchCallback)
	e.core.SetInputState(e.inputStateCallback)
	e.core.SetInputPoll(e.inputPollCallback)

	// e.core.SetAudioCallback(nil)
	// e.core.SetFrameTimeCallback(nil)
	// e.core.SetDiskControlCallback(nil)
	e.core.MemoryMap = nil

	return nil
}

func (e *Emulator) Init() {
	e.core.Init()
}

func (e *Emulator) DeInit() {
	e.core.Deinit()
}

func (e *Emulator) LoadGame(path string) error {
	log.Println("core: ", e.core)

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	cData := C.CBytes(data)
	defer C.free(cData)

	gameInfo := libretro.GameInfo{
		Path: path,
		Size: int64(len(data)),
		Data: unsafe.Pointer(cData),
	}

	log.Printf("Game Path: %s\n", gameInfo.Path)
	log.Printf("Game Size: %d\n", gameInfo.Size)
	log.Printf("Game Data: %v\n", gameInfo.Data)

	isSuccess := e.core.LoadGame(gameInfo)
	if !isSuccess {
		return errors.New("load game failed")
	}
	return nil
}

func (e *Emulator) UnloadGame() {
	e.core.UnloadGame()
}

// Run runs the game for one video frame.
func (e *Emulator) Run() {
	// log.Println("run core: ", e.core)
	e.core.Run()
}

// GetSystemAVInfo returns information about
// system audio/video timings and geometry.
// Can be called only after retro_load_game() has successfully completed.
func (e *Emulator) GetSystemAVInfo() libretro.SystemAVInfo {
	log.Println("Load AV info")
	res := e.core.GetSystemAVInfo()
	return res
}

func (e *Emulator) LogCallback(level uint32, msg string) {
	var logLevels = map[uint32]string{
		libretro.LogLevelDebug: "DEBUG",
		libretro.LogLevelInfo:  "INFO",
		libretro.LogLevelWarn:  "WARN",
		libretro.LogLevelError: "ERROR",
		libretro.LogLevelDummy: "DUMMY",
	}

	log.Printf("[%s]: %s", logLevels[level], msg)
}

func (e *Emulator) BindLogCallback(data unsafe.Pointer, logFunc func(uint32, string)) {
	e.core.BindLogCallback(data, logFunc)
}

func (e *Emulator) GetSystemDirectory() string {
	return e.systemDir
}

func (e *Emulator) IsRunning() bool {
	return e.state == Running
}

func (e *Emulator) SetState(newState EmulatorState) {
	e.state = newState
}

func (e *Emulator) IsReady() bool {
	return e.state == Ready
}

func (e *Emulator) GetState(newState EmulatorState) EmulatorState {
	return e.state
}

func (e *Emulator) inputPollCallback() {}

func (e *Emulator) inputStateCallback(port uint, device uint32, index uint, id uint) int16 {
	if port >= MAX_PLAYERS {
		return 0
	}

	return e.players[port].GetKeyState(device, index, id)
}

func (e *Emulator) SetKeyboardState(port uint, id uint, pressed bool) {
	if port >= MAX_PLAYERS {
		return
	}

	e.players[port].SetKeyboardState(id, pressed)
	if retroID, found := KbToRetroPad(id); found {
		e.players[port].SetRetroPadState(retroID, pressed)
	}
}

func (e *Emulator) SetMouseState(port uint, id uint) {
	if port >= MAX_PLAYERS {
		return
	}

	e.players[port].SetMouseState(id)
}

func (e *Emulator) SetMousePos(port uint, x, y int) {
	if port >= MAX_PLAYERS {
		return
	}

	e.players[port].SetMousePos(x, y)
}
