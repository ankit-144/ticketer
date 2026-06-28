package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"ticketer/internal/auth"
	authpostgres "ticketer/internal/auth/postgres"
	"ticketer/internal/availability"
	"ticketer/internal/booking"
	bookingpostgres "ticketer/internal/booking/postgres"
	"ticketer/internal/catalog"
	catalogpostgres "ticketer/internal/catalog/postgres"
	"ticketer/internal/core/database"
	"ticketer/internal/core/lock"
	"ticketer/internal/events"
	"ticketer/internal/pricing"
	"go.uber.org/fx"
	"go.uber.org/zap"
)


var Module = fx.Module("app",
	// Infrastructure
	fx.Provide(
		zap.NewProduction,
		database.NewDB,
		fx.Annotate(lock.NewInMemoryLockService, fx.As(new(lock.LockService))),
		func(logger *zap.Logger) (events.EventPublisher, error) {
			uri := os.Getenv("RABBITMQ_URI")
			if uri == "" {
				uri = "amqp://guest:guest@localhost:5672/"
			}
			pub, err := events.NewRabbitMQPublisher(uri, "notification.exchange", "notification.routing.key")
			if err != nil {
				logger.Warn("Failed to connect to RabbitMQ, using nil publisher", zap.Error(err))
				return nil, nil // Return nil so app starts even if RMQ is down (BookingService checks for nil)
			}
			return pub, nil
		},
	),

	// Repositories
	fx.Provide(
		fx.Annotate(catalogpostgres.NewMovieRepository, fx.As(new(catalog.MovieRepository))),
		fx.Annotate(catalogpostgres.NewShowRepository, fx.As(new(catalog.ShowRepository))),
		fx.Annotate(catalogpostgres.NewShowSeatRepository, fx.As(new(catalog.ShowSeatRepository))),
		fx.Annotate(catalogpostgres.NewTheaterRepository, fx.As(new(catalog.TheaterRepository))),
		fx.Annotate(bookingpostgres.NewBookingRepository, fx.As(new(booking.BookingRepository))),
		fx.Annotate(authpostgres.NewUserRepository, fx.As(new(auth.UserRepository))),
	),

	// Domain Services
	fx.Provide(
		fx.Annotate(availability.New, fx.As(new(availability.Service))),
		fx.Annotate(pricing.New, fx.As(new(pricing.Service))),
		fx.Annotate(booking.NewBookingService, fx.As(new(booking.Service))),
		catalog.NewTheaterService,
		catalog.NewMovieService,
		catalog.NewShowService,
		fx.Annotate(auth.NewService, fx.As(new(auth.Service))),
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
		fx.Annotate(
			auth.NewHandler,
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
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}
		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS"{
			w.WriteHeader(http.StatusOK)
			return 
		}
		 next.ServeHTTP(w,r)

	})
}
func startServer(lc fx.Lifecycle, mux *http.ServeMux, logger *zap.Logger) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    ":" + port,
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
