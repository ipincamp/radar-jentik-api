package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driving/http/handlers"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driving/http/middleware"
	"github.com/ipincamp/radar-jentik-api/pkg/config"
	"github.com/ipincamp/radar-jentik-api/pkg/utils"
)

// Server struct bertindak sebagai Driving Adapter untuk HTTP
type Server struct {
	app                     *fiber.App
	config                  *config.Config
	authHandler             *handlers.AuthHandler
	inspectionReportHandler *handlers.InspectionReportHandler
	villageHandler          *handlers.VillageHandler
	idwHandler              *handlers.IDWHandler
	containerTypeHandler    *handlers.ContainerTypeHandler
	uploadHandler           *handlers.UploadHandler
}

// Tambahkan argument baru di NewServer
func NewServer(
	cfg *config.Config,
	authH *handlers.AuthHandler,
	inspectionReportH *handlers.InspectionReportHandler,
	villageH *handlers.VillageHandler,
	idwH *handlers.IDWHandler,
	containerTypeH *handlers.ContainerTypeHandler,
	uploadH *handlers.UploadHandler,
) *Server {
	app := fiber.New(fiber.Config{AppName: "Radar Jentik API"})

	// Middleware Standar
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Buka folder agar foto bisa dilihat lewat URL
	app.Static("/public", "./public")

	server := &Server{
		app:                     app,
		config:                  cfg,
		authHandler:             authH,
		inspectionReportHandler: inspectionReportH,
		villageHandler:          villageH,
		idwHandler:              idwH,
		containerTypeHandler:    containerTypeH,
		uploadHandler:           uploadH,
	}

	// Setup Routes
	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	api := s.app.Group("/api/v1")
	api.Get("/health", s.healthCheck)

	// Auth Routes (Public)
	auth := api.Group("/auth")
	auth.Post("/register", s.authHandler.Register)
	auth.Post("/login", s.authHandler.Login)
	auth.Post("/logout", middleware.Protected(s.config), s.authHandler.Logout)
	auth.Get("/users", middleware.Protected(s.config), s.authHandler.ListUsers)

	users := api.Group("/users", middleware.Protected(s.config))
	users.Post("/", s.authHandler.CreateUser)
	users.Put("/:id", s.authHandler.UpdateUser)
	users.Delete("/:id", s.authHandler.DeleteUser)

	// Master Data (Bisa diakses Kader untuk Dropdown)
	api.Get("/villages", s.villageHandler.GetAll)
	api.Get("/container-types", s.containerTypeHandler.GetActive)

	// Upload API (Protected)
	api.Post("/uploads", middleware.Protected(s.config), s.uploadHandler.UploadPhoto)

	// IDW dan Laporan (Sama seperti sebelumnya)
	api.Post("/estimations/idw", middleware.Protected(s.config), s.idwHandler.Calculate)
	api.Post("/estimations/idw/predict-point", middleware.Protected(s.config), s.idwHandler.PredictSinglePoint)

	reports := api.Group("/reports", middleware.Protected(s.config))
	reports.Post("/", s.inspectionReportHandler.Create)
	reports.Post("/bulk", s.inspectionReportHandler.CreateBulk)
	reports.Get("/history", s.inspectionReportHandler.GetHistory)
	reports.Get("/pending", s.inspectionReportHandler.GetPending)
	reports.Put("/:id/validate", s.inspectionReportHandler.ValidateReport)
	reports.Get("/map", s.inspectionReportHandler.GetMapData)
	reports.Get("/export", s.inspectionReportHandler.ExportExcel)
}

// healthCheck handler untuk memastikan server berjalan dengan baik
func (s *Server) healthCheck(c *fiber.Ctx) error {
	return utils.Success(c, fiber.StatusOK, "Server is running", nil)
}

// Start menjalankan server HTTP
func (s *Server) Start(addr string) error { return s.app.Listen(addr) }

// Shutdown mematikan server secara aman (graceful shutdown)
func (s *Server) Shutdown(ctx context.Context) error { return s.app.ShutdownWithContext(ctx) }
