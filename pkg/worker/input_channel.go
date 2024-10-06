package worker

import (
	"encoding/json"

	"github.com/pion/webrtc/v3"
)

type (
	keyboardData struct {
		User        uint          `json:"user"`
		ButtonState []buttonState `json:"button_state"`
	}

	buttonState struct {
		Button  uint `json:"button"`
		Pressed bool `json:"pressed"`
	}

	// currently not used
	mouseData struct {
		User   uint `json:"user"`
		Button int  `json:"button"` // left, middle, right
		PosX   int  `json:"pos_x"`
		PosY   int  `json:"pox_y"`
	}
)

func (w *Worker) handleKeyboardChannel() func(msg webrtc.DataChannelMessage) {
	return func(msg webrtc.DataChannelMessage) {
		var kb = &keyboardData{}
		err := json.Unmarshal(msg.Data, kb)
		if err != nil {
			return
		}

		user := kb.User
		for _, bt := range kb.ButtonState {
			w.emulator.SetKeyboardState(user, bt.Button, bt.Pressed)
		}
	}
}

func (w *Worker) handleMouseChannel() func(msg webrtc.DataChannelMessage) {
	return func(msg webrtc.DataChannelMessage) {
		var mouse = &mouseData{}
		err := json.Unmarshal(msg.Data, mouse)
		if err != nil {
			return
		}

		user := mouse.User
		w.emulator.SetMouseState(user, uint(mouse.Button))
		w.emulator.SetMousePos(user, mouse.PosX, mouse.PosY)
	}
}
