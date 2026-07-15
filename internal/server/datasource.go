package server

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"
)

// We define our DataSource as a standard fs.FS
type DataSource = fs.FS

// 1. Embed Data Source (using standard go:embed)
// The actual embed variable must be in the main package or a package that contains the files.
// But we can pass it here. Actually, to keep it simple, we will embed it directly here if we could.
// But `game_data` will be at the root of the project.
// Wait, `go:embed` requires the files to be inside the same package or subdirectories.
// So we define it here, and the generator will place `game_data` next to this package?
// No, the generator will place `game_data` at the project root.
// We can't use `//go:embed ../../game_data` - go:embed doesn't allow parent directories.
// So the `go:embed` MUST be in a package at the root or we copy game_data.
// Since the project generator creates the project, we'll put `//go:embed game_data/*` in a special root package, or we just put the server at the root.
// Let's just put `//go:embed game_data/*` in this file, and the user must put their files in `internal/server/game_data`?
// No, that's ugly. We will put `//go:embed game_data/*` in `cmd/game/main.go` and pass it to `server.NewEmbedDataSource(embeddedFS)`.

func NewEmbedDataSource(f fs.FS) (DataSource, error) {
	// Extract the "game_data" subfolder
	subFS, err := fs.Sub(f, "game_data")
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded game data: %w", err)
	}
	return subFS, nil
}

// 2. Directory Data Source
func NewDirDataSource(dirPath string) DataSource {
	return os.DirFS(dirPath)
}

// 3. XOR Zip Data Source
func NewXorDataSource(archivePath string, key string) (DataSource, error) {
	// Read the entire XORed file into memory
	data, err := os.ReadFile(archivePath)
	if err != nil {
		return nil, err
	}

	// Decrypt in memory
	keyBytes := []byte(key)
	keyLen := len(keyBytes)
	if keyLen == 0 {
		return nil, fmt.Errorf("empty XOR key")
	}

	for i := 0; i < len(data); i++ {
		data[i] ^= keyBytes[i%keyLen]
	}

	// Open ZIP from memory
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	return zipReader, nil
}
