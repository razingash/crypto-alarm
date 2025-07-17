package routes

import (
	"crypto-gateway/internal/web/handlers"
	"log"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

var upgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		allowedOrigins := map[string]bool{
			"http://localhost:3000":  true,
			"http://localhost:3001":  true,
			"http://localhost:80":    true,
			"https://localhost:443":  true,
			"http://crypto_frontend": true,
			"http://localhost":       true,
		}
		return allowedOrigins[string(ctx.Request.Header.Peek("Origin"))]
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(ctx *fasthttp.RequestCtx) {
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		handlers.SendRuntimeMetricsWS(conn)
	})
	if err != nil {
		log.Println("Upgrade error:", err)
	}
}

func SetupMetricsRoutes(app *fiber.App) {
	group := app.Group("/api/v1/metrics")

	group.Get("/availability", handlers.GetAvailabilityMetrics)
	group.Get("/info", handlers.GetStaticMetrics)
	group.Get("/detailed", handlers.GetErrorsDetailedInfo)
	group.Get("/basic", handlers.GetErrorsBasicInfo)
	group.Get("/binance-api-weight", handlers.GetBinanceApiWeightMetrics)

	group.Get("/ws", func(c fiber.Ctx) error {
		wsHandler(c.RequestCtx())
		return nil
	})
}
