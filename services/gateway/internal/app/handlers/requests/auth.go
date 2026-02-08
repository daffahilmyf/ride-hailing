package requests

type RegisterRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	DeviceID string `json:"device_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
}

type VerifyRequest struct {
	Channel string `json:"channel"`
	Target  string `json:"target"`
	Code    string `json:"code"`
}

type LogoutDeviceRequest struct {
	DeviceID string `json:"device_id"`
}
