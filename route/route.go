package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2" 
	"github.com/jackc/pgx/v5/pgxpool" 
	"api-students/app/service" 
	"api-students/helper" 
	"api-students/middleware"
)

func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService) {
	api := app.Group("/api/v1")
 	
	api.Get("/health",healthCheck(pool))
	u := api.Group("/students", middleware.RequireJSON)
 	u.Get("/", studentService.List)
 	u.Get("/:id", studentService.Get)
 	u.Post("/", studentService.Create)
 	u.Put("/:id", studentService.Replace)
 	u.Patch("/:id", studentService.Patch)
 	u.Delete("/:id", studentService.Delete)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler { 
	return func(c *fiber.Ctx) error { 
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second) 
		defer cancel() 
		
		if err := pool.Ping(ctx); err != nil { 
			return helper.Fail(c, fiber.StatusServiceUnavailable, 
				"database tidak dapat dihubungi") 
		} 
		return helper.Ok(c, fiber.StatusOK,"server dan database berjalan", nil) }	
}