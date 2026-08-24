package handlers

import (
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/ipincamp/radar-jentik-api/pkg/utils"
)

type MalariaSurveyHandler struct {
	service ports.MalariaSurveyService
}

func NewMalariaSurveyHandler(service ports.MalariaSurveyService) *MalariaSurveyHandler {
	return &MalariaSurveyHandler{service: service}
}

// DTO untuk validasi dan penangkapan request
type CreateMalariaSurveyRequest struct {
	VillageID         string  `json:"village_id"`
	Dusun             string  `json:"dusun"`
	RT                string  `json:"rt"`
	RW                string  `json:"rw"`
	Latitude          float64 `json:"latitude" validate:"required"`
	Longitude         float64 `json:"longitude" validate:"required"`
	BreedingPlaceType string  `json:"breeding_place_type"`
	PhysicalLighting  string  `json:"physical_lighting"`
	PhysicalWaterFlow string  `json:"physical_water_flow"`
	BioPlants         string  `json:"bio_plants"`
	BioAnimals        string  `json:"bio_animals"`
	ChemSalinity      float64 `json:"chem_salinity"`
	ChemPH            float64 `json:"chem_ph"`
	ChemWaterTemp     float64 `json:"chem_water_temp"`
	Area              float64 `json:"area"`
	Depth             float64 `json:"depth"`
	IsLarvaeFound     bool    `json:"is_larvae_found"`
	LarvaeSpecies     string  `json:"larvae_species"`
	ScoopCount        int     `json:"scoop_count"`
	LarvaeCount       int     `json:"larvae_count"`
	InspectedAt       string  `json:"inspected_at"` // Format ISO 8601 dari aplikasi
}

func (h *MalariaSurveyHandler) Create(c *fiber.Ctx) error {
	req := new(CreateMalariaSurveyRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	// Mengambil ID User (Kader) dari JWT Token Middleware
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "Sesi tidak valid, harap login ulang", "")
	}

	// Mapping dari DTO ke Domain Entity
	survey := &domain.MalariaSurvey{
		UserID:            userID,
		VillageID:         req.VillageID,
		Dusun:             req.Dusun,
		RT:                req.RT,
		RW:                req.RW,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		BreedingPlaceType: req.BreedingPlaceType,
		PhysicalLighting:  req.PhysicalLighting,
		PhysicalWaterFlow: req.PhysicalWaterFlow,
		BioPlants:         req.BioPlants,
		BioAnimals:        req.BioAnimals,
		ChemSalinity:      req.ChemSalinity,
		ChemPH:            req.ChemPH,
		ChemWaterTemp:     req.ChemWaterTemp,
		Area:              req.Area,
		Depth:             req.Depth,
		IsLarvaeFound:     req.IsLarvaeFound,
		LarvaeSpecies:     req.LarvaeSpecies,
		ScoopCount:        req.ScoopCount,
		LarvaeCount:       req.LarvaeCount,
	}

	// Jika Flutter mengirimkan waktu khusus, kita gunakan waktu tersebut
	if req.InspectedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.InspectedAt)
		if err == nil {
			survey.InspectedAt = parsedTime
		}
	}

	// Eksekusi Service
	if err := h.service.CreateSurvey(c.Context(), survey); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal menyimpan survei malaria", err.Error())
	}

	return utils.Success(c, fiber.StatusCreated, "Survei malaria berhasil disimpan", nil)
}

func (h *MalariaSurveyHandler) GetHistory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "Sesi tidak valid", "")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	surveys, totalData, err := h.service.GetPaginatedHistory(c.Context(), userID, page, limit)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil riwayat", err.Error())
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(limit)))
	meta := utils.PaginationMeta{
		CurrentPage: page, PageSize: limit, TotalItems: totalData, TotalPages: totalPages,
	}

	return utils.Paginated(c, fiber.StatusOK, "Riwayat malaria berhasil diambil", surveys, meta)
}

func (h *MalariaSurveyHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	req := new(CreateMalariaSurveyRequest) // Gunakan ulang DTO Create

	if err := c.BodyParser(req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	survey := &domain.MalariaSurvey{
		VillageID:         req.VillageID,
		Dusun:             req.Dusun,
		RT:                req.RT,
		RW:                req.RW,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		BreedingPlaceType: req.BreedingPlaceType,
		PhysicalLighting:  req.PhysicalLighting,
		PhysicalWaterFlow: req.PhysicalWaterFlow,
		BioPlants:         req.BioPlants,
		BioAnimals:        req.BioAnimals,
		ChemSalinity:      req.ChemSalinity,
		ChemPH:            req.ChemPH,
		ChemWaterTemp:     req.ChemWaterTemp,
		Area:              req.Area,
		Depth:             req.Depth,
		IsLarvaeFound:     req.IsLarvaeFound,
		LarvaeSpecies:     req.LarvaeSpecies,
		ScoopCount:        req.ScoopCount,
		LarvaeCount:       req.LarvaeCount,
	}

	if req.InspectedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.InspectedAt)
		if err == nil {
			survey.InspectedAt = parsedTime
		}
	}

	if err := h.service.UpdateSurvey(c.Context(), id, survey); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal memperbarui survei malaria", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, "Survei malaria berhasil diperbarui", nil)
}

func (h *MalariaSurveyHandler) ExportExcel(c *fiber.Ctx) error {
	userID, ok1 := c.Locals("user_id").(string)
	role, ok2 := c.Locals("role").(string)
	if !ok1 || !ok2 {
		return utils.Error(c, fiber.StatusUnauthorized, "Sesi tidak valid", "")
	}

	fileBytes, fileName, err := h.service.ExportToExcel(c.Context(), userID, role)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengekspor data ke Excel", err.Error())
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	return c.Send(fileBytes)
}
