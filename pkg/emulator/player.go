package emulator

import (
	"cloud_gaming/pkg/libretro"
	"sync"
	"sync/atomic"
)

type (
	Player struct {
		keyboard KeyBoard
		retropad RetroPad
		mouse    Mouse
	}

	KeyBoard struct {
		mu     sync.Mutex
		states [libretro.NO_KB_KEYS]bool
	}

	RetroPad struct {
		mu     sync.Mutex
		states [16]bool // only 16 standard buttons in retropad
	}

	Mouse struct {
		dx, dy atomic.Int32
		state  atomic.Int32
	}
)

const (
	LeftMouse int = iota + 1
	MiddleMouse
	RightMouse
)

func (p *Player) GetKeyState(device uint32, index uint, id uint) int16 {
	switch device {
	case libretro.KEYBOARD:
		if p.keyboard.GetState(id) {
			return 1
		}
		return 0
	case libretro.JOYPAD:
		if p.retropad.GetState(id) {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func (p *Player) SetKeyboardState(id uint, pressed bool) {
	p.keyboard.SetState(id, pressed)
}

func (p *Player) SetMouseState(id uint) {
	p.mouse.SetState(id)
}

func (p *Player) SetMousePos(x, y int) {
	p.mouse.ShiftPosX(int32(x))
	p.mouse.ShiftPosY(int32(y))
}

func (p *Player) SetRetroPadState(id uint, pressed bool) {
	p.retropad.SetState(id, pressed)
}

func (kb *KeyBoard) GetState(id uint) bool {
	return kb.states[id]
}

func (kb *KeyBoard) SetState(id uint, pressed bool) {
	kb.mu.Lock()
	defer kb.mu.Unlock()
	kb.states[id] = pressed

}

func (m *Mouse) GetPosX() int32 {
	return m.dx.Swap(0)
}

func (m *Mouse) GetPosY() int32 {
	return m.dy.Swap(0)
}

func (m *Mouse) ShiftPosX(dx int32) {
	m.dx.Add(dx)
}

func (m *Mouse) ShiftPosY(dy int32) {
	m.dy.Add(dy)
}

func (m *Mouse) GetState(id uint) bool {
	return m.state.Load() == int32(id)
}

func (m *Mouse) SetState(id uint) {
	m.state.Store(int32(id))
}

func (rp *RetroPad) GetState(id uint) bool {
	return rp.states[id]
}

func (rp *RetroPad) SetState(id uint, pressed bool) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.states[id] = pressed
}
