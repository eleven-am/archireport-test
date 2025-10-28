package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/eleven-am/enclave/internal/server"
)

// Config captures runtime configuration for the enclave service.
type Config struct {
	DatabasePath string
	ListenAddr   string
}

func provideConfig() Config {
	cfg := Config{
		DatabasePath: os.Getenv("ENCLAVE_DATABASE"),
		ListenAddr:   os.Getenv("ENCLAVE_ADDR"),
	}
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = "enclave.db"
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	return cfg
}

type serverParams struct {
	fx.In

	Config Config
}

func provideServer(p serverParams) (*server.Server, error) {
	return server.New(context.Background(), server.Config{DatabasePath: p.Config.DatabasePath})
}

type lifecycleParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Server    *server.Server
	Config    Config
}

func registerLifecycle(p lifecycleParams) {
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := p.Config.ListenAddr
			log.Printf("starting enclave server on %s", addr)
			go func() {
				if err := p.Server.App.Start(addr); err != nil && err != http.ErrServerClosed {
					log.Printf("server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			var result error
			if err := p.Server.App.Shutdown(shutdownCtx); err != nil {
				result = err
			}
			if err := p.Server.Close(); err != nil && result == nil {
				result = err
			}
			return result
		},
	})
}

// New constructs the fx application responsible for bootstrapping the enclave service.
func New(opts ...fx.Option) *fx.App {
	base := fx.Options(
		fx.Provide(
			provideConfig,
			provideServer,
		),
		fx.Invoke(registerLifecycle),
	)
	all := append([]fx.Option{base}, opts...)
	return fx.New(all...)
}

// Run creates and runs the enclave fx application.
func Run(opts ...fx.Option) error {
	application := New(opts...)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	startCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := application.Start(startCtx); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
	case <-application.Done():
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer stopCancel()
	return application.Stop(stopCtx)
}
