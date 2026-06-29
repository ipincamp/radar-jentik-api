package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type ContainerTypeHandler struct {
	service ports.ContainerTypeService
}

func NewContainerTypeHandler(service ports.ContainerTypeService) *ContainerTypeHandler {
	return &ContainerTypeHandler{service: service}
}

func (h *ContainerTypeHandler) GetActive(c *fiber.Ctx) error {
	types, err := h.service.GetActiveTypes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data jenis wadah"})
	}
	return c.JSON(fiber.Map{"data": types})
}
