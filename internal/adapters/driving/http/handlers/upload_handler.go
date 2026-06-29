package handlers

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) UploadPhoto(c *fiber.Ctx) error {
	// 1. Tangkap file form-data dengan key "photo" dari aplikasi mobile
	file, err := c.FormFile("photo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal membaca file foto"})
	}

	// 2. Validasi Ekstensi (.jpg / .png)
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Hanya menerima .jpg atau .png"})
	}

	// 3. Rename file menjadi acak agar tidak bentrok
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	savePath := fmt.Sprintf("./public/uploads/%s", fileName)

	// 4. Simpan ke storage lokal server
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan foto ke server"})
	}

	// 5. Kembalikan URL publik ke Flutter (Misal: https://api.anda.com/public/uploads/xxx.jpg)
	fileURL := fmt.Sprintf("%s/public/uploads/%s", c.BaseURL(), fileName)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Upload berhasil",
		"data": fiber.Map{
			"photo_url": fileURL,
		},
	})
}
