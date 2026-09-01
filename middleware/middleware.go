package middleware

import (
	"log/slog"
	"strings"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"api-students/helper"
)

func Register(app *fiber.App, logger *slog.Logger){
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New())
	app.Use(RequestLogger(logger))
}

func RequestLogger(logger *slog.Logger) fiber.Handler{
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		requestID := c.Locals("requestid").(string)

		logger.Info("http_request",
			slog.String("request id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		)
		return err
	}
}

var metodeBerbody = map[string]bool{
	fiber.MethodPost: true,
	fiber.MethodPut: true,
	fiber.MethodPatch: true,
}

// requireJSON menolak request berisi body yang Content-Type-nya bukan JSON.
// Status yang tepat untuk kasus ini adalah 415, bukan 400.
func RequireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json")
		 }
 	}
 	return c.Next()
}
