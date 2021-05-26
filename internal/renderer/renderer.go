package renderer

import (
	"context"
)

type Renderer interface {
	Render(ctx context.Context, url string, opts Opts) ([]byte, error)
}
