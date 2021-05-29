package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

func handleError(h func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
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
