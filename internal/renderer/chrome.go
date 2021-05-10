package renderer

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"golang.org/x/xerrors"
)

type Chrome struct {
	Debug bool
}

func (chrome *Chrome) buildContextOptions() []chromedp.ContextOption {
	opts := []chromedp.ContextOption{}

	if chrome.Debug {
		opts = append(opts, chromedp.WithDebugf(log.Printf))
	}

	return opts
}

func (chrome *Chrome) buildAllocator(ctx context.Context) (context.Context, context.CancelFunc) {
	return chromedp.NewExecAllocator(ctx,
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
		)...,
	)
}

func (chrome *Chrome) Render(
	ctx context.Context,
	url string,
	opts *Opts,
) (io.Reader, error) {
	// ctx, cancel := chrome.buildAllocator(ctx)
	// defer cancel()

	// create context
	ctx, cancel := chromedp.NewContext(
		ctx,
		chrome.buildContextOptions()...,
	)
	defer cancel()

	var actions []chromedp.Action

	// go to url
	actions = append(actions, chromedp.Navigate(url))

	// set size and scale
	actions = append(actions, chromedp.EmulateViewport(
		int64(opts.getWidth()),
		int64(opts.getHeight()),
		chromedp.EmulateScale(opts.getScale()),
	))

	actions = append(actions, chromedp.Sleep(time.Second*5))

	res := []byte{}

	actions = append(actions, captureScreenshot(&res, opts.Format))

	if err := chromedp.Run(ctx, actions...); err != nil {
		return nil, xerrors.Errorf("make screen shot")
	}

	return bytes.NewReader(res), nil
}

func captureScreenshot(res *[]byte, format ImageFormat) chromedp.Action {
	if res == nil {
		panic("res cannot be nil")
	}

	return chromedp.ActionFunc(func(ctx context.Context) error {
		var err error

		call := page.CaptureScreenshot()

		switch format {
		case ImageTypeJPEG:
			call = call.WithFormat(page.CaptureScreenshotFormatJpeg)
		case ImageTypePNG:
			call = call.WithFormat(page.CaptureScreenshotFormatPng)
		}

		*res, err = call.Do(ctx)
		return err
	})
}
