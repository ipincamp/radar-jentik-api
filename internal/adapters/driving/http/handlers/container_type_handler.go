package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/ipincamp/radar-jentik-api/pkg/utils"
)

type ContainerTypeHandler struct {
	service ports.ContainerTypeService
}

func NewContainerTypeHandler(service ports.ContainerTypeService) *ContainerTypeHandler {
	return &ContainerTypeHandler{service: service}
}

// Fungsi untuk mendapatkan semua jenis wadah yang aktif
func (h *ContainerTypeHandler) GetActive(c *fiber.Ctx) error {
	types, err := h.service.GetActiveTypes(c.Context())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data jenis wadah", err.Error())
	}
	return utils.Success(c, fiber.StatusOK, "Data jenis wadah berhasil diambil", types)
}
