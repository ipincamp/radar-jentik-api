package domain

// SamplePoint merepresentasikan entitas titik observasi lapangan untuk kalkulasi spasial
type SamplePoint struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Value float64 `json:"value"` // Bobot nilai kerawanan (0.0 atau 100.0)
}

// GridPoint merepresentasikan sel matriks hasil interpolasi IDW
type GridPoint struct {
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	EstimatedValue float64 `json:"estimated_value"`
}

// IDWRequest merepresentasikan parameter area pembatas (bounding box) dari client
type IDWRequest struct {
	MinLat     float64       `json:"min_lat"`
	MinLon     float64       `json:"min_lon"`
	MaxLat     float64       `json:"max_lat"`
	MaxLon     float64       `json:"max_lon"`
	Resolution float64       `json:"resolution"` // Tingkat kerapatan peta (Contoh: 0.001)
	Power      float64       `json:"power"`      // Pangkat matematis IDW (Default: 2.0)
	Samples    []SamplePoint `json:"samples"`
}

type IDWPointRequest struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
