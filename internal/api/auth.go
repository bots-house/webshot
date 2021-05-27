package api

import (
	"context"
	"net/http"
)

type Auth interface {
	Allow(ctx context.Context, r *http.Request) error
}
