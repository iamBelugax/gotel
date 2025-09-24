package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iamBelugax/gotel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

var (
	serviceName    = "example-app"
	serviceVersion = "1.0.0"
	environment    = "development"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	provider, err := gotel.NewProvider(ctx,
		gotel.WithEndpoint("localhost:4317"),
		gotel.WithResourceAttr("team", "platform"),
		gotel.WithServiceInfo(serviceName, serviceVersion, environment),
	)
	if err != nil {
		log.Fatalf("Failed to initialize OTEL provider: %v", err)
	}
	defer func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown OTEL provider: %v", err)
		}
	}()

	logger := provider.Logger()
	tracer := gotel.NewTracer(provider.Tracer())

	metricRegistry := gotel.NewMetricRegistry(provider.Meter(), "")
	commonMetrics, err := gotel.NewCommonMetrics(metricRegistry)
	if err != nil {
		log.Fatalf("Failed to create common metrics: %v", err)
	}

	customCounter, err := metricRegistry.Counter("custom_operations_total", "Total custom operations")
	if err != nil {
		log.Fatalf("Failed to create custom counter: %v", err)
	}

	app := &Application{
		logger:        logger,
		tracer:        tracer,
		metrics:       commonMetrics,
		customCounter: customCounter,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.handleHome)
	mux.HandleFunc("/users", app.handleUsers)

	httpMiddleware := gotel.NewHTTPMiddleware(serviceName, tracer, commonMetrics)
	handler := httpMiddleware.Handler(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		logger.Info(ctx, "Starting server on port 8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error(ctx, "Server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "Shutting down server...")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error(ctx, "Server shutdown error", zap.Error(err))
	}
}

type Application struct {
	tracer        *gotel.Tracer
	logger        *gotel.ZapLogger
	customCounter metric.Int64Counter
	metrics       *gotel.CommonMetrics
}

func (a *Application) handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a.logger.Info(ctx, "Handling home request")
	fmt.Fprintln(w, "Hello, from the home page!")
}

func (a *Application) handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a.tracer.WithSpan(ctx, "handleUsers", func(ctx context.Context, span *gotel.Span) error {
		span.WithAttributes(attribute.String("user_id", "123"))
		a.logger.Info(ctx, "Handling users request")

		a.customCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("user", "test")))
		time.Sleep(100 * time.Millisecond)
		fmt.Fprintln(w, "Hello, from the users page!")
		return nil
	})
}
