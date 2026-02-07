package requests

type UpdateDriverStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=OFFLINE ONLINE_AVAILABLE"`
}

type UpdateDriverLocationRequest struct {
	Lat       float64 `json:"lat" binding:"required"`
	Lng       float64 `json:"lng" binding:"required"`
	AccuracyM float64 `json:"accuracy_m" binding:"required"`
}

type NearbyDriversRequest struct {
	Lat     float64 `json:"lat" binding:"required"`
	Lng     float64 `json:"lng" binding:"required"`
	RadiusM float64 `json:"radius_m" binding:"required"`
	Limit   int32   `json:"limit"`
}
