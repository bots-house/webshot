package main

import (
	"net/http"
	"os"

	"github.com/bots-house/webshot/internal/api"
	"github.com/bots-house/webshot/internal/renderer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	addr := "0.0.0.0:8000"

	rndr := &renderer.Chrome{Debug: false}

	api := api.New(api.Deps{Renderer: rndr})

	log.Info().Str("addr", "http://"+addr).Msg("listen http")
	http.ListenAndServe("0.0.0.0:8000", api)
}
