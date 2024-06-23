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
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/translator/api/handlers"
	"github.com/translator/api/middlewares"
	"github.com/translator/api/models"
)

var (
	addr    = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	version string
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

	r.HandleFunc("/index", handlers.Index).Methods("GET")
	r.HandleFunc(url+"/translate", middlewares.Chain(handlers.Translate, middlewares.ValidateContentType(), middlewares.ValidateAuthorization())).Methods("POST")

	srv := &http.Server{
		Addr:         *addr,
		Handler:      c.Then(r),
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	go serveHTTP(srv)
	<-quit

	// Gracefully shutdown connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
}

func serveHTTP(srv *http.Server) {
	log.Info().Msgf("translator API Server started at %s", srv.Addr)
	err := srv.ListenAndServe()

	if err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Starting Server listener failed")
	}
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
