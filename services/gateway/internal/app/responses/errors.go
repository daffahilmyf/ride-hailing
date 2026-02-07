package responses

import "net/http"

type ErrorCode string

const (
	CodeValidationError ErrorCode = "VALIDATION_ERROR"
	CodeConflict       ErrorCode = "CONFLICT"
	CodeOfferExpired   ErrorCode = "OFFER_EXPIRED"
	CodeRideNotActive  ErrorCode = "RIDE_NOT_ACTIVE"
	CodeNoDriver       ErrorCode = "NO_DRIVER"
	CodeRateLimited    ErrorCode = "RATE_LIMITED"
	CodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	CodeForbidden      ErrorCode = "FORBIDDEN"
	CodeNotFound       ErrorCode = "NOT_FOUND"
	CodeInternal       ErrorCode = "INTERNAL_ERROR"
	CodeNotImplemented ErrorCode = "NOT_IMPLEMENTED"
)

type ErrorDef struct {
	Type       string
	Code       string
	Message    string
	HTTPStatus int
}

func ErrorByCode(code ErrorCode) ErrorDef {
	switch code {
	case CodeValidationError:
		return ErrorDef{Type: "BAD_REQUEST", Code: string(code), Message: "invalid request", HTTPStatus: http.StatusBadRequest}
	case CodeConflict:
		return ErrorDef{Type: "CONFLICT", Code: string(code), Message: "state conflict", HTTPStatus: http.StatusConflict}
	case CodeOfferExpired:
		return ErrorDef{Type: "CONFLICT", Code: string(code), Message: "offer expired", HTTPStatus: http.StatusConflict}
	case CodeRideNotActive:
		return ErrorDef{Type: "CONFLICT", Code: string(code), Message: "ride not active", HTTPStatus: http.StatusConflict}
	case CodeNoDriver:
		return ErrorDef{Type: "NOT_FOUND", Code: string(code), Message: "no driver available", HTTPStatus: http.StatusNotFound}
	case CodeRateLimited:
		return ErrorDef{Type: "RATE_LIMITED", Code: string(code), Message: "rate limited", HTTPStatus: http.StatusTooManyRequests}
	case CodeUnauthorized:
		return ErrorDef{Type: "UNAUTHORIZED", Code: string(code), Message: "unauthorized", HTTPStatus: http.StatusUnauthorized}
	case CodeForbidden:
		return ErrorDef{Type: "FORBIDDEN", Code: string(code), Message: "forbidden", HTTPStatus: http.StatusForbidden}
	case CodeNotFound:
		return ErrorDef{Type: "NOT_FOUND", Code: string(code), Message: "not found", HTTPStatus: http.StatusNotFound}
	case CodeNotImplemented:
		return ErrorDef{Type: "NOT_IMPLEMENTED", Code: string(code), Message: "endpoint not implemented", HTTPStatus: http.StatusNotImplemented}
	default:
		return ErrorDef{Type: "INTERNAL", Code: string(CodeInternal), Message: "internal error", HTTPStatus: http.StatusInternalServerError}
	}
}
