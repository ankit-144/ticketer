package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"ticketer/internal/availability"
	"ticketer/internal/booking"
	bookingpostgres "ticketer/internal/booking/postgres"
	"ticketer/internal/catalog"
	catalogpostgres "ticketer/internal/catalog/postgres"
	"ticketer/internal/core/lock"
	"ticketer/internal/pricing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewDB(lc fx.Lifecycle, logger *zap.Logger) (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://ticketer:password@localhost:5432/ticketer?sslmode=disable"
    }
	
	logger.Info("Connecting to database", zap.String("url", dbURL))

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrate: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	logger.Info("Database migrations applied successfully")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing database connection pool")
			pool.Close()
			return nil
		},
	})

	return pool, nil
}

var Module = fx.Module("app",
	// Infrastructure
	fx.Provide(
		zap.NewProduction,
		NewDB,
		fx.Annotate(lock.NewInMemoryLockService, fx.As(new(lock.LockService))),
	),

	// Repositories
	fx.Provide(
		fx.Annotate(catalogpostgres.NewMovieRepository, fx.As(new(catalog.MovieRepository))),
		fx.Annotate(catalogpostgres.NewShowRepository, fx.As(new(catalog.ShowRepository))),
		fx.Annotate(catalogpostgres.NewShowSeatRepository, fx.As(new(catalog.ShowSeatRepository))),
		fx.Annotate(catalogpostgres.NewTheaterRepository, fx.As(new(catalog.TheaterRepository))),
		fx.Annotate(bookingpostgres.NewBookingRepository, fx.As(new(booking.BookingRepository))),
	),

	// Domain Services
	fx.Provide(
		fx.Annotate(availability.New, fx.As(new(availability.Service))),
		fx.Annotate(pricing.New, fx.As(new(pricing.Service))),
		fx.Annotate(booking.NewBookingService, fx.As(new(booking.Service))),
		catalog.NewTheaterService,
		catalog.NewMovieService,
		catalog.NewShowService,
	),

	// HTTP Handlers
	fx.Provide(
		fx.Annotate(
			booking.NewHandler,
			fx.As(new(booking.RouteRegistrar)),
			fx.ResultTags(`group:"routes"`),
		),
		fx.Annotate(
			catalog.NewHandler,
			fx.As(new(booking.RouteRegistrar)),
			fx.ResultTags(`group:"routes"`),
		),
		fx.Annotate(
			availability.NewHandler,
			fx.As(new(booking.RouteRegistrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),

	// HTTP Server
	fx.Provide(NewHTTPServer),
	fx.Invoke(startServer),
)

type ServerParams struct {
	fx.In
	Registrars []booking.RouteRegistrar `group:"routes"`
}

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
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods","GET, POST,PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers","Accept, Content-Type, Content-Length, Authorization")
		if r.Method == "OPTIONS"{
			w.WriteHeader(http.StatusOK)
			return 
		}
		 next.ServeHTTP(w,r)

	})
}
func startServer(lc fx.Lifecycle, mux *http.ServeMux, logger *zap.Logger) {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: enableCORS(mux),
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
