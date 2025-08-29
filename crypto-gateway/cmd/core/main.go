package main

import (
	"context"
	"crypto-gateway/config"
	"crypto-gateway/internal/appmetrics"
	"crypto-gateway/internal/modules/strategy/service"
	"crypto-gateway/internal/modules/strategy/web"
	"crypto-gateway/internal/web/db"
	"crypto-gateway/internal/web/routes"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	service.StartTime = time.Now().Unix()
	config.LoadConfig()
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:80", "https://localhost:443"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Upgrade", "Connection"},
		ExposeHeaders:    []string{"Upgrade"},
		AllowCredentials: true,
	}))

	app.Use(func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()
		duration := time.Since(start)

		log.Printf("Request to %s %s took %v", c.Method(), c.OriginalURL(), duration)
		return err
	})

	db.InitDB()

	service.Collector = appmetrics.NewLoadMetricsCollector(60)
	service.StController = service.NewBinanceAPIController(5700)
	service.StBinanceApi = service.NewBinanceAPI(service.StController)
	service.StOrchestrator = service.NewBinanceAPIOrchestrator(service.StBinanceApi)
	service.AverageLoadMetrics = service.NewAverageLoadMetricsManager(service.Collector)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service.StOrchestrator.Start(ctx)
	service.SetupInitialSettings(ctx)
	appmetrics.AvailabilityMetricEvent(1, 1, "webserwer UP")

	routes.SetupNotificationRoutes(app)
	web.SetupTriggersRoutes(app)
	routes.SetupSettingsRoutes(app)
	routes.SetupMetricsRoutes(app)
	routes.SetupVariableRoutes(app)
	routes.SetupWorkspaceRoutes(app)
	routes.SetupWorkspaceWidgetRoutes(app)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutdown signal received")

		appmetrics.AvailabilityMetricEvent(1, 0, "webserver DOWN")

		_, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := app.Shutdown(); err != nil {
			appmetrics.AnalyticsServiceLogging(4, "Error during server shutdown", err)
			log.Printf("Error during server shutdown: %v", err)
		}

		cancel()
		os.Exit(0)
	}()

	log.Println("Server started on port 8001")
	if err := app.Listen(":8001"); err != nil {
		appmetrics.AnalyticsServiceLogging(4, "Server encountered an error", err)
		log.Printf("Server encountered an error: %v", err)
	}
}
