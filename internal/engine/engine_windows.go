//go:build windows

package engine

import (
	"strconv"

	"github.com/jchv/go-webview2"
)

type Engine struct {
	w webview2.WebView
}

func New(title, widthStr, heightStr string) *Engine {
	w, _ := strconv.Atoi(widthStr)
	h, _ := strconv.Atoi(heightStr)

	debug := false
	wv := webview2.New(debug)
	wv.SetTitle(title)
	wv.SetSize(w, h, webview2.HintNone)

	return &Engine{w: wv}
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

