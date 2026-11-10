package route

import (
	"context"
	"time"
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	MongoClient       *mongo.Client
	JWT               *helper.JWTManager
	Permissions       *helper.PermissionSet
	StudentService    *service.StudentService
	AuthService       *service.AuthService
	PortfolioService  *service.PortfolioService
	CourseService     *service.CourseService
	NotificationService *service.NotificationService
	SupervisionService *service.SupervisionService

}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool, deps.MongoClient))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)
	
	students := api.Group("/students",
		middleware.RequireJSON, middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions

	// Hak dapat diputuskan tanpa melihat data -> middleware.
	students.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	students.Post("/", middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)

	// Hak bergantung pada kepemilikan data -> diperiksa di service.
	students.Get("/:id", deps.StudentService.Get)
	students.Get("/:id/prestasi", deps.StudentService.GetWithPrestation)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)

	students.Get("/:id/portfolio", deps.PortfolioService.Get)
	students.Put("/:id/portfolio", deps.PortfolioService.Put)
	students.Post("/:id/portfolio/projects", deps.PortfolioService.AddProject)
	students.Delete("/:id/portfolio/projects/:projectId", deps.PortfolioService.DeleteProject)
	
	portfolios := api.Group("/portfolios",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	portfolios.Get("/stats/technologies", deps.PortfolioService.TopTechnologies)

	// --- kelas dan KRS --- 
	courses := api.Group("/courses", 
		middleware.RequireJSON, 
		middleware.RequireAuth(deps.JWT)) 
	// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai 
	// urutan pendaftaran; bila dibalik, "mine" akan dianggap sebagai :id. 
	courses.Get("/", deps.CourseService.List) 
	courses.Get("/mine", deps.CourseService.Mine) 
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), deps.CourseService.UpdateQuota)
	
	enrollments := api.Group("/enrollments", 
		middleware.RequireJSON, 
		middleware.RequireAuth(deps.JWT)) 
	enrollments.Post("/swap", deps.CourseService.Swap)

	supervisors := api.Group("/supervisors",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	supervisors.Post("/:id/assign", middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisionService.Assign)
	supervisors.Patch("/:id", middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisionService.UpdateQuota)

	supervisions := api.Group("/supervisions",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	supervisions.Post("/switch", middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisionService.Switch)

	notifications := api.Group("/notifications", 
		middleware.RequireJSON, 
		middleware.RequireAuth(deps.JWT)) 
	notifications.Get("/mine", deps.NotificationService.Mine) 
	
	admin := api.Group("/admin", 
		middleware.RequireJSON, 
		middleware.RequireAuth(deps.JWT)) 
	admin.Get("/queue", middleware.RequirePermission(perms, "queue:monitor"), deps.NotificationService.Monitor)
}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}
		if err :=mongoClient.Ping(ctx, nil); err != nil {
			return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
