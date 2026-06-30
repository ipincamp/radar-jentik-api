package handlers

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ipincamp/radar-jentik-api/pkg/utils"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// Fungsi UploadPhoto menangani upload foto dari aplikasi mobile.
func (h *UploadHandler) UploadPhoto(c *fiber.Ctx) error {
	// 1. Tangkap file form-data dengan key "photo" dari aplikasi mobile
	file, err := c.FormFile("photo")
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal membaca file foto", err.Error())
	}

	// 2. Validasi Ekstensi (.jpg / .png)
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return utils.Error(c, fiber.StatusBadRequest, "Hanya menerima .jpg atau .png", "")
	}

	// 3. Rename file menjadi acak agar tidak bentrok
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	savePath := fmt.Sprintf("./public/uploads/%s", fileName)

	// 4. Simpan ke storage lokal server
	if err := c.SaveFile(file, savePath); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal menyimpan foto ke server", err.Error())
	}

	// 5. Kembalikan URL publik ke Flutter (Misal: https://api.anda.com/public/uploads/xxx.jpg)
	fileURL := fmt.Sprintf("%s/public/uploads/%s", c.BaseURL(), fileName)

	return utils.Success(c, fiber.StatusOK, "Upload berhasil", fiber.Map{
		"photo_url": fileURL,
	})
}
