package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// FS embed this folder and childs into binary
//go:embed *
var FS embed.FS

var indexTmpl = template.Must(template.ParseFS(FS, "index.html"))

func New() chi.Router {
	router := chi.NewRouter()

	router.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(FS))))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		w.Header().Set("Content-Type", "text/html")

		if err := indexTmpl.Execute(w, nil); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("render template")
			http.Error(w, fmt.Sprintf("fail to render template: %s", err), http.StatusInternalServerError)
		}
	})

	return router
}
