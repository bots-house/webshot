package internal

import (
	"fmt"
	"strings"
)

type ImageFormat int8

const (
	ImageFormatPNG ImageFormat = iota
	ImageFormatJPEG

	pngFormat  = "png"
	jpegFormat = "jpeg"
)

func ParseImageType(v string) (ImageFormat, error) {
	switch strings.ToLower(v) {
	case pngFormat:
		return ImageFormatPNG, nil
	case jpegFormat:
		return ImageFormatJPEG, nil
	default:
		return ImageFormat(-1), fmt.Errorf("unsupported image type: %s", v)
	}
}

func (it *ImageFormat) UnmarshalText(text []byte) (err error) {
	*it, err = ParseImageType(string(text))
	return
}

func (it ImageFormat) Ext() string {
	switch it {
	case ImageFormatPNG:
		return pngFormat
	case ImageFormatJPEG:
		return jpegFormat
	default:
		return "bin"
	}
}

func (it ImageFormat) ContentType() string {
	switch it {
	case ImageFormatPNG:
		return "image/png"
	case ImageFormatJPEG:
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func (it ImageFormat) String() string {
	switch it {
	case ImageFormatJPEG:
		return jpegFormat
	case ImageFormatPNG:
		return pngFormat
	default:
		return "unknown"
	}
}
