package api

import (
	"encoding/json"
	"net/http"

	"github.com/bots-house/webshot/internal"
	"github.com/rs/zerolog/log"
)

func NewVersionHandler(info internal.BuildInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(info); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("encode build info failed")
		}
	}
}
