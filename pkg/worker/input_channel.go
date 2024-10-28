package worker

import (
	"cloud_gaming/pkg/log"
	"encoding/json"

	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
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

func (w *Worker) handleKeyboardChannel(msg webrtc.DataChannelMessage) {
	var kb = &keyboardData{}
	err := json.Unmarshal(msg.Data, kb)
	if err != nil {
		log.Error("unmarshal keyboard data failed", zap.Error(err))
		return
	}

	user := kb.User
	for _, bt := range kb.ButtonState {
		w.emulator.SetKeyboardState(user, bt.Button, bt.Pressed)
	}
}

func (w *Worker) handleMouseChannel(msg webrtc.DataChannelMessage) {
	var mouse = &mouseData{}
	err := json.Unmarshal(msg.Data, mouse)
	if err != nil {
		log.Error("unmarshal mouse data failed", zap.Error(err))
		return
	}

	user := mouse.User
	w.emulator.SetMouseState(user, uint(mouse.Button))
	w.emulator.SetMousePos(user, mouse.PosX, mouse.PosY)
}
