package api

import (
	"net/http"

	"github.com/bots-house/webshot/internal/renderer"
)

type Deps struct {
	Renderer renderer.Renderer
}

func New(deps Deps) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/", IndexHandler())
	mux.Handle("/screenshot", ScreenshotHandler(deps.Renderer))

	return mux
}
