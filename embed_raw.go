//go:build !embed_xor

package html5gameplayer

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed game_data/*
var gameDataRaw embed.FS

func GetGameData() (fs.FS, error) {
	return fs.Sub(gameDataRaw, "game_data")
}

func GetGameDataXor() ([]byte, error) {
	return nil, fmt.Errorf("embed-xor not enabled")
}
