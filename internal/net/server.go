package net

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type Service interface {
	http.Handler

	MountPath() string
}

type Server struct {
	http *http.Server
}

type ServerFlags struct {
	Host   string
	Port   uint
	Prefix string
}

func NewServer(flags *ServerFlags, services ...Service) *Server {
	mux := http.NewServeMux()
	setupControllers(mux, flags.Prefix, services...)

	return &Server{
		http: &http.Server{
			Addr:         flags.Host + ":" + fmt.Sprintf("%d", flags.Port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		},
	}
}

func (s *Server) Run(certPath, keyPath string) error {
	sCtx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	eg, egCtx := errgroup.WithContext(sCtx)

	eg.Go(func() error {
		slog.Info(fmt.Sprintf("App server starting on %s", s.http.Addr))

		if certPath != "" || keyPath != "" {
			return s.http.ListenAndServeTLS(certPath, keyPath)
		} else {
			return s.http.ListenAndServe()
		}
	})

	eg.Go(func() error {
		<-egCtx.Done()
		// if context.Background is "Done" or the timeout is exceeded, it'll cause an immediate shutdown
		return s.Shutdown(context.Background(), 20*time.Second) // no idea how much timeout is needed
	})

	return eg.Wait()
}

func (s *Server) Shutdown(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := s.http.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error(fmt.Errorf("server shutdown: %w", err).Error())
		return err
	}

	return nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("project-racer"))
}

func setupControllers(mux *http.ServeMux, prefix string, services ...Service) {
	base := "/" + prefix + "/"
	mux.HandleFunc(base, homeHandler)
	for _, service := range services {
		path := base + service.MountPath()
		mux.Handle(path+"/", http.StripPrefix(path, service))
	}
}
