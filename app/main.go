package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/translator/app/handlers"
	"github.com/translator/app/middlewares"
	"github.com/translator/app/models"
)

var (
	addr        = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	metricsAddr = flag.String("metrics-address", ":2112", "The address to listen on for Prometheus metrics requests.")
	version     string
)

func init() {
	version = os.Getenv("COMMIT_ID")
}

func main() {
	flag.Parse()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// logs
	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("version", version).
		Logger()

	ctx := context.Background()
	ctx = context.WithValue(ctx, models.IDKey{}, uuid.New())

	log.Info().Interface("event_id", ctx.Value(models.IDKey{})).Msg("starting translator API service")

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	r := mux.NewRouter()
	r.Use(middlewares.RecoverPanic)

	c := alice.New(hlog.NewHandler(log), hlog.AccessHandler(accessLogger))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))

	url := "/api/v1"

	r.HandleFunc("/healthcheck", handlers.Healthcheck).Methods("GET")

	r.HandleFunc(url+"/translate", middlewares.Chain(handlers.Translate, middlewares.ValidateContentType(), middlewares.ValidateAuthorization())).Methods("POST")

	srv := &http.Server{
		Addr:         *addr,
		Handler:      c.Then(r),
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	go serveHTTP(ctx, log, srv)
	go serveMetrics(ctx, log, *metricsAddr)
	<-quit

	// Gracefully shutdown connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
}

func accessLogger(r *http.Request, status, size int, dur time.Duration) {
	hlog.FromRequest(r).Info().
		Str("host", r.Host).
		Int("status", status).
		Str("url", r.RequestURI).
		Str("method", r.Method).
		Int("size", size).
		Dur("duration_ms", dur).
		Msg("request")
}

func serveHTTP(ctx context.Context, log zerolog.Logger, srv *http.Server) {
	log.Info().Msgf("translator API Server started at %s", srv.Addr)
	err := srv.ListenAndServe()

	if err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Starting Server listener failed")
	}
}

func serveMetrics(ctx context.Context, log zerolog.Logger, addr string) {
	log.Info().Interface("event_id", ctx.Value(models.IDKey{})).Msgf("translator API Prometheus metrics on port %s", addr)

	http.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Error().Caller().Interface("event_id", ctx.Value(models.IDKey{})).Err(err).Msg("starting Prometheus listener failed")
	}
}
