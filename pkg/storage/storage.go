package storage

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type (
	Storage struct {
		games []GameMeta
		cores []CoreMeta
	}

	GameMeta struct {
		// e.g: mario.nes
		Name     string // mario
		FileType string // nes
		Path     string // absolute path to game file
	}

	CoreMeta struct {
		Name          string
		SupportedType map[string]interface{} // the game extension that Core supports
		Path          string
	}
)

func New() *Storage {
	s := &Storage{}

	s.loadAllGamesMetadata()
	s.LoadCoresMetadata()
	return s
}

func (s *Storage) loadAllGamesMetadata() {
	dir := "./pkg/storage/game"

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Println("error readDir: %w", err)
		return
	}

	path, err := filepath.Abs(dir)
	if err != nil {
		log.Println("error absolute path: %w")
		return
	}

	res := make([]GameMeta, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		arr := strings.Split(f.Name(), ".")
		name := arr[0]
		fileType := arr[1]

		res = append(res, GameMeta{
			Name:     name,
			FileType: fileType,
			Path:     filepath.Join(path, f.Name()),
		})
	}

	s.games = res
}

func (s *Storage) LoadCoresMetadata() {
	dir := "./pkg/storage/core"

	path, err := filepath.Abs(dir)
	if err != nil {
		return
	}

	res := make([]CoreMeta, 0, 3)

	res = append(res, CoreMeta{
		Name: "mame2010_libretro",
		SupportedType: map[string]interface{}{
			"zip": struct{}{}, "chd": struct{}{},
		},
		Path: filepath.Join(path, "mame2010_libretro.so"),
	})

	res = append(res, CoreMeta{
		Name: "nestopia_libretro",
		SupportedType: map[string]interface{}{
			"nes": struct{}{},
		},
		Path: filepath.Join(path, "nestopia_libretro.so"),
	})

	res = append(res, CoreMeta{
		Name: "snes9x2010_libretro",
		SupportedType: map[string]interface{}{
			"sfc": struct{}{}, "smc": struct{}{},
		},
		Path: filepath.Join(path, "snes9x2010_libretro.so"),
	})

	res = append(res, CoreMeta{
		Name: "mednafen_gba_libretro",
		SupportedType: map[string]interface{}{
			"gba": struct{}{},
		},
		Path: filepath.Join(path, "mednafen_gba_libretro.so"),
	})

	s.cores = res
}

func (s *Storage) GetAllGamesMetadata() []GameMeta {
	return s.games
}

func (s *Storage) GetGameMetadata(name string) (GameMeta, error) {
	for _, gameMeta := range s.games {
		if gameMeta.Name == name {
			return gameMeta, nil
		}
	}

	return GameMeta{}, errors.New("game not found")
}

func (s *Storage) GetSuitableCore(fileType string) (CoreMeta, error) {
	for _, core := range s.cores {
		if _, ok := core.SupportedType[fileType]; ok {
			return core, nil
		}
	}

	return CoreMeta{}, errors.New("filetype not supported")
}
