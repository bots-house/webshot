package api2

import (
	"net/http"

	"github.com/bots-house/webshot/internal/api2/gen/restapi"
	"github.com/bots-house/webshot/internal/api2/gen/restapi/operations"
	"github.com/bots-house/webshot/internal/renderer"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
)

type Handler struct {
	Renderer renderer.Renderer
}

func (h Handler) newAPI() *operations.WebShotAPI {
	spec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)

	if err != nil {
		panic("load spec failed: " + err.Error())
	}
	api := operations.NewWebShotAPI(spec)

	api.UseSwaggerUI()

	return api
}

func (h Handler) setupProducersAndConsumers(api *operations.WebShotAPI) {
	// set JSON producer/consumer
	api.JSONProducer = runtime.JSONProducer()
	api.JSONConsumer = runtime.JSONConsumer()
}

func (h Handler) setupHandlers(api *operations.WebShotAPI) {

}

func (h Handler) wrapMiddleware(handler http.Handler) http.Handler {
	// handler = common.WrapMiddlewareFS(handler, h.Service.Config.MediaStoragePath)
	// handler = h.wrapMiddlewareSentryHub(handler)
	// handler = h.wrapMiddlewareLogger(handler)

	// handler = h.wrapMiddlewareRecovery(handler)

	return handler
}

func (h Handler) Make() http.Handler {
	api := h.newAPI()

	h.setupProducersAndConsumers(api)
	h.setupHandlers(api)
	// h.setupMiddleware(api)
	// h.setupAuth(api)

	return h.wrapMiddleware(api.Serve(nil))
}
