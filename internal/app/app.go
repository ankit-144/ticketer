package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"ticketer/internal/booking"
	bookingmemory "ticketer/internal/booking/memory"
	"ticketer/internal/catalog"
	catalogmemory "ticketer/internal/catalog/memory"
	"ticketer/internal/core/lock"
	"ticketer/internal/pricing"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module is the single FX module that wires every dependency for the
// ticketer service: repositories, services, handlers, and HTTP server.
//
// To swap an implementation (e.g. in-memory → Postgres), change the
// concrete constructor in the relevant fx.Provide block below.
var Module = fx.Module("app",
	// ── Repositories ────────────────────────────────────────────────
	fx.Provide(
		fx.Annotate(catalogmemory.NewMovieRepository, fx.As(new(catalog.MovieRepository))),
		fx.Annotate(catalogmemory.NewShowRepository, fx.As(new(catalog.ShowRepository))),
		fx.Annotate(catalogmemory.NewShowSeatRepository, fx.As(new(catalog.ShowSeatRepository))),
		fx.Annotate(catalogmemory.NewTheaterRepository, fx.As(new(catalog.TheaterRepository))),
		fx.Annotate(bookingmemory.NewBookingRepository, fx.As(new(booking.BookingRepository))),
	),

	// ── Infrastructure ──────────────────────────────────────────────
	fx.Provide(
		fx.Annotate(lock.NewInMemoryLockService, fx.As(new(lock.LockService))),
		zap.NewProduction,
	),

	// ── Domain Services ─────────────────────────────────────────────
	fx.Provide(
		fx.Annotate(pricing.New, fx.As(new(pricing.Service))),
		fx.Annotate(booking.NewBookingService, fx.As(new(booking.Service))),
	),

	// ── HTTP Handlers ───────────────────────────────────────────────
	fx.Provide(
		fx.Annotate(
			booking.NewHandler,
			fx.As(new(booking.RouteRegistrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),

	// ── HTTP Server ─────────────────────────────────────────────────
	fx.Provide(NewHTTPServer),
	fx.Invoke(startServer),
)

// ── Server helpers ──────────────────────────────────────────────────

// ServerParams collects route registrars from all handlers via the
// "routes" FX group tag.
type ServerParams struct {
	fx.In
	Registrars []booking.RouteRegistrar `group:"routes"`
}

// NewHTTPServer builds the mux and registers every handler's routes.
func NewHTTPServer(p ServerParams) *http.ServeMux {
	mux := http.NewServeMux()

	for _, r := range p.Registrars {
		r.RegisterRoutes(mux)
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	return mux
}

// startServer hooks the HTTP server into the FX lifecycle.
func startServer(lc fx.Lifecycle, mux *http.ServeMux, logger *zap.Logger) {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			logger.Info("HTTP server started", zap.String("addr", srv.Addr))
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("HTTP server shutting down")
			return srv.Shutdown(ctx)
		},
	})
}
