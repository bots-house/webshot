package handler

import (
	"net/http"

	"github.com/bots-house/webshot/internal/handler/api"
	"github.com/bots-house/webshot/internal/handler/middleware"
	"github.com/bots-house/webshot/internal/handler/web"

	"github.com/bots-house/webshot/internal/service"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Builder struct {
	Service *service.Service
	Auth    api.Auth
}

func (builder *Builder) Build() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(chimiddleware.Recoverer)

	router.Mount("/", web.New())

	router.Get(
		"/image",
		api.NewImageHandler(builder.Service, builder.Auth),
	)

	return router
}
