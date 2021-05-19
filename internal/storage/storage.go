package storage

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/bots-house/webshot/internal"
	"golang.org/x/xerrors"
)

type Meta struct {
	// URL of target site
	URL *url.URL

	// Hash of options
	Opts string

	// Format of file
	Format internal.ImageFormat
}

type Upload struct {
	Meta
	TTL  time.Duration
	Body io.Reader
}

func (f *Meta) Path() string {
	h := sha1.New()
	h.Write([]byte(f.URL.String()))
	h.Write([]byte(f.Opts))
	hash := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("/%s/%s.%s", f.URL.Hostname(), hash, f.Format.Ext())
}

var (
	ErrFileNotFound = xerrors.New("file not found")
)

type Storage interface {
	// Get returns URL of file in storage if it exists
	Get(ctx context.Context, meta Meta) (io.Reader, error)

	// Has file or not
	Has(ctx context.Context, meta Meta) (bool, error)

	// Put image to storage
	Upload(ctx context.Context, upload Upload) error
}
