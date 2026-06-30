package utils

import "github.com/gofiber/fiber/v2"

// Format Standar JSON Response
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Helper untuk Response Sukses
func Success(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Helper untuk Response Gagal / Error
func Error(c *fiber.Ctx, statusCode int, message string, errors interface{}) error {
	return c.Status(statusCode).JSON(StandardResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// Struktur Metadata Pagination
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	PageSize    int   `json:"page_size"`
	TotalItems  int64 `json:"total_items"`
}

// Struktur Response Khusus Pagination
type PaginatedResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// Helper untuk Response Sukses dengan Pagination
func Paginated(c *fiber.Ctx, statusCode int, message string, data interface{}, meta PaginationMeta) error {
	return c.Status(statusCode).JSON(PaginatedResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}
