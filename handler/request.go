package handler

type RequestBody struct {
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"long"`
	MaxDistance float64 `json:"max_distance"`
}
