package worker

import (
	"cloud_gaming/pkg/emulator"
	"cloud_gaming/pkg/libretro"
	"errors"
	"log"
)

type (
	StartGameRequest struct {
		Game string `json:"game"`
	}

	StopGameRequest struct{}
)

func (w *Worker) startEmulator(r *StartGameRequest) error {
	log.Println("in here")
	if !w.emulator.IsReady() {
		log.Println("error emulator is not ready")
		return errors.New("emulator is running")
	}

	gameMeta, err := w.storage.GetGameMetadata(r.Game)
	if err != nil {
		log.Println("error get game metadata")
		return err
	}

	coreMeta, err := w.storage.GetSuitableCore(gameMeta.FileType)
	if err != nil {
		log.Println("error get emulator")
		return err
	}

	log.Println("core meta: ", coreMeta.Name, coreMeta.Path, coreMeta.SupportedType)
	err = w.emulator.LoadCore(
		coreMeta.Path,
		w.environmentCallback,
		w.videoRefreshCallback,
		w.audioSampleCallback,
		w.audioSampleBatchCallback,
	)
	if err != nil {
		log.Println("load core error: %w", err)
		return err
	}

	log.Println("emulator init")
	w.emulator.Init()
	// time.Sleep(5 * time.Second)
	err = w.emulator.LoadGame(gameMeta.Path)
	if err != nil {
		log.Println("load game error: %w", err)
		return err
	}

	// init video/audio pipeline here because
	// some infos can only be retrieved after core and game are loaded
	log.Println("load audio/video pipeline")
	systemAVInfo := w.emulator.GetSystemAVInfo()
	w.setSystemAVInfo(&systemAVInfo)
	log.Println("system audio.video info: ", systemAVInfo)

	go w.startGame()
	return nil
}

func (w *Worker) startGame() {
	w.emulator.SetState(emulator.Running)
	// log.Println("start game")

	for w.emulator.IsRunning() {
		w.emulator.Run()
	}

	// log.Println("game stops ?")
	w.stopGame()
}

func (w *Worker) stopGame() {
	w.emulator.SetState(emulator.Deinitializing)
	w.emulator.UnloadGame()
	w.emulator.DeInit()

	w.videoPipe.Close()
	w.audioPipe.Close()

	// wait for video/audio pipeline to close
	// need better approach

	w.emulator.SetState(emulator.Ready)
}

func (w *Worker) stopEmulator() {
	w.emulator.SetState(emulator.Deinitializing)
}

func (w *Worker) setSystemAVInfo(systemAVInfo *libretro.SystemAVInfo) {
	w.videoPipe.SetSystemVideoInfo(systemAVInfo)
	w.audioPipe.SetSystemAudioInfo(systemAVInfo)
}
