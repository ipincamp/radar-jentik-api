package domain

// SamplePoint merepresentasikan rumah/kontainer yang sudah diinspeksi
type SamplePoint struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Value float64 `json:"value"` // Nilai kerawanan (misal: HI/CI atau skor DBD)
}

// GridPoint merepresentasikan titik-titik pada peta yang nilai kerawanannya diestimasi
type GridPoint struct {
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	EstimatedValue float64 `json:"estimated_value"`
}

// IDWRequest adalah payload yang akan diterima oleh endpoint REST API
type IDWRequest struct {
	MinLat     float64       `json:"min_lat"`
	MinLon     float64       `json:"min_lon"`
	MaxLat     float64       `json:"max_lat"`
	MaxLon     float64       `json:"max_lon"`
	Resolution float64       `json:"resolution"` // Kerapatan grid (misal 0.001 derajat)
	Power      float64       `json:"power"`      // Pangkat IDW (biasanya 2)
	Samples    []SamplePoint `json:"samples"`    // Titik data observasi
}
