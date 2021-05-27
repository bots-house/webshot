package api

import (
	"encoding/json"
	"net/http"

	"github.com/bots-house/webshot/internal/api/middleware"
	"github.com/bots-house/webshot/internal/service"
	"github.com/justinas/alice"
	"github.com/rs/zerolog/log"
)

func New(srv *service.Service, auth Auth) *http.ServeMux {
	mux := http.NewServeMux()

	chain := alice.New(
		middleware.RequestID,
	)

	mux.Handle("/", chain.Then(IndexHandler()))
	mux.Handle("/screenshot", chain.Then(ScreenshotHandler(srv, auth)))

	return mux
}

func handleError(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := h(w, r)

		if err != nil {
			switch e := err.(type) {
			case *HTTPError:
				w.WriteHeader(e.Code)
				if err := json.NewEncoder(w).Encode(e); err != nil {
					log.Ctx(ctx).Error().Err(err).Msg("encode status error")
					return
				}
			default:
				herr := httpError(err, http.StatusInternalServerError)
				w.WriteHeader(herr.Code)
				if err := json.NewEncoder(w).Encode(herr); err != nil {
					log.Ctx(ctx).Error().Err(err).Msg("encode status error")
					return
				}
			}
		}
	})
}
