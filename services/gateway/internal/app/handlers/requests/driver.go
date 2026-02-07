package requests

type UpdateDriverStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=OFFLINE ONLINE_AVAILABLE"`
}

type UpdateDriverLocationRequest struct {
	Lat       float64 `json:"lat" binding:"required"`
	Lng       float64 `json:"lng" binding:"required"`
	AccuracyM float64 `json:"accuracy_m" binding:"required"`
}
