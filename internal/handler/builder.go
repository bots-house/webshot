package handler

import (
	"net/http"
	"time"

	"github.com/bots-house/webshot/internal"
	"github.com/bots-house/webshot/internal/handler/api"
	"github.com/bots-house/webshot/internal/handler/middleware"
	"github.com/bots-house/webshot/internal/handler/web"
	sentryhttp "github.com/getsentry/sentry-go/http"

	"github.com/bots-house/webshot/internal/service"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Builder struct {
	Service   *service.Service
	Auth      api.Auth
	BuildInfo internal.BuildInfo
	Sentry    bool
}

type SentryWrapper interface {
	Handle(handler http.Handler) http.Handler
}

type sentryWrapperStub struct{}

func (stub sentryWrapperStub) Handle(h http.Handler) http.Handler {
	return h
}

func (builder *Builder) Build() http.Handler {
	router := chi.NewRouter()

	var sentryWrapper SentryWrapper

	if builder.Sentry {
		sentryWrapper = sentryhttp.New(sentryhttp.Options{Repanic: true})
	} else {
		sentryWrapper = sentryWrapperStub{}
	}

	router.Use(middleware.RequestID)
	router.Use(chimiddleware.Recoverer)

	router.Mount("/", web.New())

	router.Method(
		http.MethodGet,
		"/image",
		sentryWrapper.Handle(api.NewImageHandler(builder.Service, builder.Auth)),
	)

	router.Get("/version", api.NewVersionHandler(builder.BuildInfo))
	router.Get("/health", api.NewHealthHandler(time.Now()))

	return router
}
