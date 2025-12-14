package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/ipincamp/radar-jentik-api/pkg/config"
)

// Server struct bertindak sebagai Driving Adapter untuk HTTP
type Server struct {
	app    *fiber.App
	config *config.Config
}

// NewServer menginisialisasi Fiber beserta middleware dasarnya
func NewServer(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		AppName: "Radar Jentik API",
		// Prefork: true, // Bisa diaktifkan nanti untuk Production performance
	})

	// Middleware Standar
	app.Use(recover.New()) // Mencegah crash panic mematikan server
	app.Use(logger.New())  // Logging request masuk
	app.Use(cors.New())    // Mengizinkan akses dari Frontend/Mobile

	server := &Server{
		app:    app,
		config: cfg,
	}

	// Setup Routes
	server.setupRoutes()

	return server
}

// setupRoutes mendaftarkan semua endpoint
func (s *Server) setupRoutes() {
	api := s.app.Group("/api")

	// Health Check Endpoint
	api.Get("/health", s.healthCheck)
}

// healthCheck handler (bisa dipisah ke file handler sendiri jika logika membesar)
func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":      "ok",
		"message":     "Service is running smoothly",
		"environment": s.config.AppEnv,
		"version":     "1.0.0",
	})
}

// Run menjalankan server pada port yang ditentukan di config
func (s *Server) Run() error {
	// Menggunakan AppPort dari godotenv.go (misal: ":3000")
	return s.app.Listen(s.config.AppPort)
}
