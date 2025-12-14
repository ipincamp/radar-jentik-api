package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type AreaHandler struct {
	service ports.AreaService
}

func NewAreaHandler(s ports.AreaService) *AreaHandler {
	return &AreaHandler{service: s}
}

// Struct untuk GeoJSON Response standar
type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   json.RawMessage        `json:"geometry"` // RawMessage agar tidak di-escape menjadi string
}

type FeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

func (h *AreaHandler) GetAll(c *fiber.Ctx) error {
	areas, err := h.service.GetAllAreas(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Transformasi ke GeoJSON FeatureCollection
	features := make([]GeoJSONFeature, 0)

	for _, area := range areas {
		// Validasi: Abaikan area yang tidak punya geometri
		if area.GeoJSON == "" {
			continue
		}

		feature := GeoJSONFeature{
			Type: "Feature",
			ID:   area.ID,
			Properties: map[string]interface{}{
				"name": area.Name,
				"type": area.Type,
				// Tambahkan properti lain jika perlu (misal: kode_wilayah)
			},
			// Konversi string GeoJSON dari DB menjadi Raw JSON Object
			Geometry: json.RawMessage(area.GeoJSON),
		}
		features = append(features, feature)
	}

	response := FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	return c.JSON(response)
}
