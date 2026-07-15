//go:build !windows

package engine

import (
	"strconv"

	"github.com/webview/webview_go"
)

type Engine struct {
	w            webview.WebView
	isFullscreen bool
	isFrameless  bool
}

func New(title, widthStr, heightStr string) *Engine {
	w, _ := strconv.Atoi(widthStr)
	h, _ := strconv.Atoi(heightStr)

	debug := false
	wv := webview.New(debug)
	wv.SetTitle(title)
	wv.SetSize(w, h, webview.HintNone)

	eng := &Engine{w: wv}

	eng.Bind("__os_toggleFullscreen", func() {
		eng.SetFullscreen(!eng.isFullscreen, eng.isFrameless)
	})

	eng.Init(`
		window.addEventListener("keydown", (e) => {
			if (e.ctrlKey && e.altKey && e.key === "Enter") {
				window.__os_toggleFullscreen();
			}
		});
	`)

	return eng
}

func (e *Engine) Destroy() {
	e.w.Destroy()
}

func (e *Engine) Navigate(url string) {
	e.w.Navigate(url)
}

func (e *Engine) Run() {
	e.w.Run()
}

func (e *Engine) Bind(name string, v interface{}) {
	e.w.Bind(name, v)
}

func (e *Engine) Init(js string) {
	e.w.Init(js)
}

func (e *Engine) SetFullscreen(fullscreen bool, frameless bool) {
	e.isFullscreen = fullscreen
	e.isFrameless = frameless
	// Not implemented natively in pure Go for GTK/Cocoa yet.
}
