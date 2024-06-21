package scraper

import (
	"context"
	"fmt"
	"mime"
	"net/http"

	"github.com/cornelk/gotokit/log"
)

// set more mime types in the browser, this for example fixes .asp files not being
// downloaded but handled as html.
var mimeTypes = map[string]string{
	".asp": "text/html; charset=utf-8",
}

func ServeDirectory(ctx context.Context, path string, port int16, logger *log.Logger) error {
	fs := http.FileServer(http.Dir(path))
	mux := http.NewServeMux()
	mux.Handle("/", fs) // server root by file system

	// update mime types
	for ext, mt := range mimeTypes {
		if err := mime.AddExtensionType(ext, mt); err != nil {
			return fmt.Errorf("adding mime type '%s': %w", ext, err)
		}
	}

	fullAddr := fmt.Sprintf("http://127.0.0.1:%d", port)
	logger.Info("Serving directory...",
		log.String("path", path),
		log.String("address", fullAddr))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		//nolint: contextcheck
		if err := server.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("shutting down webserver: %w", err)
		}
		return nil

	case err := <-serverErr:
		return fmt.Errorf("starting webserver: %w", err)
	}
}
