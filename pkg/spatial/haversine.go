package spatial

import "math"

const earthRadiusKm = 6371.0

// HaversineDistance menghitung jarak antara dua koordinat (latitude, longitude) dalam Kilometer
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Konversi derajat ke radian
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0

	// Rumus Haversine
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}
