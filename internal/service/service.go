package service

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"time"

	"github.com/bots-house/webshot/internal/renderer"
	"github.com/bots-house/webshot/internal/storage"
	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
)

type Service struct {
	Renderer renderer.Renderer
	Storage  storage.Storage
}

type CacheOpts struct {
	TTL   time.Duration
	Fresh bool
}

func (opts *CacheOpts) getTTL() time.Duration {
	if opts.TTL == 0 {
		return time.Hour * 24 * 30
	}

	return opts.TTL
}

type ShotOpts struct {
	Render renderer.Opts
	Cache  CacheOpts
}

func (srv *Service) Shot(
	ctx context.Context,
	targetURL string,
	opts ShotOpts,
) (io.Reader, error) {
	if srv.Storage == nil {
		return srv.shotNoStorage(ctx, targetURL, opts.Render)
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, xerrors.Errorf("parse url: %w", err)
	}

	meta := storage.Meta{
		URL:    u,
		Opts:   opts.Render.Hash(),
		Format: opts.Render.Format,
	}

	renderAndSave := func(ctx context.Context) (io.Reader, error) {
		output, err := srv.Renderer.Render(ctx, targetURL, opts.Render)
		if err != nil {
			return nil, xerrors.Errorf("render error: %w", err)
		}

		if err := srv.Storage.Upload(ctx, storage.Upload{
			Meta: meta,
			TTL:  opts.Cache.getTTL(),
			Body: bytes.NewReader(output),
		}); err != nil {
			return nil, xerrors.Errorf("upload to stroge: %w", err)
		}

		return bytes.NewReader(output), nil
	}

	// if client want fresh screen
	if opts.Cache.Fresh {
		return renderAndSave(ctx)
	}

	// try to use cached image
	body, err := srv.Storage.Get(ctx, meta)
	if err == storage.ErrFileNotFound || err == storage.ErrFileExpired || err == storage.ErrFileCorrupted {
		log.Ctx(ctx).Debug().Err(err).Msg("something wrong with file, render new")
		return renderAndSave(ctx)
	} else if err != nil {
		return nil, xerrors.Errorf("storage get: %w", err)
	}

	return body, nil
}

func (srv *Service) shotNoStorage(ctx context.Context, url string, opts renderer.Opts) (io.Reader, error) {
	output, err := srv.Renderer.Render(ctx, url, opts)
	if err != nil {
		return nil, xerrors.Errorf("render error: %w", err)
	}

	return bytes.NewReader(output), nil
}
