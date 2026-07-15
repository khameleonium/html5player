package api

import (
	"fmt"
)

// API struct exposed to JavaScript
type API struct {
}

func NewAPI() *API {
	return &API{}
}

// Ping is a test method. In JS: window.api.Ping("Test").then(console.log)
func (a *API) Ping(message string) string {
	fmt.Println("JS sent:", message)
	return "Go received: " + message
}

// Add your custom methods here! For example:
// func (a *API) SaveGame(data string) error { ... }
// func (a *API) LoadGame() (string, error) { ... }
