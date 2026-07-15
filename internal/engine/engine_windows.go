//go:build windows

package engine

import (
	"strconv"
	"syscall"

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

func (e *Engine) SetFullscreen(fullscreen bool, frameless bool) {
	if !fullscreen && !frameless {
		return
	}

	hwnd := uintptr(e.w.Window())

	user32 := syscall.NewLazyDLL("user32.dll")
	setWindowLongPtr := user32.NewProc("SetWindowLongPtrW")
	if err := setWindowLongPtr.Find(); err != nil {
		setWindowLongPtr = user32.NewProc("SetWindowLongW")
	}
	getWindowLongPtr := user32.NewProc("GetWindowLongPtrW")
	if err := getWindowLongPtr.Find(); err != nil {
		getWindowLongPtr = user32.NewProc("GetWindowLongW")
	}
	showWindow := user32.NewProc("ShowWindow")

	gwlStyle := int32(-16)
	const WS_CAPTION = 0x00C00000
	const WS_THICKFRAME = 0x00040000
	const SW_SHOWMAXIMIZED = 3

	if frameless || fullscreen {
		style, _, _ := getWindowLongPtr.Call(hwnd, uintptr(gwlStyle))
		style &^= (WS_CAPTION | WS_THICKFRAME)
		setWindowLongPtr.Call(hwnd, uintptr(gwlStyle), style)
	}

	if fullscreen {
		showWindow.Call(hwnd, uintptr(SW_SHOWMAXIMIZED))
	}
}
