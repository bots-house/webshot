package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/bots-house/webshot/internal"
	"golang.org/x/xerrors"
)

// Meta contains metadata about screenshot
type Meta struct {
	// URL of target site
	URL *url.URL

	// Hash of options
	Opts string

	// Format of file
	Format internal.ImageFormat
}

// Upload define object to upload to s3.
type Upload struct {
	Meta
	TTL  time.Duration
	Body io.Reader
}

var (
	ErrFileNotFound  = xerrors.New("file not found")
	ErrFileCorrupted = xerrors.New("file corrupted")
	ErrFileExpired   = xerrors.New("file expired")
)

type Storage interface {
	// Get returns URL of file in storage if it exists
	Get(ctx context.Context, meta Meta) (io.Reader, error)

	// Put image to storage
	Upload(ctx context.Context, upload Upload) error
}
