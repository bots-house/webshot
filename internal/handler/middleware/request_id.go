package middleware

import (
	"net/http"

	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

var (
	inputHeaders = []string{
		"X-Request-ID",
		"CF-Request-ID",
	}
	outputHeader = "X-Request-ID"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// find first matched header
		var id string

		for _, headerName := range inputHeaders {
			v := r.Header.Get(headerName)
			if v != "" {
				id = v
				break
			}
		}

		if id == "" {
			id = xid.New().String()
		}

		ctx := r.Context()

		logger := log.Ctx(ctx).
			With().
			Str("req_id", id).
			Logger()

		ctx = logger.WithContext(ctx)

		rw.Header().Set(outputHeader, id)

		r = r.WithContext(ctx)
		next.ServeHTTP(rw, r)
	})
}
