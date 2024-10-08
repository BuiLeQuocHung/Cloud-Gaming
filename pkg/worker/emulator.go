package worker

import (
	"cloud_gaming/pkg/emulator"
	"cloud_gaming/pkg/libretro"
	"cloud_gaming/pkg/log"
	"errors"

	"go.uber.org/zap"
)

type (
	StartGameRequest struct {
		Game string `json:"game"`
	}

	StopGameRequest struct{}
)

func (w *Worker) startEmulator(r *StartGameRequest) error {
	if !w.emulator.IsReady() {
		return errors.New("emulator is running")
	}

	gameMeta, err := w.storage.GetGameMetadata(r.Game)
	if err != nil {
		return err
	}

	coreMeta, err := w.storage.GetSuitableCore(gameMeta.FileType)
	if err != nil {
		return err
	}

	err = w.emulator.LoadCore(
		coreMeta.Path,
		w.environmentCallback,
		w.videoRefreshCallback,
		w.audioSampleCallback,
		w.audioSampleBatchCallback,
	)
	if err != nil {
		return err
	}

	w.emulator.Init()
	err = w.emulator.LoadGame(gameMeta.Path)
	if err != nil {
		return err
	}

	// init video/audio pipeline here because
	// some infos can only be retrieved after core and game are loaded
	systemAVInfo := w.emulator.GetSystemAVInfo()
	w.setSystemAVInfo(&systemAVInfo)

	log.Debug("system av info", zap.Any("info", systemAVInfo))
	log.Debug("start game", zap.String("game", r.Game))
	go w.startGame()
	return nil
}

func (w *Worker) startGame() {
	w.emulator.SetState(emulator.Running)

	for w.emulator.IsRunning() {
		w.emulator.Run()
	}

	w.stopGame()
}

func (w *Worker) stopGame() {
	w.emulator.SetState(emulator.Deinitializing)
	w.emulator.UnloadGame()
	w.emulator.DeInit()

	w.videoPipe.Close()
	w.audioPipe.Close()

	w.emulator.SetState(emulator.Ready)
}

func (w *Worker) stopEmulator() {
	if w.emulator.IsRunning() {
		w.emulator.SetState(emulator.Deinitializing)
	}
}

func (w *Worker) setSystemAVInfo(systemAVInfo *libretro.SystemAVInfo) {
	w.videoPipe.SetSystemVideoInfo(systemAVInfo)
	w.audioPipe.SetSystemAudioInfo(systemAVInfo)
}
