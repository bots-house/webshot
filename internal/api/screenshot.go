package api

import (
	"io"
	"net/http"

	"github.com/gorilla/schema"
	"golang.org/x/xerrors"

	"github.com/bots-house/webshot/internal/renderer"
)

type ScreenshotInput struct {
	URL string `schema:"url,required"`

	Width  int     `schema:"width"`
	Height int     `schema:"height"`
	Scale  float64 `schema:"scale"`

	Format  renderer.ImageFormat `schema:"format"`
	Quality int                  `schema:"quality"`

	ClipX *float64 `schema:"clip_x"`
	ClipY *float64 `schema:"clip_y"`

	ClipWidth  *float64 `schema:"clip_width"`
	ClipHeight *float64 `schema:"clip_height"`
}

func ScreenshotHandler(rndr renderer.Renderer) http.Handler {
	return handleError(func(w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			err = xerrors.Errorf("parse form: %w", err)
			return httpError(err, http.StatusBadRequest)
		}

		input := &ScreenshotInput{}

		if err := schema.NewDecoder().Decode(input, r.Form); err != nil {
			err = xerrors.Errorf("decode form: %w", err)
			return httpError(err, http.StatusUnprocessableEntity)
		}

		opts := renderer.Opts{
			Width:   input.Width,
			Height:  input.Height,
			Scale:   input.Scale,
			Format:  input.Format,
			Quality: input.Quality,
			Clip: renderer.OptsClip{
				X:      input.ClipX,
				Y:      input.ClipY,
				Width:  input.ClipWidth,
				Height: input.ClipHeight,
			},
		}

		if err := opts.Validate(); err != nil {
			err = xerrors.Errorf("validate opts: %w", err)
			return httpError(err, http.StatusUnprocessableEntity)
		}

		ctx := r.Context()

		output, err := rndr.Render(ctx, input.URL, &opts)
		if err != nil {
			return xerrors.Errorf("render error: %w", err)
		}

		w.Header().Set("Content-Type", opts.Format.ContentType())

		_, err = io.Copy(w, output)
		if err != nil {
			return xerrors.Errorf("copy output: %w", err)
		}

		return nil
	})
}
