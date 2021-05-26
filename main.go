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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/bots-house/webshot/internal/api"
	"github.com/bots-house/webshot/internal/renderer"
	"github.com/bots-house/webshot/internal/service"
	"github.com/bots-house/webshot/internal/storage"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
)

type Config struct {
	HTTP struct {
		Addr string `long:"addr" description:"http addr to listen" env:"ADDR" default:":8000"`
	} `group:"HTTP" namespace:"http" env-namespace:"HTTP"`

	Browser struct {
		Addr string `long:"addr" description:"remote browser connection string. Allowed is ws://... or http://" env:"ADDR"`
	} `group:"Browser" namespace:"browser" env-namespace:"BROWSER"`

	Storage struct {
		S3 struct {
			Key      string `long:"key" description:"s3 key" env:"KEY"`
			Secret   string `long:"secret" description:"s3 secret" env:"SECRET"`
			Region   string `long:"region" description:"s3 region" env:"REGION"`
			Bucket   string `long:"bucket" description:"s3 bucket" env:"BUCKET"`
			Endpoint string `long:"endpoint" description:"s3 endpoint" env:"ENDPOINT"`
			Subdir   string `long:"subdir" description:"s3 bucket subdir" env:"SUBDIR"`
		} `group:"S3" namespace:"s3" env-namespace:"S3"`
	} `group:"Storage" namespace:"storage" env-namespace:"STORAGE"`

	Log struct {
		Pretty bool `long:"pretty" description:"enable pretty logging" env:"PRETTY"`
		Debug  bool `long:"debug" description:"enable debug level" env:"DEBUG"`
	} `group:"Logging" namespace:"log" env-namespace:"LOG"`
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
	config := loadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = withLogger(ctx, config)

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

	storage, err := newStorage(ctx, config)
	if err != nil {
		return xerrors.Errorf("new storage: %w", err)
	}

	renderer, err := newRenderer(ctx, config)
	if err != nil {
		return xerrors.Errorf("new renderer: %w", err)
	}

	srv := &service.Service{
		Renderer: renderer,
		Storage:  storage,
	}

	api := api.New(srv)
	return listenAndServe(ctx, config.HTTP.Addr, api)
}

func listenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		<-ctx.Done()

		log.Ctx(ctx).Warn().Msg("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			time.Second*10,
		)

		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("shutdown error")
		}
	}()

	log.Ctx(ctx).Info().Str("addr", addr).Msg("listen http...")
	if err := server.ListenAndServe(); err == http.ErrServerClosed {
		return nil
	} else {
		return err
	}
}

func withLogger(ctx context.Context, cfg Config) context.Context {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if cfg.Log.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if cfg.Log.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return log.Logger.WithContext(ctx)
}

func newRenderer(ctx context.Context, cfg Config) (renderer.Renderer, error) {
	var resolver renderer.ChromeResolver

	if cfg.Browser.Addr != "" {
		log.Ctx(ctx).Info().Str("addr", cfg.Browser.Addr).Msg("init remote chrome renderer")

		var err error
		resolver, err = renderer.NewChromeResolver(cfg.Browser.Addr)
		if err != nil {
			return nil, xerrors.Errorf("new chrome resolver '%s': %w", cfg.Browser.Addr, err)
		}
	} else {
		log.Ctx(ctx).Info().Msg("init local chrome renderer")
	}

	return &renderer.Chrome{Resolver: resolver}, nil
}

func newStorage(ctx context.Context, cfg Config) (storage.Storage, error) {
	if cfg.Storage.S3.Key == "" {
		return nil, nil
	}

	log.Info().
		Str("endpoint", cfg.Storage.S3.Endpoint).
		Str("region", cfg.Storage.S3.Region).
		Str("bucket", cfg.Storage.S3.Bucket).
		Msg("init s3 storage")

	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			cfg.Storage.S3.Key,
			cfg.Storage.S3.Secret,
			"",
		),
		Endpoint: aws.String(cfg.Storage.S3.Endpoint),
		Region:   aws.String(cfg.Storage.S3.Region),
	}

	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, xerrors.Errorf("aws new session: %w", err)
	}

	return storage.NewS3(awsSession, cfg.Storage.S3.Bucket, cfg.Storage.S3.Subdir), nil
}
