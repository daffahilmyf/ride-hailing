package requests

type CreateRideRequest struct {
	PickupLat  float64 `json:"pickup_lat" binding:"required"`
	PickupLng  float64 `json:"pickup_lng" binding:"required"`
	DropoffLat float64 `json:"dropoff_lat" binding:"required"`
	DropoffLng float64 `json:"dropoff_lng" binding:"required"`
}

type CancelRideRequest struct {
	Reason string `json:"reason" binding:"required"`
}
