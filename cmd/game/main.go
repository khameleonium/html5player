package main

import (
	"html5gameplayer"
	"html5gameplayer/internal/api"
	"html5gameplayer/internal/engine"
	"html5gameplayer/internal/server"
	"log"
	"os"
)

// These variables are injected at build time via -ldflags
var (
	RunMode     = "embed" // "embed", "dir", "xor"
	XorKey      = ""      // Key for XOR decryption
	PakName     = "game_res.pak"
	DirName     = "game_data"
	Title       = "My HTML5 Game"
	Width       = "1280"
	Height      = "720"
	Fullscreen  = "false"
	Frameless   = "false"
	ContextMenu = "false"
)

func main() {
	var ds server.DataSource
	var err error

	switch RunMode {
	case "dir":
		ds = server.NewDirDataSource(DirName)
	case "xor":
		if XorKey == "" {
			log.Fatal("XOR key is empty!")
		}
		ds, err = server.NewXorDataSource(PakName, XorKey)
		if err != nil {
			log.Fatalf("Failed to initialize XOR datasource: %v", err)
		}
	case "embed-xor":
		if XorKey == "" {
			log.Fatal("XOR key is empty!")
		}
		data, err := html5gameplayer.GetGameDataXor()
		if err != nil {
			log.Fatalf("Failed to get embedded xor data: %v", err)
		}
		ds, err = server.NewXorBytesDataSource(data, XorKey)
		if err != nil {
			log.Fatalf("Failed to initialize embed-xor datasource: %v", err)
		}
	case "embed":
		fallthrough
	default:
		rawFS, err := html5gameplayer.GetGameData()
		if err != nil {
			log.Fatalf("Failed to get raw embed data: %v", err)
		}
		ds, err = server.NewEmbedDataSource(rawFS)
		if err != nil {
			log.Fatalf("Failed to initialize embed datasource: %v", err)
		}
	}

	// 1. Start Local HTTP Server
	port, err := server.Start(ds)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	url := "http://127.0.0.1:" + port

	// 2. Init Engine (WebView)
	eng := engine.New(Title, Width, Height)
	eng.SetFullscreen(Fullscreen == "true", Frameless == "true")
	defer eng.Destroy()

	if ContextMenu != "true" {
		eng.Init(`window.addEventListener("contextmenu", (e) => e.preventDefault());`)
	}

	// 3. Bind Go API to JavaScript (accessible via window.api)
	myApi := api.NewAPI()
	eng.Bind("api", myApi)

	// 4. Run Game
	eng.Navigate(url)
	eng.Run()

	server.Stop()
	os.Exit(0)
}
