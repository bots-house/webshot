package api

import (
	"io"
	"net/http"
	"time"

	"github.com/gorilla/schema"
	"golang.org/x/xerrors"

	"github.com/bots-house/webshot/internal"
	"github.com/bots-house/webshot/internal/renderer"
	"github.com/bots-house/webshot/internal/service"
)

type ScreenshotInput struct {
	URL string `schema:"url,required"`

	Width  int     `schema:"width"`
	Height int     `schema:"height"`
	Scale  float64 `schema:"scale"`

	Format  internal.ImageFormat `schema:"format"`
	Quality int                  `schema:"quality"`

	ScrollPage bool `schema:"scroll_page"`
	FullPage   bool `schema:"full_page"`
	Delay      int  `schema:"delay"`

	ClipX      *float64 `schema:"clip_x"`
	ClipY      *float64 `schema:"clip_y"`
	ClipWidth  *float64 `schema:"clip_width"`
	ClipHeight *float64 `schema:"clip_height"`

	Fresh bool `schema:"fresh"`
	TTL   int  `schema:"ttl"`
}

func NewImageHandler(srv *service.Service, auth Auth) http.HandlerFunc {
	return handleError(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		if err := r.ParseForm(); err != nil {
			err = xerrors.Errorf("parse form: %w", err)
			return httpError(err, http.StatusBadRequest)
		}

		input := &ScreenshotInput{}

		decoder := schema.NewDecoder()
		decoder.IgnoreUnknownKeys(true)

		if err := decoder.Decode(input, r.Form); err != nil {
			err = xerrors.Errorf("decode form: %w", err)
			return httpError(err, http.StatusUnprocessableEntity)
		}

		if auth != nil {
			if err := auth.Allow(ctx, r); err != nil {
				err = xerrors.Errorf("unathorized: %w", err)
				return httpError(err, http.StatusUnauthorized)
			}
		}

		renderOpts := renderer.Opts{
			Width:      input.Width,
			Height:     input.Height,
			Scale:      input.Scale,
			Format:     input.Format,
			Quality:    input.Quality,
			Delay:      time.Millisecond * time.Duration(input.Delay),
			FullPage:   input.FullPage,
			ScrollPage: input.ScrollPage,
			Clip: renderer.OptsClip{
				X:      input.ClipX,
				Y:      input.ClipY,
				Width:  input.ClipWidth,
				Height: input.ClipHeight,
			},
		}

		if err := renderOpts.Validate(); err != nil {
			err = xerrors.Errorf("validate opts: %w", err)
			return httpError(err, http.StatusUnprocessableEntity)
		}

		cacheOpts := service.CacheOpts{
			TTL:   time.Second * time.Duration(input.TTL),
			Fresh: input.Fresh,
		}

		output, err := srv.Shot(ctx, input.URL, service.ShotOpts{
			Render: renderOpts,
			Cache:  cacheOpts,
		})

		if err != nil {
			return xerrors.Errorf("render error: %w", err)
		}

		w.Header().Set("Content-Type", renderOpts.Format.ContentType())

		_, err = io.Copy(w, output)
		if err != nil {
			return xerrors.Errorf("copy output: %w", err)
		}

		return nil
	})
}
