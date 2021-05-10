package renderer

import (
	"context"
	"io"
)

type Renderer interface {
	Render(ctx context.Context, url string, opts *Opts) (io.Reader, error)
}
