package renderer

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

type ImageFormat int8

const (
	ImageTypePNG ImageFormat = iota
	ImageTypeJPEG
)

func ParseImageType(v string) (ImageFormat, error) {
	switch strings.ToLower(v) {
	case "png":
		return ImageTypePNG, nil
	case "jpeg":
		return ImageTypeJPEG, nil
	default:
		return ImageFormat(-1), fmt.Errorf("unsupported image type: %s", v)
	}
}

func (it ImageFormat) ContentType() string {
	switch it {
	case ImageTypePNG:
		return "image/png"
	case ImageTypeJPEG:
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func (it ImageFormat) String() string {
	switch it {
	case ImageTypeJPEG:
		return "jpeg"
	case ImageTypePNG:
		return "png"
	default:
		return "unknown"
	}
}

type Opts struct {
	// Viewport width in pixels of the browser render. Default is 1680
	Width int

	// Viewport height in pixels of the browser render. Default is 867
	Height int

	// Scale from 0
	Scale float64

	// Format of image
	Format ImageFormat

	// Quality of image
	Quality int

	// Clip of viewport.
	// All fields is required.
	Clip OptsClip
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
