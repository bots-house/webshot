package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/tomasen/realip"

	"github.com/bots-house/webshot/internal/renderer"
)

func ScreenshotHandler(rndr renderer.Renderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		url := r.Form.Get("url")
		if url == "" {
			http.Error(w, "missing required parameter `url`", http.StatusUnprocessableEntity)
			return
		}

		var opts renderer.Opts

		width := r.Form.Get("width")
		if width != "" {
			var err error

			opts.Width, err = strconv.Atoi(width)
			if err != nil {
				http.Error(w, fmt.Sprintf("parameter `width` not is int: %v", err), http.StatusUnprocessableEntity)
				return
			}
		}

		height := r.Form.Get("height")
		if height != "" {
			var err error

			opts.Height, err = strconv.Atoi(height)
			if err != nil {
				http.Error(w, fmt.Sprintf("parameter `height` not is int: %v", err), http.StatusUnprocessableEntity)
				return
			}
		}

		scale := r.Form.Get("scale")
		if scale != "" {
			var err error

			opts.Scale, err = strconv.ParseFloat(scale, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("parameter `scale` not is float64: %v", err), http.StatusUnprocessableEntity)
				return
			}
		}

		opts.Format = renderer.ImageTypePNG
		if v := r.Form.Get("format"); v != "" {
			var err error
			opts.Format, err = renderer.ParseImageType(v)
			if err != nil {
				http.Error(w, fmt.Sprintf("parameter `format` is invalid: %s", v), http.StatusUnprocessableEntity)
				return
			}
		}

		if v := r.Form.Get("quality"); v != "" {
			var err error
			opts.Quality, err = strconv.Atoi(v)
			if err != nil {
				http.Error(w, fmt.Sprintf("parameter `quality` is not int: %v", err), http.StatusUnprocessableEntity)
				return
			}
		}

		log.Info().
			Str("url", url).
			Int("width", opts.Width).
			Int("height", opts.Height).
			Float64("scale", opts.Scale).
			Str("format", opts.Format.String()).
			Str("ip", realip.FromRequest(r)).
			Msg("screenshot call")

		// if err := json.NewEncoder(w).Encode(struct {
		// 	URL  string
		// 	Opts renderer.Opts
		// }{
		// 	URL:  url,
		// 	Opts: opts,
		// }); err != nil {
		// 	log.Printf("dump request error: %v", err)
		// }
		ctx := r.Context()

		output, err := rndr.Render(ctx, url, &opts)
		if err != nil {
			log.Error().Err(err).Msg("render error")
			http.Error(w, fmt.Sprintf("render error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", opts.Format.ContentType())

		n, err := io.Copy(w, output)
		if err != nil {
			log.Error().Err(err).Msg("copy output")
			return
		}

		log.Debug().Int64("size", n).Msg("image generated")
	})
}
