package responses

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapGRPCError(err error) (ErrorCode, interface{}) {
	st, ok := status.FromError(err)
	if !ok {
		return CodeInternal, nil
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return CodeValidationError, nil
	case codes.Unauthenticated:
		return CodeUnauthorized, nil
	case codes.PermissionDenied:
		return CodeForbidden, nil
	case codes.NotFound:
		return CodeNotFound, nil
	case codes.AlreadyExists:
		return CodeConflict, nil
	case codes.FailedPrecondition:
		return CodeConflict, map[string]string{"reason": "FAILED_PRECONDITION"}
	case codes.Unavailable:
		return CodeInternal, map[string]string{"reason": "UPSTREAM_UNAVAILABLE"}
	default:
		return CodeInternal, nil
	}
}
