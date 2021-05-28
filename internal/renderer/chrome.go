package renderer

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/bots-house/webshot/internal"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"golang.org/x/xerrors"
)

type Chrome struct {
	Debug bool

	Resolver ChromeResolver
}

func (chrome *Chrome) buildContextOptions() []chromedp.ContextOption {
	opts := []chromedp.ContextOption{}

	if chrome.Debug {
		opts = append(opts, chromedp.WithDebugf(log.Printf))
	}

	return opts
}

// func (chrome *Chrome) buildAllocator(ctx context.Context) (context.Context, context.CancelFunc) {
// 	return chromedp.NewExecAllocator(ctx,
// 		append(
// 			chromedp.DefaultExecAllocatorOptions[:],
// 			chromedp.Flag("headless", false),
// 		)...,
// 	)
// }

func (chrome *Chrome) Render(
	ctx context.Context,
	url string,
	opts Opts,
) (c []byte, err error) {
	defer func(started time.Time) {
		var ev *zerolog.Event

		if err != nil {
			ev = log.Ctx(ctx).Error().Err(err)
		} else {
			ev = log.Ctx(ctx).Info()
		}

		ev = ev.Str("url", url).
			Int("width", opts.getWidth()).
			Int("height", opts.getHeight()).
			Float64("scale", opts.getScale()).
			Str("format", opts.Format.String()).
			Dur("took", time.Since(started)).
			Dur("delay", opts.Delay).
			Bool("full_page", opts.FullPage).
			Bool("scroll_page", opts.ScrollPage)

		if opts.Clip.IsSet() {
			ev = ev.
				Float64("clip_x", *opts.Clip.X).
				Float64("clip_y", *opts.Clip.Y).
				Float64("clip_width", *opts.Clip.Width).
				Float64("clip_height", *opts.Clip.Height)
		}

		ev.Msg("screenshot")

	}(time.Now())

	if chrome.Resolver != nil {
		wsurl, err := chrome.Resolver.BrowserWebSocketURL(ctx)
		if err != nil {
			return nil, xerrors.Errorf("resolve remote browser: %w", err)
		}

		log.Ctx(ctx).Debug().Str("url", wsurl).Msg("use remote browser")

		var cancel context.CancelFunc
		ctx, cancel = chromedp.NewRemoteAllocator(ctx, wsurl)
		defer cancel()
	} else {
		log.Ctx(ctx).Debug().Msg("use embedded browser")
	}

	// create context
	ctx, cancel := chromedp.NewContext(
		ctx,
		chrome.buildContextOptions()...,
	)
	defer cancel()

	var actions []chromedp.Action

	// go to url
	actions = append(actions, logAction(ctx,
		"navigate", logFields{
			"url": url,
		},
		chromedp.Navigate(url),
	))

	// set size and scale
	actions = append(actions, logAction(ctx,
		"emulate viewport", logFields{
			"width":  opts.getWidth(),
			"height": opts.getHeight(),
			"scale":  opts.getScale(),
		},

		chromedp.EmulateViewport(
			int64(opts.getWidth()),
			int64(opts.getHeight()),
			chromedp.EmulateScale(opts.getScale()),
		),
	))

	if opts.ScrollPage {
		actions = append(actions,
			logAction(ctx,
				"scroll to bottom",
				nil,
				chromedp.Evaluate(`
					window.scrollTo(0, document.body.scrollHeight);
				`, nil),
			),
			logAction(ctx,
				"delay",
				logFields{"v": opts.Delay.String()},
				chromedp.Sleep(opts.Delay),
			),
			logAction(ctx,
				"scroll to top",
				nil,
				chromedp.Evaluate(`
					window.scrollTo(0, 0);
				`, nil),
			),
		)
	}

	if opts.Delay != 0 && !opts.ScrollPage {
		actions = append(actions, logAction(ctx,
			"delay",
			logFields{"v": opts.Delay.String()},
			chromedp.Sleep(opts.Delay),
		))
	}

	var res []byte

	if opts.FullPage {
		actions = append(actions, logAction(ctx,
			"full page screenshot",
			nil,
			captureScreenshotFull(&res, &opts),
		))
	} else {
		actions = append(actions, logAction(ctx,
			"screenshot",
			nil,
			captureScreenshot(&res, &opts),
		))
	}

	if err := chromedp.Run(ctx, actions...); err != nil {
		return nil, xerrors.Errorf("make screen shot: %w", err)
	}

	return res, nil
}

type logFields map[string]interface{}

func logAction(ctx context.Context, name string, fields logFields, action chromedp.Action) chromedp.Action {
	return chromedp.ActionFunc(func(c context.Context) (err error) {
		defer func(started time.Time) {
			var ev *zerolog.Event

			if err != nil {
				ev = log.Ctx(ctx).Error().Err(err)
			} else {
				ev = log.Ctx(ctx).Debug().Fields(fields)
			}

			ev.Dur("took", time.Since(started)).Msg(fmt.Sprintf("do %s", name))
		}(time.Now())

		return action.Do(c)
	})
}

// based on chromedp.FullScrenshot
func captureScreenshotFull(res *[]byte, opts *Opts) chromedp.Action {
	if res == nil {
		panic("res cannot be nil")
	}
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// get layout metrics
		_, _, contentSize, _, _, cssContentSize, err := page.GetLayoutMetrics().Do(ctx)
		if err != nil {
			return err
		}
		// protocol v90 changed the return parameter name (contentSize -> cssContentSize)
		if cssContentSize != nil {
			contentSize = cssContentSize
		}
		width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

		if opts.hasWidth() {
			width = int64(opts.getWidth())
		}

		// force viewport emulation
		err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
			WithScreenOrientation(&emulation.ScreenOrientation{
				Type:  emulation.OrientationTypePortraitPrimary,
				Angle: 0,
			}).
			Do(ctx)
		if err != nil {
			return err
		}
		// capture screenshot
		*res, err = page.CaptureScreenshot().
			WithQuality(int64(opts.Quality)).
			WithClip(&page.Viewport{
				X:      contentSize.X,
				Y:      contentSize.Y,
				Width:  contentSize.Width,
				Height: contentSize.Height,
				Scale:  opts.getScale(),
			}).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}

func captureScreenshot(res *[]byte, opts *Opts) chromedp.Action {
	if res == nil {
		panic("res cannot be nil")
	}

	return chromedp.ActionFunc(func(ctx context.Context) error {
		var err error

		call := page.CaptureScreenshot()

		switch opts.Format {
		case internal.ImageFormatJPEG:
			call = call.WithFormat(page.CaptureScreenshotFormatJpeg)
		case internal.ImageFormatPNG:
			call = call.WithFormat(page.CaptureScreenshotFormatPng)
		}

		call = call.WithQuality(int64(opts.Quality))

		if opts.Clip.IsSet() {
			call = call.WithClip(&page.Viewport{
				X:      *opts.Clip.X,
				Y:      *opts.Clip.Y,
				Width:  *opts.Clip.Width,
				Height: *opts.Clip.Height,
				Scale:  1.0,
			})
		}

		*res, err = call.Do(ctx)
		return err
	})
}
