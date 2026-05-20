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
}

// NewServer menginisialisasi Fiber beserta middleware dasarnya
func NewServer(
	cfg *config.Config,
	authH *handlers.AuthHandler,
	inspectionReportH *handlers.InspectionReportHandler,
) *Server {
	app := fiber.New(fiber.Config{
		AppName: "Radar Jentik API",
		// Prefork: true, // Bisa diaktifkan nanti untuk Production performance
	})

	// Middleware Standar
	app.Use(recover.New()) // Mencegah crash panic mematikan server
	app.Use(logger.New())  // Logging request masuk
	app.Use(cors.New())    // Mengizinkan CORS agar bisa diakses oleh Flutter Mobile/Web

	server := &Server{
		app:                     app,
		config:                  cfg,
		authHandler:             authH,
		inspectionReportHandler: inspectionReportH,
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
	auth.Get("/users", middleware.Protected(s.config), s.authHandler.ListUsers) // Manajemen Kader oleh Petugas

	// Inspection Report Routes (Protected - Wajib membawa Bearer Token JWT)
	reports := api.Group("/reports", middleware.Protected(s.config))

	// Rute untuk peran Kader
	reports.Post("/", s.inspectionReportHandler.Create)           // Halaman Form Lapor (Kader)
	reports.Get("/history", s.inspectionReportHandler.GetHistory) // Halaman Riwayat Laporan (Kader)

	// Rute untuk peran Petugas Puskesmas
	reports.Get("/pending", s.inspectionReportHandler.GetPending)          // Halaman Daftar Validasi (Petugas)
	reports.Put("/:id/validate", s.inspectionReportHandler.ValidateReport) // Tombol Aksi Terima/Tolak (Petugas)

	// Rute Gabungan untuk Peta Spasial IDW
	reports.Get("/map", s.inspectionReportHandler.GetMapData) // Halaman Peta Zonasi (Kader & Petugas)
}

// healthCheck handler untuk memastikan server berjalan dengan baik
func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "Radar Jentik API is running smoothly",
	})
}

// Start menjalankan server HTTP pada port yang ditentukan di .env
func (s *Server) Start(addr string) error {
	return s.app.Listen(addr)
}

// Shutdown mematikan server secara aman (graceful shutdown)
func (s *Server) Shutdown(ctx context.Context) error {
	// Memanggil method ShutdownWithContext dari underlying fiber.App
	// Ini akan memastikan server menyelesaikan request yang sedang berjalan
	// sebelum benar-benar mati, sesuai context timeout yang di-set di main.go
	return s.app.ShutdownWithContext(ctx)
}
