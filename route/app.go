package config

import (
	"praktikummongo/app/repository"
	"praktikummongo/app/service"
	"praktikummongo/middleware"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewApp(db *mongo.Database) *fiber.App {
	app := fiber.New()

	// Repository
	userRepo := repository.NewUserRepository(db)
	alumniRepo := repository.NewAlumniRepository(db)
	pekerjaanRepo := repository.NewPekerjaanRepository(db)

	// Service
	authService := service.NewAuthService(userRepo)
	alumniService := service.NewAlumniService(alumniRepo, db)
	pekerjaanService := service.NewPekerjaanService(pekerjaanRepo)

	// ------------------- ROUTE SETUP -------------------

	// Auth
	app.Post("/login", authService.Login)
	app.Post("/register", authService.Register)

	api := app.Group("/api")

	// ------------------- ALUMNI -------------------
	alumni := api.Group("/alumni", middleware.JWTMiddleware)
	alumni.Get("/", middleware.RoleMiddleware("admin", "user"), alumniService.GetAll)
	alumni.Get("/:id", middleware.RoleMiddleware("admin", "user"), alumniService.GetByID)
	alumni.Post("/", middleware.RoleMiddleware("admin"), alumniService.Create)
	alumni.Put("/:id", middleware.RoleMiddleware("admin"), alumniService.Update)
	alumni.Delete("/:id", middleware.RoleMiddleware("admin"), alumniService.Delete)

	alumni.Get("/jumlah-angkatan", middleware.RoleMiddleware("admin", "user"), alumniService.GetJumlahByAngkatan)
	alumni.Get("/jumlah-pekerjaan", middleware.RoleMiddleware("admin", "user"), alumniService.GetAlumniDenganDuaPekerjaan)

	// ------------------- PEKERJAAN -------------------
	pekerjaan := api.Group("/pekerjaan", middleware.JWTMiddleware)
	pekerjaan.Get("/", middleware.RoleMiddleware("admin", "user"), pekerjaanService.GetAll)
	pekerjaan.Get("/:id", middleware.RoleMiddleware("admin", "user"), pekerjaanService.GetByID)
	pekerjaan.Post("/", middleware.RoleMiddleware("admin", "user"), pekerjaanService.Create)
	pekerjaan.Put("/:id", middleware.RoleMiddleware("admin", "user"), pekerjaanService.Update)
	pekerjaan.Delete("/:id", middleware.RoleMiddleware("admin", "user"), pekerjaanService.Delete)
	pekerjaan.Get("/trash", middleware.RoleMiddleware("admin", "user"), pekerjaanService.GetTrash)
	pekerjaan.Put("/restore/:id", middleware.RoleMiddleware("admin", "user"), pekerjaanService.Restore)
	pekerjaan.Delete("/hard/:id", middleware.RoleMiddleware("admin", "user"), pekerjaanService.HardDelete)

	return app
}
