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
)

// Server struct bertindak sebagai Driving Adapter untuk HTTP
type Server struct {
	app                     *fiber.App
	config                  *config.Config
	authHandler             *handlers.AuthHandler
	inspectionReportHandler *handlers.InspectionReportHandler
	villageHandler          *handlers.VillageHandler
}

// NewServer menginisialisasi Fiber beserta middleware dasarnya
func NewServer(
	cfg *config.Config,
	authH *handlers.AuthHandler,
	inspectionReportH *handlers.InspectionReportHandler,
	villageH *handlers.VillageHandler,
) *Server {
	app := fiber.New(fiber.Config{
		AppName: "Radar Jentik API",
		// Prefork: true, // Bisa diaktifkan nanti untuk Production performance
	})

	// Middleware Standar
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	server := &Server{
		app:                     app,
		config:                  cfg,
		authHandler:             authH,
		inspectionReportHandler: inspectionReportH,
		villageHandler:          villageH,
	}

	// Setup Routes
	server.setupRoutes()

	return server
}

// setupRoutes mendaftarkan semua endpoint yang diperlukan oleh 11 halaman Flutter
func (s *Server) setupRoutes() {
	api := s.app.Group("/api/v1")

	// Health Check Endpoint
	api.Get("/health", s.healthCheck)

	// Auth Routes (Public)
	auth := api.Group("/auth")
	auth.Post("/register", s.authHandler.Register)
	auth.Post("/login", s.authHandler.Login)
	auth.Post("/logout", middleware.Protected(s.config), s.authHandler.Logout)
	auth.Get("/users", middleware.Protected(s.config), s.authHandler.ListUsers)

	// Master Data Desa (Dibuat public agar bisa diakses saat dropdown registrasi di Flutter)
	api.Get("/villages", s.villageHandler.GetAll)

	// Inspection Report Routes (Protected - Wajib membawa Bearer Token JWT)
	reports := api.Group("/reports", middleware.Protected(s.config))

	// Rute untuk peran Kader
	reports.Post("/", s.inspectionReportHandler.Create)
	reports.Get("/history", s.inspectionReportHandler.GetHistory)

	// Rute untuk peran Petugas Puskesmas
	reports.Get("/pending", s.inspectionReportHandler.GetPending)
	reports.Put("/:id/validate", s.inspectionReportHandler.ValidateReport)

	// Rute Gabungan untuk Peta Spasial IDW
	reports.Get("/map", s.inspectionReportHandler.GetMapData)
}

// healthCheck handler untuk memastikan server berjalan dengan baik
func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "Radar Jentik API is running smoothly",
	})
}

// Start menjalankan server HTTP
func (s *Server) Start(addr string) error {
	return s.app.Listen(addr)
}

// Shutdown mematikan server secara aman (graceful shutdown)
func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
