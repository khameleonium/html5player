//go:build embed_xor

package html5gameplayer

import (
	_ "embed"
	"fmt"
	"io/fs"
)

//go:embed game_data_xor.pak
var gameDataXor []byte

func GetGameData() (fs.FS, error) {
	return nil, fmt.Errorf("raw game data not available in embed-xor mode")
}

func GetGameDataXor() ([]byte, error) {
	return gameDataXor, nil
}
