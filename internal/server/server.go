package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

var httpServer *http.Server

// Start launches a local HTTP server serving the provided DataSource
func Start(ds DataSource) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	httpServer = &http.Server{
		Handler: http.FileServer(http.FS(ds)),
	}

	go func() {
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	return fmt.Sprintf("%d", port), nil
}

func Stop() {
	if httpServer != nil {
		httpServer.Shutdown(context.Background())
	}
}
