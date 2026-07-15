package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"html5gameplayer/internal/api"
	"html5gameplayer/internal/engine"
	"html5gameplayer/internal/server"
)

type Config struct {
	Title       string
	Width       int
	Height      int
	DevTools    bool
	MuteAudio   bool
	Fullscreen  bool
	Frameless   bool
	ContextMenu bool
}

const ConfigFile = "html5player_config.json"

func loadConfig() Config {
	// Default config
	cfg := Config{
		Title:       "HTML5 Portable Game",
		Width:       1280,
		Height:      720,
		DevTools:    false,
		MuteAudio:   false,
		Fullscreen:  false,
		Frameless:   false,
		ContextMenu: false,
	}

	if data, err := os.ReadFile(ConfigFile); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			fmt.Println("Ошибка чтения конфига, используем стандартные настройки:", err)
		}
	} else {
		// Create default config if it doesn't exist
		if data, err := json.MarshalIndent(cfg, "", "  "); err == nil {
			os.WriteFile(ConfigFile, data, 0644)
		}
	}
	return cfg
}

func main() {
	cfg := loadConfig()

	// 1. Serve current directory "."
	ds := server.NewDirDataSource(".")

	// Check if index.html exists
	if _, err := os.Stat("index.html"); os.IsNotExist(err) {
		log.Println("ПРЕДУПРЕЖДЕНИЕ: index.html не найден в текущей папке. Убедитесь, что лаунчер лежит рядом с игрой!")
	}

	// 2. Start Local HTTP Server
	port, err := server.Start(ds)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	url := "http://127.0.0.1:" + port

	// 3. Init Engine
	eng := engine.New(cfg.Title, strconv.Itoa(cfg.Width), strconv.Itoa(cfg.Height))
	eng.SetFullscreen(cfg.Fullscreen, cfg.Frameless)
	defer eng.Destroy()

	// Inject Mute Script
	if cfg.MuteAudio {
		eng.Init(`
			window.addEventListener("DOMContentLoaded", () => {
				const muteMedia = () => {
					document.querySelectorAll('audio, video').forEach(m => m.muted = true);
				};
				muteMedia();
				setInterval(muteMedia, 1000); // Mute dynamically added elements
			});
		`)
	}

	// Disable Context Menu
	if !cfg.ContextMenu {
		eng.Init(`window.addEventListener("contextmenu", (e) => e.preventDefault());`)
	}

	// 4. Bind Go API
	myApi := api.NewAPI()
	eng.Bind("api", myApi)

	// 5. Run Game
	eng.Navigate(url)
	eng.Run()

	server.Stop()
	os.Exit(0)
}
