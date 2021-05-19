package renderer

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"strconv"

	"github.com/bots-house/webshot/internal"
	"golang.org/x/xerrors"
)

type Opts struct {
	// Viewport width in pixels of the browser render. Default is 1680
	Width int

	// Viewport height in pixels of the browser render. Default is 867
	Height int

	// Scale from 0
	Scale float64

	// Format of image
	Format internal.ImageFormat

	// Quality of image
	Quality int

	// Clip of viewport.
	// All fields is required.
	Clip OptsClip
}

func (opts Opts) Hash() string {
	buf := &bytes.Buffer{}

	buf.WriteString(strconv.Itoa(opts.getWidth()))
	buf.WriteString(strconv.Itoa(opts.getHeight()))
	buf.WriteString(strconv.FormatFloat(opts.getScale(), 'f', -1, 64))
	buf.WriteString(opts.Format.String())
	buf.WriteString(strconv.FormatFloat(float64(opts.Quality), 'f', -1, 64))

	if opts.Clip.IsSet() {
		buf.WriteString(strconv.FormatFloat(float64(*opts.Clip.X), 'f', -1, 64))
		buf.WriteString(strconv.FormatFloat(float64(*opts.Clip.Y), 'f', -1, 64))
		buf.WriteString(strconv.FormatFloat(float64(*opts.Clip.Width), 'f', -1, 64))
		buf.WriteString(strconv.FormatFloat(float64(*opts.Clip.Height), 'f', -1, 64))
	}

	h := sha1.New()

	io.Copy(h, buf)

	return hex.EncodeToString(h.Sum(nil))
}

func (opts *Opts) Validate() error {
	if err := opts.Clip.Validate(); err != nil {
		return xerrors.Errorf("validate clip: %w", err)
	}

	return nil
}

type OptsClip struct {
	X, Y          *float64
	Width, Height *float64
}

func (optsClip *OptsClip) IsSet() bool {
	return optsClip.X != nil && optsClip.Y != nil && optsClip.Width != nil && optsClip.Height != nil
}

func (optsClip *OptsClip) SetX(v float64) {
	optsClip.X = &v
}

func (optsClip *OptsClip) SetY(v float64) {
	optsClip.Y = &v
}

func (optsClip *OptsClip) SetWidth(v float64) {
	optsClip.Width = &v
}

func (optsClip *OptsClip) SetHeight(v float64) {
	optsClip.Height = &v
}

func (optsClip OptsClip) Validate() error {
	if optsClip.X == nil && optsClip.Y == nil && optsClip.Width == nil && optsClip.Height == nil {
		return nil
	}
	if optsClip.X == nil {
		return xerrors.Errorf("missing field `x`")
	}

	if optsClip.Y == nil {
		return xerrors.Errorf("missing field `y`")
	}

	if optsClip.Width == nil {
		return xerrors.Errorf("missing field `width`")
	}

	if optsClip.Height == nil {
		return xerrors.Errorf("missing field `height`")
	}

	return nil
}

const (
	defaultWidth  = 1680
	defaultHeight = 867
	defaultScale  = 1
)

func (opts *Opts) getScale() float64 {
	if opts.Scale == 0 {
		return defaultScale
	}

	return opts.Scale
}

func (opts *Opts) getWidth() int {
	if opts.Width == 0 {
		return defaultWidth
	}

	return opts.Width
}

func (opts *Opts) getHeight() int {
	if opts.Height == 0 {
		return defaultHeight
	}

	return opts.Height
}
