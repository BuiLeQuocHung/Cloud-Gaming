package worker

import (
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
		log.Error("get game metadata failed", zap.Error(err))
		return err
	}

	coreMeta, err := w.storage.GetSuitableCore(gameMeta.FileType)
	if err != nil {
		log.Error("get core metadata failed", zap.Error(err))
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
		log.Error("load core failed", zap.Error(err))
		return err
	}

	w.emulator.Init()
	err = w.emulator.LoadGame(gameMeta.Path)
	if err != nil {
		log.Error("load game failed", zap.Error(err))
		return err
	}

	systemAVInfo := w.emulator.GetSystemAVInfo()
	w.setSystemAVInfo(&systemAVInfo)

	w.videoPipe.Start()
	w.emulator.StartGame()
	return nil
}

func (w *Worker) stopEmulator() {
	w.emulator.StopGame()
	w.videoPipe.Close()
	w.audioPipe.Close()
}

func (w *Worker) setSystemAVInfo(systemAVInfo *libretro.SystemAVInfo) {
	w.videoPipe.SetSystemVideoInfo(systemAVInfo)
	w.audioPipe.SetSystemAudioInfo(systemAVInfo)
}
