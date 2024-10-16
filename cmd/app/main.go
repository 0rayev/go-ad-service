package main

import (
	"ad_service/internal/ad"
	"ad_service/internal/config"
	"ad_service/internal/database"
	"ad_service/pkg/metrics"
	"ad_service/pkg/middleware"
	"ad_service/pkg/tracing"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration using viper
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	// Connect to the database using loaded config
	db, err := database.Connect(cfg.MySQL)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	// Initialize repository, service, and handler
	repo := ad.Repository{DB: db}
	service := &ad.AdService{Repo: &repo}
	handler := ad.NewHandler(service)

	// Initialize Prometheus metrics
	metrics.InitMetrics()

	// Initialize OpenTelemetry tracing
	cleanup := tracing.InitTracer()
	defer cleanup()

	// Set up Gin router
	r := gin.Default()

	// Metrics endpoint for Prometheus
	r.GET("/metrics", gin.WrapH(metrics.PrometheusHandler()))

	// Add middleware to track Prometheus metrics for every request
	r.Use(metrics.MetricsMiddlewareGin())

	// API Endpoints
	r.POST("/ads", handler.AddAd)
	r.GET("/ads", handler.GetAllAds)
	r.GET("/ads/:id", handler.GetAdByID)
	r.PUT("/ads/:id", handler.UpdateAd)
	r.DELETE("/ads/:id", handler.DeleteAd)

	// Configure the HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	//GracefulShutdown
	middleware.GracefulShutdown(srv)
}
