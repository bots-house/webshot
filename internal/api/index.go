package api

import (
	"net/http"

	"github.com/bots-house/webshot/internal/api/public"
)

func IndexHandler() http.Handler {
	return http.FileServer(http.FS(public.FS))
}
