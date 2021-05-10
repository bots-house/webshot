package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bots-house/webshot/internal/api"
	"github.com/bots-house/webshot/internal/renderer"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	HTTP struct {
		Addr string `long:"addr" description:"http addr to listen" env:"ADDR" default:":8000"`
	} `group:"HTTP" namespace:"http" env-namespace:"HTTP"`
}

func loadConfig() Config {
	config := Config{}

	parser := flags.NewParser(&config, flags.Default)
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			fmt.Print(err)
			os.Exit(1)
		default:
			fmt.Print(err)
			os.Exit(1)
		}
	}

	return config
}

var (
	buildVersion = "unknown"
	buildRef     = "unknown"
	buildTime    = "unknown"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config := loadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx, cancel = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Ctx(ctx).Info().
		Dict("build", zerolog.Dict().
			Str("version", buildVersion).
			Str("ref", buildRef).
			Str("time", buildTime),
		).
		Msg("start webshot")

	if err := run(ctx, config); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		defer os.Exit(2)
	}
}

func run(ctx context.Context, config Config) error {
	log.Info().Str("addr", config.HTTP.Addr).Msg("listen http...")

	rndr := &renderer.Chrome{Debug: false}

	api := api.New(api.Deps{Renderer: rndr})

	server := newServer(
		config.HTTP.Addr,
		api,
	)

	go func() {
		<-ctx.Done()

		log.Warn().Msg("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			time.Second*10,
		)

		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Warn().Err(err).Msg("shutdown error")
		}
	}()

	if err := server.ListenAndServe(); err == http.ErrServerClosed {
		return nil
	} else {
		return err
	}
}

func newServer(addr string, handler http.Handler) *http.Server {
	baseCtx := context.Background()
	baseCtx = log.Logger.WithContext(baseCtx)

	return &http.Server{
		Addr:    addr,
		Handler: handler,
		BaseContext: func(_ net.Listener) context.Context {
			return baseCtx
		},
	}
}
