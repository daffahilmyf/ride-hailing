package responses

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapGRPCError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode ErrorCode
	}{
		{"invalid_argument", status.Error(codes.InvalidArgument, "bad"), CodeValidationError},
		{"unauthenticated", status.Error(codes.Unauthenticated, "no"), CodeUnauthorized},
		{"permission", status.Error(codes.PermissionDenied, "no"), CodeForbidden},
		{"not_found", status.Error(codes.NotFound, "missing"), CodeNotFound},
		{"already_exists", status.Error(codes.AlreadyExists, "dup"), CodeConflict},
		{"failed_precondition", status.Error(codes.FailedPrecondition, "pre"), CodeConflict},
		{"unavailable", status.Error(codes.Unavailable, "down"), CodeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := MapGRPCError(tt.err)
			if code != tt.wantCode {
				t.Fatalf("expected %s, got %s", tt.wantCode, code)
			}
		})
	}
}
